package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Ullaakut/masscan"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	scanner, err := masscan.NewScanner(
		masscan.WithTargets("192.168.1.0/24"),
		masscan.WithPorts("53", "161"),
		masscan.WithInterface("eth0"),
		masscan.WithSourceIP("192.168.1.10"),
		masscan.WithSourcePort(61000),
		masscan.WithAdapterIP("192.168.1.10"),
		masscan.WithAdapterPort(61000),
		masscan.WithAdapterMAC("00:11:22:33:44:55"),
		masscan.WithRouterMAC("aa:bb:cc:dd:ee:ff"),
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

	for _, host := range result.Hosts {
		for _, port := range host.Ports {
			fmt.Printf("%s %d/%s %s\n", host.Address, port.Number, port.Protocol, port.Status)
		}
	}
}
