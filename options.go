package masscan

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

// WithBinaryPath sets the masscan binary path for a scanner.
func WithBinaryPath(binaryPath string) Option {
	return func(s *Scanner) error {
		s.binaryPath = binaryPath
		return nil
	}
}

// WithFilterPort allows to set a custom function to filter out ports that
// don't fulfill a given condition. When the given function returns true,
// the port is kept, otherwise it is removed from the result. Can be used
// along with WithFilterHost.
func WithFilterPort(portFilter func(Port) bool) Option {
	return func(s *Scanner) error {
		s.portFilter = portFilter
		return nil
	}
}

// WithFilterHost allows to set a custom function to filter out hosts that
// don't fulfill a given condition. When the given function returns true,
// the host is kept, otherwise it is removed from the result. Can be used
// along with WithFilterPort.
func WithFilterHost(hostFilter func(Host) bool) Option {
	return func(s *Scanner) error {
		s.hostFilter = hostFilter
		return nil
	}
}

// WithTargets sets targets to scan (CIDR/range/single address/hostname).
func WithTargets(targets ...string) Option {
	return func(s *Scanner) error {
		normalizedTargets, err := normalizeTargets(targets, net.LookupIP)
		if err != nil {
			return err
		}

		s.args = append(s.args, normalizedTargets...)
		return nil
	}
}

func normalizeTargets(targets []string, lookupIP func(host string) ([]net.IP, error)) ([]string, error) {
	normalizedTargets := make([]string, 0, len(targets))

	for _, target := range targets {
		if isAddressTarget(target) {
			normalizedTargets = append(normalizedTargets, target)
			continue
		}

		resolvedIPs, err := lookupIP(target)
		if err != nil || len(resolvedIPs) == 0 {
			return nil, fmt.Errorf("%w: %s", ErrResolveName, target)
		}

		for _, resolvedIP := range resolvedIPs {
			normalizedTargets = append(normalizedTargets, resolvedIP.String())
		}
	}

	return normalizedTargets, nil
}

func isAddressTarget(target string) bool {
	if net.ParseIP(target) != nil {
		return true
	}

	if _, _, err := net.ParseCIDR(target); err == nil {
		return true
	}

	parts := strings.Split(target, "-")
	if len(parts) != 2 {
		return false
	}

	return net.ParseIP(parts[0]) != nil && net.ParseIP(parts[1]) != nil
}

// WithConfigPath sets the configuration file path (--conf).
func WithConfigPath(config string) Option {
	return appendKVOption("--conf", config)
}

// WithExclude excludes targets from scan (--exclude).
func WithExclude(excludes ...string) Option {
	return appendKVOption("--exclude", strings.Join(excludes, ","))
}

// WithExcludeFile excludes targets from a file (--excludefile).
func WithExcludeFile(path string) Option {
	return appendKVOption("--excludefile", path)
}

// WithPorts sets ports to scan (-p).
func WithPorts(ports ...string) Option {
	return appendKVOption("-p", strings.Join(ports, ","))
}

// WithTopPorts sets top port count (--top-ports).
func WithTopPorts(count int) Option {
	return appendEqualsOption("--top-ports", strconv.Itoa(count))
}

// WithRate sets packet sending rate (--rate).
func WithRate(maxRate int) Option {
	return appendEqualsOption("--rate", strconv.Itoa(maxRate))
}

// WithWait sets waiting time in seconds (--wait).
func WithWait(delay int) Option {
	return appendEqualsOption("--wait", strconv.Itoa(delay))
}

// WithInterface sets outgoing interface (--interface).
func WithInterface(iface string) Option {
	return appendEqualsOption("--interface", iface)
}

// WithSourceIP sets source IP address (--source-ip).
func WithSourceIP(sourceIP string) Option {
	return appendEqualsOption("--source-ip", sourceIP)
}

// WithSourcePort sets source port (--source-port).
func WithSourcePort(sourcePort int) Option {
	return appendEqualsOption("--source-port", strconv.Itoa(sourcePort))
}

// WithRouterMAC sets router MAC (--router-mac).
func WithRouterMAC(mac string) Option {
	return appendEqualsOption("--router-mac", mac)
}

// WithAdapterIP sets adapter IP address (--adapter-ip).
func WithAdapterIP(adapterIP string) Option {
	return appendEqualsOption("--adapter-ip", adapterIP)
}

// WithAdapterPort sets adapter/source port (--adapter-port).
func WithAdapterPort(port int) Option {
	return appendEqualsOption("--adapter-port", strconv.Itoa(port))
}

// WithAdapterMAC sets adapter MAC address (--adapter-mac).
func WithAdapterMAC(mac string) Option {
	return appendEqualsOption("--adapter-mac", mac)
}

// WithShard enables distributed scanning (--shard).
func WithShard(x, y int) Option {
	return appendEqualsOption("--shard", fmt.Sprintf("%d/%d", x, y))
}

// WithSeed sets randomization seed (--seed).
func WithSeed(seed int) Option {
	return appendEqualsOption("--seed", strconv.Itoa(seed))
}

// WithBanners enables banner grabbing (--banners).
func WithBanners() Option {
	return appendArgsOption("--banners")
}

// WithPing scans in ping mode only (--ping).
func WithPing() Option {
	return appendArgsOption("--ping")
}

// WithResumeIndex resumes from paused index (--resume-index).
func WithResumeIndex(index int) Option {
	return appendEqualsOption("--resume-index", strconv.Itoa(index))
}

// WithResumeCount scans count from resumed point (--resume-count).
func WithResumeCount(count int) Option {
	return appendEqualsOption("--resume-count", strconv.Itoa(count))
}

// WithOpenOnly limits output to open ports only (--open-only).
func WithOpenOnly() Option {
	return appendArgsOption("--open-only")
}

// WithDebug enables masscan debug mode (--debug).
func WithDebug() Option {
	return appendArgsOption("--debug")
}

// WithOutputFormat sets output format for parsing. If explicit -o* args are passed,
// masscan will use them instead.
func WithOutputFormat(format OutputFormat) Option {
	return func(s *Scanner) error {
		if _, ok := format.outputFlag(); !ok {
			return fmt.Errorf("%w: %s", ErrUnsupportedOutputFormat, format)
		}
		s.output = format
		return nil
	}
}

// WithRawFlag appends a raw masscan flag with no value.
func WithRawFlag(flag string) Option {
	return appendArgsOption(flag)
}

// WithRawOption appends a raw masscan option and value pair.
func WithRawOption(flag, value string) Option {
	return appendKVOption(flag, value)
}

func appendArgsOption(args ...string) Option {
	return func(s *Scanner) error {
		s.args = append(s.args, args...)
		return nil
	}
}

func appendKVOption(flag, value string) Option {
	return appendArgsOption(flag, value)
}

func appendEqualsOption(flag, value string) Option {
	return appendArgsOption(fmt.Sprintf("%s=%s", flag, value))
}
