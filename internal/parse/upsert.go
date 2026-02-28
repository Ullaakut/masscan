package parse

func upsertPort(result *Result, address, timestamp string, port Port) {
	for idx := range result.Hosts {
		if result.Hosts[idx].Address != address {
			continue
		}

		if result.Hosts[idx].Timestamp == "" && timestamp != "" {
			result.Hosts[idx].Timestamp = timestamp
		}

		result.Hosts[idx].Ports = append(result.Hosts[idx].Ports, port)
		return
	}

	result.Hosts = append(result.Hosts, Host{
		Address:   address,
		Timestamp: timestamp,
		Ports:     []Port{port},
	})
}
