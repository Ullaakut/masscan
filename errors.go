package masscan

import (
	"errors"
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
