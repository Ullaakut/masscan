package parse

import (
	"encoding/xml"
	"fmt"
)

type xmlRun struct {
	Hosts []xmlHost `xml:"host"`
}

type xmlHost struct {
	Addresses []xmlAddress `xml:"address"`
	Ports     []xmlPort    `xml:"ports>port"`
	StartTime string       `xml:"starttime,attr"`
	EndTime   string       `xml:"endtime,attr"`
}

type xmlAddress struct {
	Addr string `xml:"addr,attr"`
}

type xmlPort struct {
	Protocol string   `xml:"protocol,attr"`
	PortID   int      `xml:"portid,attr"`
	State    xmlState `xml:"state"`
}

type xmlState struct {
	Status string `xml:"state,attr"`
	Reason string `xml:"reason,attr"`
	TTL    int    `xml:"reason_ttl,attr"`
}

func parseXML(contents []byte) (*Result, error) {
	var parsed xmlRun
	err := xml.Unmarshal(contents, &parsed)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidOutput, err)
	}

	result := &Result{}
	for _, host := range parsed.Hosts {
		timestamp := host.StartTime
		if timestamp == "" {
			timestamp = host.EndTime
		}

		mapped := Host{Timestamp: timestamp}
		if len(host.Addresses) > 0 {
			mapped.Address = host.Addresses[0].Addr
		}

		for _, port := range host.Ports {
			mapped.Ports = append(mapped.Ports, Port{
				Number:    port.PortID,
				Protocol:  port.Protocol,
				Status:    port.State.Status,
				Reason:    port.State.Reason,
				ReasonTTL: port.State.TTL,
			})
		}

		result.Hosts = append(result.Hosts, mapped)
	}

	return result, nil
}
