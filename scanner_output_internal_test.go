package masscan

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDetectOutputConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		args      []string
		expectNil bool
		format    OutputFormat
		path      string
		setBy     string
	}{
		{
			name:      "no output arguments",
			args:      []string{"-p80", "192.0.2.1"},
			expectNil: true,
		},
		{
			name:   "flag with separate path",
			args:   []string{"-p80", "-oX", "scan.xml"},
			format: OutputFormatXML,
			path:   "scan.xml",
			setBy:  "-oX",
		},
		{
			name:   "flag with inline path",
			args:   []string{"-oJscan.json"},
			format: OutputFormatJSON,
			path:   "scan.json",
			setBy:  "-oJscan.json",
		},
		{
			name:   "flag without path defaults to stdout",
			args:   []string{"-oG"},
			format: OutputFormatGrepable,
			path:   "-",
			setBy:  "-oG",
		},
		{
			name:   "last output flag wins",
			args:   []string{"-oX", "a.xml", "-oL", "b.list"},
			format: OutputFormatList,
			path:   "b.list",
			setBy:  "-oL",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			cfg := detectOutputConfig(test.args)
			if test.expectNil {
				assert.Nil(t, cfg)
				return
			}

			require.NotNil(t, cfg)
			assert.Equal(t, test.format, cfg.format)
			assert.Equal(t, test.path, cfg.path)
			assert.Equal(t, test.setBy, cfg.setBy)
		})
	}
}

func TestBuildArgs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		scanner        Scanner
		expectedSuffix []string
		expectedLen    int
	}{
		{
			name: "keeps explicit output arguments unchanged",
			scanner: Scanner{
				args:   []string{"-p80", "-oL", "results.txt"},
				output: OutputFormatJSON,
			},
			expectedSuffix: []string{"-oL", "results.txt"},
			expectedLen:    3,
		},
		{
			name: "appends default output to stdout when missing",
			scanner: Scanner{
				args:   []string{"-p80"},
				output: OutputFormatXML,
			},
			expectedSuffix: []string{"-oX", "-"},
			expectedLen:    3,
		},
		{
			name: "appends output file path when configured",
			scanner: func() Scanner {
				path := "scan.json"
				return Scanner{
					args:   []string{"-p443"},
					output: OutputFormatJSON,
					toFile: &path,
				}
			}(),
			expectedSuffix: []string{"-oJ", "scan.json"},
			expectedLen:    3,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			actual := test.scanner.buildArgs()
			assert.Equal(t, test.expectedLen, len(actual))
			assert.True(t, sliceHasSuffix(actual, test.expectedSuffix), "args %v do not end with suffix %v", actual, test.expectedSuffix)
		})
	}
}
