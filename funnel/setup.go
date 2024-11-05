package funnel

import (
	"errors"
	"fmt"
	"net"

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
		destinationString string
	)

	c.Args(&destinationString)
	if destinationString == "" {
		return plugin.Error("funnel", c.ArgErr())
	}

	destination := net.ParseIP(destinationString)
	if destination == nil {
		// resolve the destination
		ips, err := net.LookupIP(destinationString)
		if err != nil {
			return plugin.Error("funnel", fmt.Errorf("failed to resolve destination: %v", err))
		}
		if len(ips) == 0 {
			return plugin.Error("funnel", errors.New("no IP addresses found for destination"))
		}

		for _, ip := range ips {
			if ip.To4() != nil {
				destination = ip
				break
			}
		}
		if destination == nil {
			return plugin.Error("funnel", errors.New("no IPv4 address found for destination"))
		}
	} else {
		if destination.To4() == nil {
			return plugin.Error("funnel", errors.New("destination must be an IPv4 address"))
		}
	}

	fnl.destination = destination

	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		fnl.Next = next
		return fnl
	})

	return nil
}
