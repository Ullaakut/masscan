# masscan

<p align="center">
  <a href="LICENSE">
    <img src="https://img.shields.io/badge/license-MIT-blue.svg?style=flat" alt="MIT license" />
  </a>
  <a href="https://pkg.go.dev/github.com/Ullaakut/masscan">
    <img src="https://pkg.go.dev/badge/github.com/Ullaakut/masscan" alt="PkgGoDev github.com/Ullaakut/masscan" />
  </a>
  <a href="https://goreportcard.com/report/github.com/Ullaakut/masscan">
    <img src="https://goreportcard.com/badge/github.com/Ullaakut/masscan" alt="Go Report Card" />
  </a>
</p>

This library provides idiomatic `masscan` bindings for Go developers.
It shells out to the `masscan` binary and parses scan output into Go structs.

## What is masscan

Masscan is a high-speed TCP port scanner focused on quickly identifying open ports across large target ranges.
Compared to nmap, masscan prioritizes speed and broad coverage over deep service fingerprinting.

## How it works

- The package executes `masscan` via `os/exec`.
- `masscan` must be installed and available in your `PATH` (or configured with `WithBinaryPath`).
- The scanner runs synchronously with `Run(ctx)`.
- Output is parsed into `Run`, `Host`, and `Port` structs.

## Supported output formats

The parser supports these formats:

- JSON (`-oJ`) — recommended default
- XML (`-oX`)
- List (`-oL`)
- Grepable (`-oG`)
- Plain discovered lines from stdout (fallback parsing)

Binary output (`-oB`) is not parsed directly and returns `ErrUnsupportedOutputFormat`.

## Privileges

Masscan usually requires raw socket privileges.
If you see `ErrRequiresRoot`, run with elevated privileges (for example, via `sudo`) or configure capabilities for your environment.

## Installation

```bash
go get github.com/Ullaakut/masscan
```

## Quick start

```go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Ullaakut/masscan"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	scanner, err := masscan.NewScanner(
		masscan.WithTargets("scanme.nmap.org"),
		masscan.WithPorts("80", "443"),
		masscan.WithRate(1000),
		masscan.WithWait(3),
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

	for _, host := range result.Hosts {
		for _, port := range host.Ports {
			fmt.Printf("%s %d/%s %s\n", host.Address, port.Number, port.Protocol, port.Status)
		}
	}
}
```

## Options API

The package provides typed options for common masscan flags, including:

- `WithTargets`, `WithPorts`, `WithTopPorts`
- `WithExclude`, `WithExcludeFile`
- `WithRate`, `WithWait`, `WithOpenOnly`
- `WithInterface`, `WithSourceIP`, `WithSourcePort`
- `WithAdapterIP`, `WithAdapterPort`, `WithAdapterMAC`, `WithRouterMAC`
- `WithShard`, `WithSeed`, `WithResumeIndex`, `WithResumeCount`
- `WithBanners`, `WithPing`, `WithDebug`
- `WithOutputFormat`

If a masscan flag is not covered yet, use:

- `WithRawFlag("--some-flag")`
- `WithRawOption("--some-option", "value")`

## Examples

- [Basic scan](examples/basic_scan/main.go)
- [Rate-limited LAN scan](examples/rate_limited_scan/main.go)
- [Banner scan](examples/banner_scan/main.go)
- [Distributed shard scan](examples/distributed_shard/main.go)
- [Interface and source control scan](examples/interface_scan/main.go)

Run an example:

```bash
go run ./examples/basic_scan/main.go
```

For privileged scans:

```bash
sudo env "PATH=$PATH" go run ./examples/basic_scan/main.go
```

## Error handling

Common sentinel errors include:

- `ErrMasscanNotInstalled`
- `ErrScanTimeout`
- `ErrScanInterrupt`
- `ErrParseOutput`
- `ErrInvalidOutput`
- `ErrUnsupportedOutputFormat`
- `ErrRequiresRoot`
- `ErrResolveName`
- `ErrMallocFailed`

## External resources

- [Masscan repository](https://github.com/robertdavidgraham/masscan)
- [Masscan man page and usage docs](https://github.com/robertdavidgraham/masscan)
