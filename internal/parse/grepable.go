package parse

import (
	"fmt"
	"strconv"
	"strings"
)

func parseGrepable(contents []byte) (*Result, error) {
	result := &Result{}
	for line := range strings.SplitSeq(string(contents), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if !strings.HasPrefix(line, "Host:") {
			continue
		}

		hostPart := strings.TrimPrefix(line, "Host:")
		hostPart = strings.TrimSpace(hostPart)
		hostFields := strings.Fields(hostPart)
		if len(hostFields) == 0 {
			continue
		}

		address := hostFields[0]
		portsIndex := strings.Index(line, "Ports:")
		if portsIndex == -1 {
			continue
		}

		portsSection := strings.TrimSpace(line[portsIndex+len("Ports:"):])
		for portSpec := range strings.SplitSeq(portsSection, ",") {
			parts := strings.Split(strings.TrimSpace(portSpec), "/")
			if len(parts) < 3 {
				continue
			}

			port, err := strconv.Atoi(parts[0])
			if err != nil {
				continue
			}

			upsertPort(result, address, "", Port{
				Number:   port,
				Status:   parts[1],
				Protocol: parts[2],
			})
		}
	}

	if len(result.Hosts) == 0 {
		return nil, fmt.Errorf("%w: no parsable grepable records found", ErrInvalidOutput)
	}

	return result, nil
}
