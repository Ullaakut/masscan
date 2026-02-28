package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Ullaakut/masscan"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	scanner, err := masscan.NewScanner(
		masscan.WithTargets("10.0.0.0/8"),
		masscan.WithTopPorts(100),
		masscan.WithShard(1, 4),
		masscan.WithSeed(20260228),
		masscan.WithRate(25000),
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
