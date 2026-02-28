package masscan

import (
	"errors"
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

func TestOutputFlag(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		format OutputFormat
		flag   string
		ok     bool
	}{
		{name: "json", format: OutputFormatJSON, flag: "-oJ", ok: true},
		{name: "xml", format: OutputFormatXML, flag: "-oX", ok: true},
		{name: "list", format: OutputFormatList, flag: "-oL", ok: true},
		{name: "grepable", format: OutputFormatGrepable, flag: "-oG", ok: true},
		{name: "binary", format: OutputFormatBinary, flag: "-oB", ok: true},
		{name: "unknown", format: OutputFormatUnknown, flag: "", ok: false},
		{name: "invalid", format: OutputFormat("invalid"), flag: "", ok: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			flag, ok := test.format.outputFlag()
			assert.Equal(t, test.ok, ok)
			assert.Equal(t, test.flag, flag)
		})
	}
}

func TestScannerWithOutputFormatMethod(t *testing.T) {
	t.Parallel()

	t.Run("sets valid output format", func(t *testing.T) {
		t.Parallel()

		s := &Scanner{output: OutputFormatJSON}
		updated, err := s.WithOutputFormat(OutputFormatXML)
		require.NoError(t, err)
		require.Same(t, s, updated)
		assert.Equal(t, OutputFormatXML, s.output)
	})

	t.Run("returns error for invalid output format", func(t *testing.T) {
		t.Parallel()

		s := &Scanner{output: OutputFormatJSON}
		updated, err := s.WithOutputFormat(OutputFormat("invalid"))
		require.Error(t, err)
		require.Same(t, s, updated)
		assert.ErrorIs(t, err, ErrUnsupportedOutputFormat)
		assert.Equal(t, OutputFormatJSON, s.output)
	})
}

func TestScannerAddOptions(t *testing.T) {
	t.Parallel()

	t.Run("applies options in order", func(t *testing.T) {
		t.Parallel()

		s := &Scanner{}
		updated, err := s.AddOptions(
			WithTargets("192.0.2.1"),
			WithPorts("443"),
		)
		require.NoError(t, err)
		require.Same(t, s, updated)
		assert.Equal(t, []string{"192.0.2.1", "-p", "443"}, s.args)
	})

	t.Run("returns wrapped option error", func(t *testing.T) {
		t.Parallel()

		s := &Scanner{}
		optionErr := errors.New("option failed")
		badOption := func(*Scanner) error { return optionErr }

		updated, err := s.AddOptions(badOption)
		require.Error(t, err)
		require.Same(t, s, updated)
		assert.ErrorIs(t, err, optionErr)
	})
}

func TestNewScanner(t *testing.T) {
	t.Run("returns option application error", func(t *testing.T) {
		t.Parallel()

		optionErr := errors.New("option failed")
		badOption := func(*Scanner) error { return optionErr }
		_, err := NewScanner(badOption)
		require.Error(t, err)
		assert.ErrorIs(t, err, optionErr)
	})

	t.Run("returns not installed error when binary lookup fails", func(t *testing.T) {
		t.Setenv("PATH", "")
		scanner, err := NewScanner()
		require.Error(t, err)
		assert.Nil(t, scanner)
		assert.ErrorIs(t, err, ErrMasscanNotInstalled)
	})
}
