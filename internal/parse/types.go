package parse

import "errors"

// Format represents a masscan output format in the parser layer.
type Format string

const (
	// FormatJSON is a JSON output format.
	FormatJSON Format = "json"
	// FormatXML is an XML output format.
	FormatXML Format = "xml"
	// FormatList is a list output format.
	FormatList Format = "list"
	// FormatGrepable is a grepable output format.
	FormatGrepable Format = "grepable"
	// FormatBinary is a binary output format.
	FormatBinary Format = "binary"
	// FormatUnknown means output format autodetection is required.
	FormatUnknown Format = "unknown"
)

var (
	// ErrUnsupportedFormat means this parser cannot decode the requested format.
	ErrUnsupportedFormat = errors.New("unsupported output format")
	// ErrInvalidOutput means contents could not be decoded for the requested format.
	ErrInvalidOutput = errors.New("invalid masscan output")
)

// Result is the parsed representation of a masscan output.
type Result struct {
	Hosts []Host
}

// Host is a discovered host in parsed output.
type Host struct {
	Address   string
	Timestamp string
	Ports     []Port
}

// Port is a discovered port in parsed output.
type Port struct {
	Number    int
	Protocol  string
	Status    string
	Reason    string
	ReasonTTL int
}
