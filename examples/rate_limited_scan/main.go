package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Ullaakut/masscan"
)

// main runs a rate-limited scan with excludes and open-port filtering.
//
// Example output:
// 192.168.1.70 80/tcp open
// 192.168.1.70 443/tcp open
func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	scanner, err := masscan.NewScanner(
		masscan.WithTargets("192.168.1.0/24"),
		masscan.WithPorts("22", "80", "443"),
		masscan.WithExclude("192.168.1.1", "192.168.1.254"),
		masscan.WithRate(1000),
		masscan.WithWait(3),
		masscan.WithOpenOnly(),
		masscan.WithOutputFormat(masscan.OutputFormatJSON),
	)
	if err != nil {
		log.Fatalf("unable to create scanner: %v", err)
	}

	result, err := scanner.Run(ctx)
	if err != nil {
		log.Fatalf("scan failed: %v", err)
	}

	for _, warning := range result.Warnings() {
		log.Printf("warning: %s", warning)
	}

	if len(result.Hosts) == 0 {
		fmt.Println("No open ports found.")
		return
	}

	for _, host := range result.Hosts {
		for _, port := range host.Ports {
			fmt.Printf("%s %d/%s %s\n", host.Address, port.Number, port.Protocol, port.Status)
		}
	}
}
