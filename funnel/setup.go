package funnel

import (
	"errors"
	"fmt"
	"net"
	"strconv"

	"github.com/coredns/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
)

func init() {
	plugin.Register("funnel", setup)
}

func setup(c *caddy.Controller) error {
	var (
		fnl               *Funnel = new(Funnel)
		trash             string
		destinationString string
	)

	c.Args(&trash, &destinationString)
	if destinationString == "" {
		return plugin.Error("funnel", c.ArgErr())
	}

	destination := net.ParseIP(destinationString)
	if destination == nil {
		// resolve the destination
		ips, err := net.LookupIP(destinationString)
		if err != nil {
			return plugin.Error("funnel", fmt.Errorf("failed to resolve destination (%s): %v", destinationString, err))
		}
		if len(ips) == 0 {
			return plugin.Error("funnel", errors.New("no IP addresses found for destination"))
		}

		for _, ip := range ips {
			if ip.To4() != nil {
				fnl.destinations = append(fnl.destinations, ip)
			}
		}

		if len(fnl.destinations) == 0 {
			return plugin.Error("funnel", errors.New("no IPv4 addresses found for destination"))
		}
	} else {
		if destination.To4() == nil {
			return plugin.Error("funnel", errors.New("destination must be an IPv4 address"))
		}

		fnl.destinations = append(fnl.destinations, destination)
	}

	fnl.destinationsCount = len(fnl.destinations)
	if fnl.destinationsCount == 0 {
		return plugin.Error("funnel", errors.New("no destinations specified"))
	}
	fnl.zones = plugin.OriginsFromArgsOrServerBlock(c.RemainingArgs(), c.ServerBlockKeys)

	for c.NextBlock() {
		args := append([]string{c.Val()}, c.RemainingArgs()...)
		if len(args) == 0 {
			return plugin.Error("funnel", c.ArgErr())
		}

		switch args[0] {
		case "ttl":
			if len(args) != 2 {
				return plugin.Error("funnel", c.ArgErr())
			}

			ttlInt, err := strconv.ParseInt(args[1], 10, 32)
			if err != nil {
				return plugin.Error("funnel", fmt.Errorf("failed to parse TTL: %v", err))
			}

			fnl.ttl = uint32(ttlInt)
		default:
			return plugin.Error("funnel", c.ArgErr())
		}
	}

	if fnl.ttl == 0 {
		fnl.ttl = 300
	}

	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		fnl.next = next
		return fnl
	})

	return nil
}
