package masscan

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
)

func (s *Scanner) newCmd(ctx context.Context) *exec.Cmd {
	args := s.buildArgs()

	//nolint:gosec // Arguments are passed directly to masscan; users intentionally control args.
	cmd := exec.CommandContext(ctx, s.binaryPath, args...)
	if s.modifySysProcAttr != nil {
		s.modifySysProcAttr(cmd.SysProcAttr)
	}
	return cmd
}

func (s *Scanner) runAndParse(ctx context.Context, cmd *exec.Cmd) (*Run, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	runErr := cmd.Run()
	result, parseErr := s.processMasscanResult(&stdout, &stderr)

	return finalizeRun(ctx, runErr, parseErr, result, &stdout, &stderr)
}

func finalizeRun(ctx context.Context, runErr, parseErr error, result *Run, stdout, stderr *bytes.Buffer) (*Run, error) {
	if runErr == nil {
		return result, parseErr
	}

	mappedErr := mapRunError(ctx, runErr)
	if mappedErr != nil && !errors.Is(mappedErr, runErr) {
		return result, mappedErr
	}

	if parseErr != nil {
		if stdout.Len() == 0 && stderr.Len() == 0 {
			return result, nil
		}
		return result, parseErr
	}
	if mappedErr != nil {
		return result, mappedErr
	}

	return result, runErr
}

func (s *Scanner) processMasscanResult(stdout, stderr *bytes.Buffer) (*Run, error) {
	result := &Run{}

	var warnings []string
	warnings, errStdout := checkStdErr(stderr)
	if errStdout != nil {
		return result, errStdout
	}

	config := s.outputConfig()
	contents := stdout.Bytes()
	if config.path != "" && config.path != "-" {
		var err error
		contents, err = os.ReadFile(config.path)
		if err != nil {
			return result, fmt.Errorf("reading output file %s: %w", config.path, err)
		}

		chmodErr := os.Chmod(config.path, 0o600)
		if chmodErr != nil {
			warnings = append(warnings, fmt.Sprintf("setting output file permissions: %s", chmodErr))
		}
	}

	parsed, err := parseOutput(contents, config.format)
	if err != nil {
		return result, fmt.Errorf("%w: %w", ErrParseOutput, err)
	}
	result = parsed

	result.warnings = append(result.warnings, warnings...)

	if s.portFilter != nil {
		choosePorts(result, s.portFilter)
	}
	if s.hostFilter != nil {
		chooseHosts(result, s.hostFilter)
	}

	return result, nil
}
