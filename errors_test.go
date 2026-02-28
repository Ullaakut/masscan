package masscan

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCheckStdErr(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		stderr       string
		expectedWarn []string
		expectedErr  error
	}{
		{
			name:         "empty stderr",
			stderr:       "",
			expectedWarn: nil,
			expectedErr:  nil,
		},
		{
			name:         "malloc failure",
			stderr:       "Malloc Failed!",
			expectedWarn: []string{"Malloc Failed!"},
			expectedErr:  ErrMallocFailed,
		},
		{
			name:         "permission denied",
			stderr:       "permission denied",
			expectedWarn: []string{"permission denied"},
			expectedErr:  ErrRequiresRoot,
		},
		{
			name:         "you must be root",
			stderr:       "You must be root",
			expectedWarn: []string{"You must be root"},
			expectedErr:  ErrRequiresRoot,
		},
		{
			name:         "requires root privileges",
			stderr:       "requires root privileges.",
			expectedWarn: []string{"requires root privileges."},
			expectedErr:  ErrRequiresRoot,
		},
		{
			name:         "could not resolve",
			stderr:       "could not resolve example.org",
			expectedWarn: []string{"could not resolve example.org"},
			expectedErr:  ErrResolveName,
		},
		{
			name:         "error resolving",
			stderr:       "error resolving target",
			expectedWarn: []string{"error resolving target"},
			expectedErr:  ErrResolveName,
		},
		{
			name:         "non fatal warnings are collected",
			stderr:       "warn one\nwarn two",
			expectedWarn: []string{"warn one", "warn two"},
			expectedErr:  nil,
		},
		{
			name:         "trims whitespace around lines",
			stderr:       "\n  warn one  \n  warn two  \n",
			expectedWarn: []string{"warn one", "warn two"},
			expectedErr:  nil,
		},
		{
			name:         "returns warnings up to fatal line",
			stderr:       "warn one\npermission denied\nwarn two",
			expectedWarn: []string{"warn one", "permission denied"},
			expectedErr:  ErrRequiresRoot,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			warnings, err := checkStdErr(bytes.NewBufferString(test.stderr))
			assert.Equal(t, test.expectedWarn, warnings)
			if test.expectedErr == nil {
				assert.NoError(t, err)
				return
			}
			assert.ErrorIs(t, err, test.expectedErr)
		})
	}
}

func TestMapRunError(t *testing.T) {
	t.Parallel()

	t.Run("nil input error returns nil", func(t *testing.T) {
		t.Parallel()

		assert.NoError(t, mapRunError(t.Context(), nil))
	})

	t.Run("deadline exceeded maps to timeout", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithDeadline(t.Context(), time.Now().Add(-time.Second))
		defer cancel()

		err := mapRunError(ctx, errors.New("run failed"))
		assert.ErrorIs(t, err, ErrScanTimeout)
	})

	t.Run("canceled maps to interrupt", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		err := mapRunError(ctx, errors.New("run failed"))
		assert.ErrorIs(t, err, ErrScanInterrupt)
	})

	t.Run("interrupt exit maps to interrupt", func(t *testing.T) {
		t.Parallel()

		err := mapRunError(t.Context(), errors.New("exit status 130"))
		assert.ErrorIs(t, err, ErrScanInterrupt)
	})

	t.Run("windows interrupt exit maps to interrupt", func(t *testing.T) {
		t.Parallel()

		err := mapRunError(t.Context(), errors.New("exit status 0xc000013a"))
		assert.ErrorIs(t, err, ErrScanInterrupt)
	})

	t.Run("canceled context takes precedence over interrupt exit", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		err := mapRunError(ctx, errors.New("exit status 130"))
		assert.ErrorIs(t, err, ErrScanInterrupt)
	})

	t.Run("other errors are returned unchanged", func(t *testing.T) {
		t.Parallel()

		runErr := errors.New("some other failure")
		err := mapRunError(t.Context(), runErr)
		assert.ErrorIs(t, err, runErr)
	})
}

func TestIsInterruptExit(t *testing.T) {
	t.Parallel()

	assert.False(t, isInterruptExit(nil))
	assert.True(t, isInterruptExit(errors.New("exit status 130")))
	assert.True(t, isInterruptExit(errors.New("exit status 0xc000013a")))
	assert.False(t, isInterruptExit(errors.New("exit status 1")))
}
