package masscan

import (
	"errors"
	"fmt"

	"github.com/Ullaakut/masscan/internal/parse"
)

func parseOutput(contents []byte, format OutputFormat) (*Run, error) {
	parsed, err := parse.Output(contents, parse.Format(format))
	if err != nil {
		switch {
		case errors.Is(err, parse.ErrInvalidOutput):
			return nil, fmt.Errorf("%w: %w", ErrInvalidOutput, err)
		case errors.Is(err, parse.ErrUnsupportedFormat):
			return nil, fmt.Errorf("%w: %w", ErrUnsupportedOutputFormat, err)
		default:
			return nil, err
		}
	}

	result := &Run{}
	for _, host := range parsed.Hosts {
		mapped := Host{Address: host.Address, Timestamp: host.Timestamp}
		for _, port := range host.Ports {
			mapped.Ports = append(mapped.Ports, Port{
				Number:    port.Number,
				Protocol:  port.Protocol,
				Status:    port.Status,
				Reason:    port.Reason,
				ReasonTTL: port.ReasonTTL,
			})
		}
		result.Hosts = append(result.Hosts, mapped)
	}

	return result, nil
}
