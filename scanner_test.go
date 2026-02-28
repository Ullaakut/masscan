package masscan

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScannerRunWithFilters(t *testing.T) {
	t.Setenv("MASSCAN_STDERR", "")
	t.Setenv("MASSCAN_OUTPUT_TO_FILE", "0")
	t.Setenv("MASSCAN_EXIT_CODE", "")

	scanner := newFakeScanner(t, "output.json",
		WithFilterHost(func(host Host) bool {
			return host.Address == "192.0.2.10"
		}),
		WithFilterPort(func(port Port) bool {
			return port.Number == 443
		}),
	)

	result, err := scanner.Run(t.Context())
	require.NoError(t, err)
	require.Len(t, result.Hosts, 1)
	assert.Equal(t, "192.0.2.10", result.Hosts[0].Address)
	require.Len(t, result.Hosts[0].Ports, 1)
	assert.Equal(t, 443, result.Hosts[0].Ports[0].Number)
}

func TestScannerRunReadsOutputFile(t *testing.T) {
	t.Setenv("MASSCAN_STDERR", "")
	t.Setenv("MASSCAN_OUTPUT_TO_FILE", "1")
	t.Setenv("MASSCAN_EXIT_CODE", "")

	scanner := newFakeScanner(t, "output_xml.xml", WithOutputFormat(OutputFormatXML))

	outPath := filepath.Join(t.TempDir(), "scan.xml")
	_, err := scanner.ToFile(outPath)
	require.NoError(t, err)

	result, err := scanner.Run(t.Context())
	require.NoError(t, err)
	require.Len(t, result.Hosts, 2)
	assert.Equal(t, "192.168.1.254", result.Hosts[0].Address)
	assert.Equal(t, "1772276417", result.Hosts[0].Timestamp)
}

func TestScannerRunReturnsFatalStderrErrors(t *testing.T) {
	t.Setenv("MASSCAN_STDERR", "You must be root")
	t.Setenv("MASSCAN_OUTPUT_TO_FILE", "0")
	t.Setenv("MASSCAN_EXIT_CODE", "1")

	scanner := newFakeScanner(t, "output.json")

	_, err := scanner.Run(t.Context())
	assert.ErrorIs(t, err, ErrRequiresRoot)
}

func TestScannerRunCollectsWarnings(t *testing.T) {
	t.Setenv("MASSCAN_STDERR", "non-fatal warning")
	t.Setenv("MASSCAN_OUTPUT_TO_FILE", "0")
	t.Setenv("MASSCAN_EXIT_CODE", "")

	scanner := newFakeScanner(t, "output.json")

	result, err := scanner.Run(t.Context())
	require.NoError(t, err)

	warnings := result.Warnings()
	require.Len(t, warnings, 1)
	assert.Equal(t, "non-fatal warning", warnings[0])
}

func newFakeScanner(t *testing.T, fixture string, extraOptions ...Option) *Scanner {
	t.Helper()

	scriptPath := filepath.Join("tests", "scripts", "fake_masscan.sh")
	fixturePath := filepath.Join("tests", "fixtures", fixture)

	options := []Option{
		WithBinaryPath("sh"),
		WithCustomArguments(scriptPath, fixturePath),
		WithOutputFormat(OutputFormatJSON),
	}
	options = append(options, extraOptions...)

	scanner, err := NewScanner(options...)
	require.NoError(t, err)

	return scanner
}
