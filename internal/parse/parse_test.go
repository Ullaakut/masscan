package parse

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOutput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		fixture    string
		format     Format
		hosts      int
		firstHost  string
		firstPorts int
		firstTime  string
	}{
		{
			name:       "json array",
			fixture:    "output.json",
			format:     FormatJSON,
			hosts:      2,
			firstHost:  "192.0.2.10",
			firstPorts: 2,
		},
		{
			name:       "json lines",
			fixture:    "output_json_lines.txt",
			format:     FormatJSON,
			hosts:      2,
			firstHost:  "192.0.2.70",
			firstPorts: 1,
		},
		{
			name:       "xml",
			fixture:    "output_xml.xml",
			format:     FormatXML,
			hosts:      2,
			firstHost:  "192.168.1.254",
			firstPorts: 1,
			firstTime:  "1772276417",
		},
		{
			name:       "list",
			fixture:    "output_list.txt",
			format:     FormatList,
			hosts:      2,
			firstHost:  "192.0.2.40",
			firstPorts: 1,
		},
		{
			name:       "grepable",
			fixture:    "output_grepable.txt",
			format:     FormatGrepable,
			hosts:      1,
			firstHost:  "192.0.2.50",
			firstPorts: 2,
		},
		{
			name:       "unknown autodetect json",
			fixture:    "output.json",
			format:     FormatUnknown,
			hosts:      2,
			firstHost:  "192.0.2.10",
			firstPorts: 2,
		},
		{
			name:       "unknown autodetect stdout",
			fixture:    "output_stdout.txt",
			format:     FormatUnknown,
			hosts:      1,
			firstHost:  "192.0.2.60",
			firstPorts: 2,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			contents := mustReadFixture(t, test.fixture)
			result, err := Output(contents, test.format)
			require.NoError(t, err)
			require.Len(t, result.Hosts, test.hosts)
			assert.Equal(t, test.firstHost, result.Hosts[0].Address)
			require.Len(t, result.Hosts[0].Ports, test.firstPorts)
			if test.firstTime != "" {
				assert.Equal(t, test.firstTime, result.Hosts[0].Timestamp)
			}
		})
	}
}

func TestOutputErrors(t *testing.T) {
	t.Parallel()

	_, err := Output([]byte("abc"), FormatBinary)
	assert.ErrorIs(t, err, ErrUnsupportedFormat)

	_, err = Output([]byte("not-json"), FormatJSON)
	assert.ErrorIs(t, err, ErrInvalidOutput)
}

func mustReadFixture(t *testing.T, name string) []byte {
	t.Helper()

	fixturePath := filepath.Join("..", "..", "tests", "fixtures", name)
	contents, err := os.ReadFile(fixturePath)
	require.NoError(t, err)

	return contents
}
