package masscan

import (
	"bytes"
	"context"
	"errors"
	"strings"
)

var (
	// ErrMasscanNotInstalled means that upon trying to manually locate masscan in the user's path,
	// it was not found. Either use the WithBinaryPath method to set it manually, or make sure that
	// the masscan binary is present in the user's $PATH.
	ErrMasscanNotInstalled = errors.New("masscan binary was not found")

	// ErrScanTimeout means that the provided context timeout triggered done before the scanner finished its scan.
	// This error is *not* returned if a scan timeout was configured using Masscan arguments, since Masscan would
	// gracefully shut down it's scanning and return some results in that case.
	ErrScanTimeout = errors.New("masscan scan timed out")

	// ErrScanInterrupt means that the scan was interrupted before the scanner finished its scan.
	// Reasons for this error might be sigint or a cancelled context.
	ErrScanInterrupt = errors.New("masscan scan interrupted")

	// ErrParseOutput means that masscan's output was not parsed successfully.
	ErrParseOutput = errors.New("masscan output parsing failure, see warnings for details")

	// ErrUnsupportedOutputFormat means that the requested masscan output format
	// cannot be parsed by this package.
	ErrUnsupportedOutputFormat = errors.New("unsupported masscan output format")

	// ErrInvalidOutput means that masscan returned output that could not be decoded.
	ErrInvalidOutput = errors.New("invalid masscan output")

	// ErrMallocFailed means that masscan failed because it could not allocate memory.
	ErrMallocFailed = errors.New("masscan malloc failed")

	// ErrRequiresRoot means that a feature (e.g. OS detection) requires root privileges.
	ErrRequiresRoot = errors.New("this feature requires root privileges")

	// ErrResolveName means that masscan could not resolve a name.
	ErrResolveName = errors.New("masscan could not resolve a name")
)

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
