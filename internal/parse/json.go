package parse

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

type jsonPort struct {
	Port   json.RawMessage `json:"port"`
	Proto  string          `json:"proto"`
	Status string          `json:"status"`
	Reason string          `json:"reason"`
	TTL    int             `json:"ttl"`
}

type jsonHost struct {
	IP        string     `json:"ip"`
	Timestamp string     `json:"timestamp"`
	Ports     []jsonPort `json:"ports"`
}

func parseJSON(contents []byte) (*Result, error) {
	result := &Result{}
	hosts, err := parseJSONHosts(contents)
	if err != nil {
		return nil, err
	}

	for _, host := range hosts {
		mapped := Host{Address: host.IP, Timestamp: host.Timestamp}
		for _, port := range host.Ports {
			portNumber, portErr := parsePortNumber(port.Port)
			if portErr != nil {
				return nil, portErr
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

func parseJSONHosts(contents []byte) ([]jsonHost, error) {
	clean := bytes.TrimSpace(contents)
	if len(clean) == 0 || clean[0] != '[' {
		return parseJSONLines(contents)
	}

	var hosts []jsonHost
	err := json.Unmarshal(clean, &hosts)
	if err == nil {
		return hosts, nil
	}

	lineHosts, lineErr := parseJSONLines(contents)
	if lineErr != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidOutput, err)
	}

	return lineHosts, nil
}

func parseJSONLines(contents []byte) ([]jsonHost, error) {
	lineCount := bytes.Count(contents, []byte{'\n'}) + 1
	hosts := make([]jsonHost, 0, lineCount)
	for line := range strings.SplitSeq(string(contents), "\n") {
		line = strings.TrimSpace(line)
		line = strings.TrimSuffix(line, ",")
		if line == "" || line == "[" || line == "]" {
			continue
		}

		var host jsonHost
		err := json.Unmarshal([]byte(line), &host)
		if err != nil {
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
	err := json.Unmarshal(raw, &asInt)
	if err == nil {
		return asInt, nil
	}

	var asString string
	err = json.Unmarshal(raw, &asString)
	if err != nil {
		return 0, fmt.Errorf("%w: invalid port value %q", ErrInvalidOutput, string(raw))
	}

	parsed, convErr := strconv.Atoi(asString)
	if convErr != nil {
		return 0, fmt.Errorf("%w: invalid port %q", ErrInvalidOutput, asString)
	}

	return parsed, nil
}
