package funnel

import (
	"context"
	"net"

	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/request"
	"github.com/miekg/dns"
)

type Funnel struct {
	destination net.IP
	zones       []string
	ttl         uint32

	Next plugin.Handler
}

func (f *Funnel) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	req := request.Request{W: w, Req: r}
	qname := req.Name()
	zone := plugin.Zones(f.zones).Matches(qname)

	if zone == "" {
		return plugin.NextOrFailure(f.Name(), f.Next, ctx, w, r)
	}

	m := new(dns.Msg)
	m.SetReply(r)
	m.Authoritative = true
	m.Answer = []dns.RR{
		&dns.A{
			Hdr: dns.RR_Header{Name: qname, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: f.ttl},
			A:   f.destination,
		},
	}

	w.WriteMsg(m)

	return dns.RcodeSuccess, nil
}

func (f *Funnel) Name() string { return "funnel" }
