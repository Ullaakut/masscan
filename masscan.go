// Package masscan provides idiomatic `masscan` bindings for go developers.
package masscan

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/mattn/go-isatty"
)

// ScanRunner represents something that can run a scan.
type ScanRunner interface {
	Run(ctx context.Context) (*Run, error)
}

// OutputFormat is the output format produced by masscan.
type OutputFormat string

// OutputFormat values supported by this package.
const (
	OutputFormatJSON     OutputFormat = "json"
	OutputFormatXML      OutputFormat = "xml"
	OutputFormatList     OutputFormat = "list"
	OutputFormatGrepable OutputFormat = "grepable"
	OutputFormatBinary   OutputFormat = "binary"
	OutputFormatUnknown  OutputFormat = "unknown"
)

func (o OutputFormat) outputFlag() (string, bool) {
	switch o {
	case OutputFormatJSON:
		return "-oJ", true
	case OutputFormatXML:
		return "-oX", true
	case OutputFormatList:
		return "-oL", true
	case OutputFormatGrepable:
		return "-oG", true
	case OutputFormatBinary:
		return "-oB", true
	default:
		return "", false
	}
}

// Port represents a discovered port.
type Port struct {
	Number    int
	Protocol  string
	Status    string
	Reason    string
	ReasonTTL int
}

// Host represents a discovered host.
type Host struct {
	Address   string
	Timestamp string
	Ports     []Port
}

// Run represents the parsed output of a masscan execution.
type Run struct {
	Hosts    []Host
	warnings []string
}

// Warnings returns parsing or runtime warnings emitted by masscan.
func (r *Run) Warnings() []string {
	if r == nil {
		return nil
	}

	output := make([]string, len(r.warnings))
	copy(output, r.warnings)
	return output
}

// Scanner represents a Masscan scanner.
type Scanner struct {
	modifySysProcAttr func(*syscall.SysProcAttr)

	args       []string
	binaryPath string

	portFilter func(Port) bool
	hostFilter func(Host) bool

	interactive bool
	toFile      *string
	output      OutputFormat
}

// Option configures a Scanner by adding or changing masscan arguments.
type Option func(*Scanner) error

// NewScanner creates a new Scanner, and can take options to apply to the scanner.
func NewScanner(options ...Option) (*Scanner, error) {
	scanner := Scanner{
		interactive: isatty.IsTerminal(os.Stdin.Fd()),
		output:      OutputFormatJSON,
	}

	for _, option := range options {
		err := option(&scanner)
		if err != nil {
			return nil, fmt.Errorf("applying option: %w", err)
		}
	}

	if scanner.binaryPath == "" {
		var err error
		scanner.binaryPath, err = exec.LookPath("masscan")
		if err != nil {
			return nil, ErrMasscanNotInstalled
		}
	}

	return &scanner, nil
}

// ToFile enables the Scanner to write the masscan XML output to a given path.
// Masscan writes the normal CLI output to stdout.
// The XML is parsed from file after the scan is finished.
func (s *Scanner) ToFile(file string) (*Scanner, error) {
	s.toFile = &file
	return s, nil
}

// WithOutputFormat sets the parser output preference when an explicit output flag is not already set.
func (s *Scanner) WithOutputFormat(format OutputFormat) (*Scanner, error) {
	if _, ok := format.outputFlag(); !ok {
		return s, fmt.Errorf("%w: %s", ErrUnsupportedOutputFormat, format)
	}
	s.output = format
	return s, nil
}

// Run executes masscan with the enabled options and parses the resulting output.
func (s *Scanner) Run(ctx context.Context) (*Run, error) {
	cmd := s.newCmd(ctx)

	return s.runAndParse(ctx, cmd)
}

// AddOptions sets more scan options after the scan is created.
func (s *Scanner) AddOptions(options ...Option) (*Scanner, error) {
	for _, option := range options {
		err := option(s)
		if err != nil {
			return s, fmt.Errorf("applying option: %w", err)
		}
	}
	return s, nil
}

// Args return the list of masscan args.
func (s *Scanner) Args() []string {
	return s.args
}

func chooseHosts(result *Run, filter func(Host) bool) {
	var filteredHosts []Host

	for _, host := range result.Hosts {
		if filter(host) {
			filteredHosts = append(filteredHosts, host)
		}
	}

	result.Hosts = filteredHosts
}

func choosePorts(result *Run, filter func(Port) bool) {
	for idx := range result.Hosts {
		var filteredPorts []Port

		for _, port := range result.Hosts[idx].Ports {
			if filter(port) {
				filteredPorts = append(filteredPorts, port)
			}
		}

		result.Hosts[idx].Ports = filteredPorts
	}
}

// WithCustomArguments sets custom arguments to give to the masscan binary.
// There should be no reason to use this, unless you are using a custom build
// of masscan or that this repository isn't up to date with the latest options
// of the official masscan release.
//
// Deprecated: You can use this as a quick way to paste a masscan command into your go code,
// but remember that the whole purpose of this repository is to be idiomatic,
// provide type checking, enums for the values that can be passed, etc.
func WithCustomArguments(args ...string) Option {
	return func(s *Scanner) error {
		s.args = append(s.args, args...)
		return nil
	}
}

// WithBinaryPath sets the masscan binary path for a scanner.
func WithBinaryPath(binaryPath string) Option {
	return func(s *Scanner) error {
		s.binaryPath = binaryPath
		return nil
	}
}

// WithFilterPort allows to set a custom function to filter out ports that
// don't fulfill a given condition. When the given function returns true,
// the port is kept, otherwise it is removed from the result. Can be used
// along with WithFilterHost.
func WithFilterPort(portFilter func(Port) bool) Option {
	return func(s *Scanner) error {
		s.portFilter = portFilter
		return nil
	}
}

// WithFilterHost allows to set a custom function to filter out hosts that
// don't fulfill a given condition. When the given function returns true,
// the host is kept, otherwise it is removed from the result. Can be used
// along with WithFilterPort.
func WithFilterHost(hostFilter func(Host) bool) Option {
	return func(s *Scanner) error {
		s.hostFilter = hostFilter
		return nil
	}
}
