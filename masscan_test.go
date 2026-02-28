package masscan

import (
	"errors"
	"path/filepath"
	"testing"
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
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	if len(result.Hosts) != 1 {
		t.Fatalf("hosts mismatch: got %d want 1", len(result.Hosts))
	}

	if result.Hosts[0].Address != "192.0.2.10" {
		t.Fatalf("host mismatch: got %q", result.Hosts[0].Address)
	}

	if len(result.Hosts[0].Ports) != 1 || result.Hosts[0].Ports[0].Number != 443 {
		t.Fatalf("filtered ports mismatch: %#v", result.Hosts[0].Ports)
	}
}

func TestScannerRunReadsOutputFile(t *testing.T) {
	t.Setenv("MASSCAN_STDERR", "")
	t.Setenv("MASSCAN_OUTPUT_TO_FILE", "1")
	t.Setenv("MASSCAN_EXIT_CODE", "")

	scanner := newFakeScanner(t, "output_xml.xml", WithOutputFormat(OutputFormatXML))

	outPath := filepath.Join(t.TempDir(), "scan.xml")
	_, err := scanner.ToFile(outPath)
	if err != nil {
		t.Fatalf("ToFile returned error: %v", err)
	}

	result, err := scanner.Run(t.Context())
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	if len(result.Hosts) != 2 {
		t.Fatalf("hosts mismatch: got %d want 2", len(result.Hosts))
	}

	if result.Hosts[0].Address != "192.168.1.254" {
		t.Fatalf("host mismatch: got %q", result.Hosts[0].Address)
	}

	if result.Hosts[0].Timestamp != "1772276417" {
		t.Fatalf("timestamp mismatch: got %q", result.Hosts[0].Timestamp)
	}
}

func TestScannerRunReturnsFatalStderrErrors(t *testing.T) {
	t.Setenv("MASSCAN_STDERR", "You must be root")
	t.Setenv("MASSCAN_OUTPUT_TO_FILE", "0")
	t.Setenv("MASSCAN_EXIT_CODE", "1")

	scanner := newFakeScanner(t, "output.json")

	_, err := scanner.Run(t.Context())
	if !errors.Is(err, ErrRequiresRoot) {
		t.Fatalf("expected ErrRequiresRoot, got: %v", err)
	}
}

func TestScannerRunCollectsWarnings(t *testing.T) {
	t.Setenv("MASSCAN_STDERR", "non-fatal warning")
	t.Setenv("MASSCAN_OUTPUT_TO_FILE", "0")
	t.Setenv("MASSCAN_EXIT_CODE", "")

	scanner := newFakeScanner(t, "output.json")

	result, err := scanner.Run(t.Context())
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	warnings := result.Warnings()
	if len(warnings) != 1 || warnings[0] != "non-fatal warning" {
		t.Fatalf("warnings mismatch: %#v", warnings)
	}
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
	if err != nil {
		t.Fatalf("NewScanner returned error: %v", err)
	}

	return scanner
}
