package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Ullaakut/masscan"
)

// main demonstrates interface, source, and adapter-level scan controls.
//
// Example output:
// 45.33.32.156 80/tcp open
func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	scanner, err := masscan.NewScanner(
		masscan.WithTargets("45.33.32.156"),
		masscan.WithPorts("80", "443"),
		masscan.WithInterface("eth0"),
		masscan.WithSourcePort(61000),
		masscan.WithRate(2000),
		masscan.WithWait(2),
		masscan.WithOutputFormat(masscan.OutputFormatJSON),
		masscan.WithRawFlag("--open-only"),
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
