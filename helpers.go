package masscan

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

type outputConfig struct {
	format OutputFormat
	path   string
	setBy  string
}

func (s *Scanner) buildArgs() []string {
	args := append([]string{}, s.args...)
	cfg := detectOutputConfig(args)
	if cfg != nil {
		return args
	}

	flag, _ := s.output.outputFlag()
	path := "-"
	if s.toFile != nil {
		path = *s.toFile
	}
	args = append(args, flag, path)
	return args
}

func detectOutputConfig(args []string) *outputConfig {
	var config *outputConfig
	for idx := 0; idx < len(args); idx++ {
		arg := args[idx]
		if format, path, ok := outputArgument(arg, args, idx); ok {
			if path == "" {
				if idx+1 < len(args) {
					path = args[idx+1]
					idx++
				} else {
					path = "-"
				}
			}
			config = &outputConfig{format: format, path: path, setBy: arg}
		}
	}

	return config
}

func outputArgument(current string, args []string, index int) (OutputFormat, string, bool) {
	mapping := []struct {
		prefix string
		format OutputFormat
	}{
		{prefix: "-oX", format: OutputFormatXML},
		{prefix: "-oJ", format: OutputFormatJSON},
		{prefix: "-oL", format: OutputFormatList},
		{prefix: "-oG", format: OutputFormatGrepable},
		{prefix: "-oB", format: OutputFormatBinary},
	}

	for _, entry := range mapping {
		if current == entry.prefix {
			if index+1 < len(args) {
				return entry.format, args[index+1], true
			}
			return entry.format, "", true
		}

		if strings.HasPrefix(current, entry.prefix) && len(current) > len(entry.prefix) {
			return entry.format, current[len(entry.prefix):], true
		}
	}

	return OutputFormatUnknown, "", false
}

func (s *Scanner) outputConfig() outputConfig {
	args := s.buildArgs()
	if cfg := detectOutputConfig(args); cfg != nil {
		return *cfg
	}

	path := "-"
	if s.toFile != nil {
		path = *s.toFile
	}

	return outputConfig{format: s.output, path: path, setBy: "default"}
}

func (s *Scanner) newCmd(ctx context.Context) *exec.Cmd {
	args := s.buildArgs()

	//nolint:gosec // Arguments are passed directly to masscan; users intentionally control args.
	cmd := exec.CommandContext(ctx, s.binaryPath, args...)
	if s.modifySysProcAttr != nil {
		s.modifySysProcAttr(cmd.SysProcAttr)
	}
	return cmd
}

func (s *Scanner) runAndParse(ctx context.Context, cmd *exec.Cmd) (*Run, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	runErr := cmd.Run()
	result, parseErr := s.processMasscanResult(&stdout, &stderr)

	return finalizeRun(ctx, runErr, parseErr, result, &stdout, &stderr)
}

func finalizeRun(ctx context.Context, runErr, parseErr error, result *Run, stdout, stderr *bytes.Buffer) (*Run, error) {
	if runErr == nil {
		return result, parseErr
	}

	mappedErr := mapRunError(ctx, runErr)
	if mappedErr != nil && !errors.Is(mappedErr, runErr) {
		return result, mappedErr
	}

	if parseErr != nil {
		if stdout.Len() == 0 && stderr.Len() == 0 {
			return result, nil
		}
		return result, parseErr
	}
	if mappedErr != nil {
		return result, mappedErr
	}

	return result, runErr
}

func (s *Scanner) processMasscanResult(stdout, stderr *bytes.Buffer) (*Run, error) {
	result := &Run{}

	var warnings []string
	warnings, errStdout := checkStdErr(stderr)
	if errStdout != nil {
		return result, errStdout
	}

	config := s.outputConfig()
	contents := stdout.Bytes()
	if config.path != "" && config.path != "-" {
		var err error
		contents, err = os.ReadFile(config.path)
		if err != nil {
			return result, fmt.Errorf("reading output file %s: %w", config.path, err)
		}

		chmodErr := os.Chmod(config.path, 0o600)
		if chmodErr != nil {
			warnings = append(warnings, fmt.Sprintf("setting output file permissions: %s", chmodErr))
		}
	}

	parsed, err := parseOutput(contents, config.format)
	if err != nil {
		return result, fmt.Errorf("%w: %v", ErrParseOutput, err)
	}
	result = parsed

	result.warnings = append(result.warnings, warnings...)

	if s.portFilter != nil {
		choosePorts(result, s.portFilter)
	}
	if s.hostFilter != nil {
		chooseHosts(result, s.hostFilter)
	}

	return result, nil
}

func parseOutput(contents []byte, format OutputFormat) (*Run, error) {
	trimmed := strings.TrimSpace(string(contents))
	if trimmed == "" {
		return &Run{}, nil
	}

	if format == OutputFormatUnknown {
		if strings.HasPrefix(trimmed, "<") {
			format = OutputFormatXML
		} else if strings.HasPrefix(trimmed, "[") || strings.HasPrefix(trimmed, "{") {
			format = OutputFormatJSON
		} else if strings.Contains(trimmed, "Host:") && strings.Contains(trimmed, "Ports:") {
			format = OutputFormatGrepable
		} else {
			format = OutputFormatList
		}
	}

	switch format {
	case OutputFormatJSON:
		return parseJSON(contents)
	case OutputFormatXML:
		return parseXML(contents)
	case OutputFormatList:
		return parseList(contents)
	case OutputFormatGrepable:
		return parseGrepable(contents)
	case OutputFormatBinary:
		return nil, fmt.Errorf("%w: binary output (-oB) cannot be parsed directly", ErrUnsupportedOutputFormat)
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedOutputFormat, format)
	}
}

type masscanJSONPort struct {
	Port   json.RawMessage `json:"port"`
	Proto  string          `json:"proto"`
	Status string          `json:"status"`
	Reason string          `json:"reason"`
	TTL    int             `json:"ttl"`
}

type masscanJSONHost struct {
	IP        string            `json:"ip"`
	Timestamp string            `json:"timestamp"`
	Ports     []masscanJSONPort `json:"ports"`
}

func parseJSON(contents []byte) (*Run, error) {
	result := &Run{}
	clean := bytes.TrimSpace(contents)

	var hosts []masscanJSONHost
	if len(clean) > 0 && clean[0] == '[' {
		if err := json.Unmarshal(clean, &hosts); err != nil {
			lineHosts, lineErr := parseJSONLines(contents)
			if lineErr != nil {
				return nil, fmt.Errorf("%w: %v", ErrInvalidOutput, err)
			}
			hosts = lineHosts
		}
	} else {
		lineHosts, err := parseJSONLines(contents)
		if err != nil {
			return nil, err
		}
		hosts = lineHosts
	}

	for _, host := range hosts {
		mapped := Host{Address: host.IP, Timestamp: host.Timestamp}
		for _, port := range host.Ports {
			portNumber, err := parsePortNumber(port.Port)
			if err != nil {
				return nil, err
			}

			mapped.Ports = append(mapped.Ports, Port{
				Number:    portNumber,
				Protocol:  port.Proto,
				Status:    port.Status,
				Reason:    port.Reason,
				ReasonTTL: port.TTL,
			})
		}

		result.Hosts = append(result.Hosts, mapped)
	}

	return result, nil
}

func parseJSONLines(contents []byte) ([]masscanJSONHost, error) {
	var hosts []masscanJSONHost
	for line := range strings.SplitSeq(string(contents), "\n") {
		line = strings.TrimSpace(line)
		line = strings.TrimSuffix(line, ",")
		if line == "" || line == "[" || line == "]" {
			continue
		}

		var host masscanJSONHost
		if err := json.Unmarshal([]byte(line), &host); err != nil {
			continue
		}

		hosts = append(hosts, host)
	}

	if len(hosts) == 0 {
		return nil, fmt.Errorf("%w: no parsable JSON records found", ErrInvalidOutput)
	}

	return hosts, nil
}

func parsePortNumber(raw json.RawMessage) (int, error) {
	var asInt int
	if err := json.Unmarshal(raw, &asInt); err == nil {
		return asInt, nil
	}

	var asString string
	if err := json.Unmarshal(raw, &asString); err != nil {
		return 0, fmt.Errorf("%w: invalid port value %q", ErrInvalidOutput, string(raw))
	}

	parsed, err := strconv.Atoi(asString)
	if err != nil {
		return 0, fmt.Errorf("%w: invalid port %q", ErrInvalidOutput, asString)
	}

	return parsed, nil
}

type xmlRun struct {
	Hosts []xmlHost `xml:"host"`
}

type xmlHost struct {
	Addresses []xmlAddress `xml:"address"`
	Ports     []xmlPort    `xml:"ports>port"`
	StartTime string       `xml:"starttime,attr"`
	EndTime   string       `xml:"endtime,attr"`
}

type xmlAddress struct {
	Addr string `xml:"addr,attr"`
}

type xmlPort struct {
	Protocol string   `xml:"protocol,attr"`
	PortID   int      `xml:"portid,attr"`
	State    xmlState `xml:"state"`
}

type xmlState struct {
	Status string `xml:"state,attr"`
	Reason string `xml:"reason,attr"`
	TTL    int    `xml:"reason_ttl,attr"`
}

func parseXML(contents []byte) (*Run, error) {
	var parsed xmlRun
	if err := xml.Unmarshal(contents, &parsed); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidOutput, err)
	}

	result := &Run{}
	for _, host := range parsed.Hosts {
		timestamp := host.StartTime
		if timestamp == "" {
			timestamp = host.EndTime
		}

		mapped := Host{Timestamp: timestamp}
		if len(host.Addresses) > 0 {
			mapped.Address = host.Addresses[0].Addr
		}

		for _, port := range host.Ports {
			mapped.Ports = append(mapped.Ports, Port{
				Number:    port.PortID,
				Protocol:  port.Protocol,
				Status:    port.State.Status,
				Reason:    port.State.Reason,
				ReasonTTL: port.State.TTL,
			})
		}

		result.Hosts = append(result.Hosts, mapped)
	}

	return result, nil
}

func parseList(contents []byte) (*Run, error) {
	result := &Run{}
	lines := strings.Split(string(contents), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}

		status := fields[0]
		proto := fields[1]
		portValue := fields[2]
		ip := fields[3]
		timestamp := ""
		if len(fields) > 4 {
			timestamp = fields[4]
		}

		port, err := strconv.Atoi(portValue)
		if err != nil {
			continue
		}

		upsertPort(result, ip, timestamp, Port{Number: port, Protocol: proto, Status: status})
	}

	if len(result.Hosts) == 0 {
		stdoutResult, err := parseStdoutLines(contents)
		if err != nil {
			return nil, err
		}
		return stdoutResult, nil
	}

	return result, nil
}

func parseGrepable(contents []byte) (*Run, error) {
	result := &Run{}
	lines := strings.Split(string(contents), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if !strings.HasPrefix(line, "Host:") {
			continue
		}

		hostPart := strings.TrimPrefix(line, "Host:")
		hostPart = strings.TrimSpace(hostPart)
		hostFields := strings.Fields(hostPart)
		if len(hostFields) == 0 {
			continue
		}

		address := hostFields[0]
		portsIndex := strings.Index(line, "Ports:")
		if portsIndex == -1 {
			continue
		}

		portsSection := strings.TrimSpace(line[portsIndex+len("Ports:"):])
		for portSpec := range strings.SplitSeq(portsSection, ",") {
			parts := strings.Split(strings.TrimSpace(portSpec), "/")
			if len(parts) < 3 {
				continue
			}

			port, err := strconv.Atoi(parts[0])
			if err != nil {
				continue
			}

			upsertPort(result, address, "", Port{
				Number:   port,
				Status:   parts[1],
				Protocol: parts[2],
			})
		}
	}

	if len(result.Hosts) == 0 {
		return nil, fmt.Errorf("%w: no parsable grepable records found", ErrInvalidOutput)
	}

	return result, nil
}

var discoveredLine = regexp.MustCompile(`Discovered\s+(\S+)\s+port\s+(\d+)\/(\w+)\s+on\s+([\w\.:\-]+)`)

func parseStdoutLines(contents []byte) (*Run, error) {
	result := &Run{}
	for _, line := range strings.Split(string(contents), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		matches := discoveredLine.FindStringSubmatch(line)
		if len(matches) != 5 {
			continue
		}

		port, err := strconv.Atoi(matches[2])
		if err != nil {
			continue
		}

		upsertPort(result, matches[4], "", Port{Status: matches[1], Number: port, Protocol: matches[3]})
	}

	if len(result.Hosts) == 0 {
		return nil, fmt.Errorf("%w: no parsable records found", ErrInvalidOutput)
	}

	return result, nil
}

func upsertPort(result *Run, address, timestamp string, port Port) {
	for idx := range result.Hosts {
		if result.Hosts[idx].Address != address {
			continue
		}

		if result.Hosts[idx].Timestamp == "" && timestamp != "" {
			result.Hosts[idx].Timestamp = timestamp
		}

		result.Hosts[idx].Ports = append(result.Hosts[idx].Ports, port)
		return
	}

	result.Hosts = append(result.Hosts, Host{
		Address:   address,
		Timestamp: timestamp,
		Ports:     []Port{port},
	})
}

func mapRunError(ctx context.Context, err error) error {
	if err == nil {
		return nil
	}

	switch {
	case errors.Is(ctx.Err(), context.DeadlineExceeded):
		return ErrScanTimeout
	case errors.Is(ctx.Err(), context.Canceled):
		return ErrScanInterrupt
	case isInterruptExit(err):
		return ErrScanInterrupt
	default:
		return err
	}
}

func isInterruptExit(err error) bool {
	if err == nil {
		return false
	}

	switch err.Error() {
	case "exit status 0xc000013a": // Exit code for ctrl+c on Windows
		return true
	case "exit status 130": // Exit code for ctrl+c on Linux
		return true
	default:
		return false
	}
}

// checkStdErr writes the output of stderr to the warnings array.
// It also processes masscan stderr output containing none-critical errors and warnings.
func checkStdErr(stderr *bytes.Buffer) (warnings []string, err error) {
	if stderr.Len() <= 0 {
		return nil, nil
	}

	stderrSplit := strings.SplitSeq(strings.Trim(stderr.String(), "\n "), "\n")

	for warning := range stderrSplit {
		warnings = append(warnings, strings.Trim(warning, " "))
		switch {
		case strings.Contains(warning, "Malloc Failed!"):
			return warnings, ErrMallocFailed
		case strings.Contains(strings.ToLower(warning), "permission denied"):
			return warnings, ErrRequiresRoot
		case strings.Contains(strings.ToLower(warning), "you must be root"):
			return warnings, ErrRequiresRoot
		case strings.Contains(warning, "requires root privileges."):
			return warnings, ErrRequiresRoot
		case strings.Contains(strings.ToLower(warning), "could not resolve"):
			return warnings, ErrResolveName
		case strings.Contains(strings.ToLower(warning), "error resolving"):
			return warnings, ErrResolveName
		default:
		}
	}
	return warnings, nil
}
