package masscan

import (
	"fmt"
	"strings"
)

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

// WithTargets sets targets to scan (CIDR/range/single address).
func WithTargets(targets ...string) Option {
	return appendArgsOption(targets...)
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
	return appendEqualsOption("--top-ports", fmt.Sprintf("%d", count))
}

// WithRate sets packet sending rate (--rate).
func WithRate(maxRate int) Option {
	return appendEqualsOption("--rate", fmt.Sprintf("%d", maxRate))
}

// WithWait sets waiting time in seconds (--wait).
func WithWait(delay int) Option {
	return appendEqualsOption("--wait", fmt.Sprintf("%d", delay))
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
	return appendEqualsOption("--source-port", fmt.Sprintf("%d", sourcePort))
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
	return appendEqualsOption("--adapter-port", fmt.Sprintf("%d", port))
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
	return appendEqualsOption("--seed", fmt.Sprintf("%d", seed))
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
	return appendEqualsOption("--resume-index", fmt.Sprintf("%d", index))
}

// WithResumeCount scans count from resumed point (--resume-count).
func WithResumeCount(count int) Option {
	return appendEqualsOption("--resume-count", fmt.Sprintf("%d", count))
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
