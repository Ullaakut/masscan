package parse

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

func parseList(contents []byte) (*Result, error) {
	result := &Result{}
	for line := range strings.SplitSeq(string(contents), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}

		status := fields[0]
		proto := fields[1]
		portValue := fields[2]
		ip := fields[3]
		timestamp := ""
		if len(fields) > 4 {
			timestamp = fields[4]
		}

		port, err := strconv.Atoi(portValue)
		if err != nil {
			continue
		}

		upsertPort(result, ip, timestamp, Port{Number: port, Protocol: proto, Status: status})
	}

	if len(result.Hosts) == 0 {
		stdoutResult, err := parseStdoutLines(contents)
		if err != nil {
			return nil, err
		}
		return stdoutResult, nil
	}

	return result, nil
}

var discoveredLine = regexp.MustCompile(`Discovered\s+(\S+)\s+port\s+(\d+)\/(\w+)\s+on\s+([\w\.:\-]+)`)

func parseStdoutLines(contents []byte) (*Result, error) {
	result := &Result{}
	for line := range strings.SplitSeq(string(contents), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		matches := discoveredLine.FindStringSubmatch(line)
		if len(matches) != 5 {
			continue
		}

		port, err := strconv.Atoi(matches[2])
		if err != nil {
			continue
		}

		upsertPort(result, matches[4], "", Port{Status: matches[1], Number: port, Protocol: matches[3]})
	}

	if len(result.Hosts) == 0 {
		return nil, fmt.Errorf("%w: no parsable records found", ErrInvalidOutput)
	}

	return result, nil
}
