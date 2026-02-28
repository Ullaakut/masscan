package parse

import (
	"fmt"
	"strings"
)

// Output parses masscan output contents with the given format.
func Output(contents []byte, format Format) (*Result, error) {
	trimmed := strings.TrimSpace(string(contents))
	if trimmed == "" {
		return &Result{}, nil
	}

	if format == FormatUnknown {
		switch {
		case strings.HasPrefix(trimmed, "<"):
			format = FormatXML
		case strings.HasPrefix(trimmed, "[") || strings.HasPrefix(trimmed, "{"):
			format = FormatJSON
		case strings.Contains(trimmed, "Host:") && strings.Contains(trimmed, "Ports:"):
			format = FormatGrepable
		default:
			format = FormatList
		}
	}

	switch format {
	case FormatJSON:
		return parseJSON(contents)
	case FormatXML:
		return parseXML(contents)
	case FormatList:
		return parseList(contents)
	case FormatGrepable:
		return parseGrepable(contents)
	case FormatBinary:
		return nil, fmt.Errorf("%w: binary output (-oB) cannot be parsed directly", ErrUnsupportedFormat)
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedFormat, format)
	}
}
