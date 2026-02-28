package masscan

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFinalizeRun(t *testing.T) {
	t.Parallel()

	parseErr := errors.New("parse failed")
	runErr := errors.New("exit status 1")
	deadlineCtx, deadlineCancel := context.WithDeadline(t.Context(), time.Now().Add(-time.Second))
	defer deadlineCancel()

	tests := []struct {
		name        string
		ctx         context.Context
		runErr      error
		parseErr    error
		stdout      string
		stderr      string
		expectedErr error
	}{
		{
			name:        "run succeeds with no parse error",
			ctx:         t.Context(),
			runErr:      nil,
			parseErr:    nil,
			expectedErr: nil,
		},
		{
			name:        "run succeeds with parse error",
			ctx:         t.Context(),
			runErr:      nil,
			parseErr:    parseErr,
			expectedErr: parseErr,
		},
		{
			name:        "mapped interrupt returned before parse handling",
			ctx:         t.Context(),
			runErr:      errors.New("exit status 130"),
			parseErr:    parseErr,
			expectedErr: ErrScanInterrupt,
		},
		{
			name:        "mapped timeout returned before parse handling",
			ctx:         deadlineCtx,
			runErr:      runErr,
			parseErr:    parseErr,
			expectedErr: ErrScanTimeout,
		},
		{
			name:        "run fails parse fails no output returns nil",
			ctx:         t.Context(),
			runErr:      runErr,
			parseErr:    parseErr,
			expectedErr: nil,
		},
		{
			name:        "run fails parse fails stdout present returns parse error",
			ctx:         t.Context(),
			runErr:      runErr,
			parseErr:    parseErr,
			stdout:      "output",
			expectedErr: parseErr,
		},
		{
			name:        "run fails parse fails stderr present returns parse error",
			ctx:         t.Context(),
			runErr:      runErr,
			parseErr:    parseErr,
			stderr:      "warn",
			expectedErr: parseErr,
		},
		{
			name:        "run fails no parse error returns run error",
			ctx:         t.Context(),
			runErr:      runErr,
			parseErr:    nil,
			expectedErr: runErr,
		},
		{
			name:        "mapped equals run error returns run error",
			ctx:         t.Context(),
			runErr:      ErrScanInterrupt,
			parseErr:    nil,
			expectedErr: ErrScanInterrupt,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			result := &Run{}
			stdout := bytes.NewBufferString(test.stdout)
			stderr := bytes.NewBufferString(test.stderr)
			actual, err := finalizeRun(test.ctx, test.runErr, test.parseErr, result, stdout, stderr)

			require.Same(t, result, actual)
			if test.expectedErr == nil {
				assert.NoError(t, err)
				return
			}
			assert.ErrorIs(t, err, test.expectedErr)
		})
	}
}
