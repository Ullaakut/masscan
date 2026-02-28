package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Ullaakut/masscan"
)

// main demonstrates distributed scanning with shard and seed options.
//
// Example output:
// 45.33.32.156 80/tcp open
// Shard result count: 1
func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	scanner, err := masscan.NewScanner(
		masscan.WithTargets("45.33.32.156"),
		masscan.WithPorts("80", "443"),
		masscan.WithShard(1, 1),
		masscan.WithSeed(20260228),
		masscan.WithRate(3000),
		masscan.WithWait(2),
		masscan.WithOutputFormat(masscan.OutputFormatList),
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

	count := 0
	for _, host := range result.Hosts {
		for _, port := range host.Ports {
			fmt.Printf("%s %d/%s %s\n", host.Address, port.Number, port.Protocol, port.Status)
			count++
		}
	}

	fmt.Printf("Shard result count: %d\n", count)
}
