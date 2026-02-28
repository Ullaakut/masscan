package masscan

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestParseOutput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		fixture    string
		format     OutputFormat
		hosts      int
		firstHost  string
		firstPorts int
		firstTime  string
	}{
		{
			name:       "json array",
			fixture:    "output.json",
			format:     OutputFormatJSON,
			hosts:      2,
			firstHost:  "192.0.2.10",
			firstPorts: 2,
		},
		{
			name:       "json lines",
			fixture:    "output_json_lines.txt",
			format:     OutputFormatJSON,
			hosts:      2,
			firstHost:  "192.0.2.70",
			firstPorts: 1,
		},
		{
			name:       "xml",
			fixture:    "output_xml.xml",
			format:     OutputFormatXML,
			hosts:      2,
			firstHost:  "192.168.1.254",
			firstPorts: 1,
			firstTime:  "1772276417",
		},
		{
			name:       "list",
			fixture:    "output_list.txt",
			format:     OutputFormatList,
			hosts:      2,
			firstHost:  "192.0.2.40",
			firstPorts: 1,
		},
		{
			name:       "grepable",
			fixture:    "output_grepable.txt",
			format:     OutputFormatGrepable,
			hosts:      1,
			firstHost:  "192.0.2.50",
			firstPorts: 2,
		},
		{
			name:       "unknown autodetect json",
			fixture:    "output.json",
			format:     OutputFormatUnknown,
			hosts:      2,
			firstHost:  "192.0.2.10",
			firstPorts: 2,
		},
		{
			name:       "unknown autodetect stdout",
			fixture:    "output_stdout.txt",
			format:     OutputFormatUnknown,
			hosts:      1,
			firstHost:  "192.0.2.60",
			firstPorts: 2,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			contents := mustReadFixture(t, test.fixture)
			result, err := parseOutput(contents, test.format)
			if err != nil {
				t.Fatalf("parseOutput returned error: %v", err)
			}

			if len(result.Hosts) != test.hosts {
				t.Fatalf("hosts mismatch: got %d want %d", len(result.Hosts), test.hosts)
			}

			if result.Hosts[0].Address != test.firstHost {
				t.Fatalf("first host mismatch: got %q want %q", result.Hosts[0].Address, test.firstHost)
			}

			if len(result.Hosts[0].Ports) != test.firstPorts {
				t.Fatalf("first host port count mismatch: got %d want %d", len(result.Hosts[0].Ports), test.firstPorts)
			}

			if test.firstTime != "" && result.Hosts[0].Timestamp != test.firstTime {
				t.Fatalf("first host timestamp mismatch: got %q want %q", result.Hosts[0].Timestamp, test.firstTime)
			}
		})
	}
}

func TestParseOutputErrors(t *testing.T) {
	t.Parallel()

	_, err := parseOutput([]byte("abc"), OutputFormatBinary)
	if !errors.Is(err, ErrUnsupportedOutputFormat) {
		t.Fatalf("expected ErrUnsupportedOutputFormat, got: %v", err)
	}

	_, err = parseOutput([]byte("not-json"), OutputFormatJSON)
	if !errors.Is(err, ErrInvalidOutput) {
		t.Fatalf("expected ErrInvalidOutput, got: %v", err)
	}
}

func mustReadFixture(t *testing.T, name string) []byte {
	t.Helper()

	fixturePath := filepath.Join("tests", "fixtures", name)
	contents, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatalf("reading fixture %s: %v", fixturePath, err)
	}

	return contents
}
