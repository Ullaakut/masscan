package masscan

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOptionsAppendExpectedArguments(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		option   Option
		expected []string
	}{
		{name: "targets", option: WithTargets("192.0.2.0/24"), expected: []string{"192.0.2.0/24"}},
		{name: "ports", option: WithPorts("80", "443"), expected: []string{"-p", "80,443"}},
		{name: "exclude", option: WithExclude("192.0.2.1", "192.0.2.2"), expected: []string{"--exclude", "192.0.2.1,192.0.2.2"}},
		{name: "rate", option: WithRate(5000), expected: []string{"--rate=5000"}},
		{name: "wait", option: WithWait(4), expected: []string{"--wait=4"}},
		{name: "interface", option: WithInterface("eth0"), expected: []string{"--interface=eth0"}},
		{name: "shard", option: WithShard(1, 3), expected: []string{"--shard=1/3"}},
		{name: "seed", option: WithSeed(12345), expected: []string{"--seed=12345"}},
		{name: "open-only", option: WithOpenOnly(), expected: []string{"--open-only"}},
		{name: "raw option", option: WithRawOption("--router-mac", "aa:bb:cc:dd:ee:ff"), expected: []string{"--router-mac", "aa:bb:cc:dd:ee:ff"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			scanner, err := NewScanner(WithBinaryPath("masscan"), test.option)
			require.NoError(t, err)
			assert.True(t, sliceHasSuffix(scanner.Args(), test.expected), "args %v do not end with expected suffix %v", scanner.Args(), test.expected)
		})
	}
}

func TestBuildArgsAppliesOutputFormatWhenMissing(t *testing.T) {
	t.Parallel()

	scanner, err := NewScanner(
		WithBinaryPath("masscan"),
		WithTargets("192.0.2.0/24"),
		WithOutputFormat(OutputFormatXML),
	)
	require.NoError(t, err)

	assert.True(t, sliceHasSuffix(scanner.buildArgs(), []string{"-oX", "-"}), "buildArgs did not append xml output args: %v", scanner.buildArgs())
}

func TestBuildArgsDoesNotOverrideExplicitOutput(t *testing.T) {
	t.Parallel()

	scanner, err := NewScanner(
		WithBinaryPath("masscan"),
		WithCustomArguments("-oL", "results.txt"),
		WithOutputFormat(OutputFormatJSON),
	)
	require.NoError(t, err)

	assert.True(t, sliceHasSuffix(scanner.buildArgs(), []string{"-oL", "results.txt"}), "buildArgs changed explicit output args: %v", scanner.buildArgs())
}

func sliceHasSuffix(values, suffix []string) bool {
	if len(values) < len(suffix) {
		return false
	}

	offset := len(values) - len(suffix)
	for index, expected := range suffix {
		if values[offset+index] != expected {
			return false
		}
	}

	return true
}
