package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Ullaakut/masscan"
)

// main runs a banner-enabled scan against a public target.
//
// Example output:
// 45.33.32.156 80/tcp open reason=syn-ack ttl=43
func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	scanner, err := masscan.NewScanner(
		masscan.WithTargets("45.33.32.156"),
		masscan.WithPorts("22", "80", "443"),
		masscan.WithBanners(),
		masscan.WithRate(500),
		masscan.WithWait(5),
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
			fmt.Printf("%s %d/%s %s reason=%s ttl=%d\n", host.Address, port.Number, port.Protocol, port.Status, port.Reason, port.ReasonTTL)
		}
	}
}
