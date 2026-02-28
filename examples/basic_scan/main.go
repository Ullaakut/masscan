package main

import (
	"context"
	"fmt"
	"log"

	"github.com/Ullaakut/masscan"
)

func main() {
	scanner, err := masscan.NewScanner(
		masscan.WithTargets("192.168.1.0/24"),
		masscan.WithPorts("80"),
	)
	if err != nil {
		log.Fatalf("unable to create masscan scanner: %v", err)
	}

	result, err := scanner.Run(context.Background())
	if err != nil {
		log.Fatalf("masscan encountered an error: %v", err)
	}

	for _, host := range result.Hosts {
		for _, port := range host.Ports {
			fmt.Printf("Host: %s Port: %d/%s (%s)\n", host.Address, port.Number, port.Protocol, port.Status)
		}
	}
}
