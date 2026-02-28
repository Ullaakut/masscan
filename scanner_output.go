package masscan

import "strings"

type outputConfig struct {
	format OutputFormat
	path   string
	setBy  string
}

func (s *Scanner) buildArgs() []string {
	args := append([]string{}, s.args...)
	cfg := detectOutputConfig(args)
	if cfg != nil {
		return args
	}

	flag, _ := s.output.outputFlag()
	path := "-"
	if s.toFile != nil {
		path = *s.toFile
	}
	args = append(args, flag, path)
	return args
}

func detectOutputConfig(args []string) *outputConfig {
	var config *outputConfig
	for idx := 0; idx < len(args); idx++ {
		arg := args[idx]
		if format, path, ok := outputArgument(arg, args, idx); ok {
			if path == "" {
				if idx+1 < len(args) {
					path = args[idx+1]
					idx++
				} else {
					path = "-"
				}
			}
			config = &outputConfig{format: format, path: path, setBy: arg}
		}
	}

	return config
}

func outputArgument(current string, args []string, index int) (OutputFormat, string, bool) {
	mapping := []struct {
		prefix string
		format OutputFormat
	}{
		{prefix: "-oX", format: OutputFormatXML},
		{prefix: "-oJ", format: OutputFormatJSON},
		{prefix: "-oL", format: OutputFormatList},
		{prefix: "-oG", format: OutputFormatGrepable},
		{prefix: "-oB", format: OutputFormatBinary},
	}

	for _, entry := range mapping {
		if current == entry.prefix {
			if index+1 < len(args) {
				return entry.format, args[index+1], true
			}
			return entry.format, "", true
		}

		if strings.HasPrefix(current, entry.prefix) && len(current) > len(entry.prefix) {
			return entry.format, current[len(entry.prefix):], true
		}
	}

	return OutputFormatUnknown, "", false
}

func (s *Scanner) outputConfig() outputConfig {
	args := s.buildArgs()
	if cfg := detectOutputConfig(args); cfg != nil {
		return *cfg
	}

	path := "-"
	if s.toFile != nil {
		path = *s.toFile
	}

	return outputConfig{format: s.output, path: path, setBy: "default"}
}
