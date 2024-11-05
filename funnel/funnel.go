package funnel

import (
	"context"
	"net"
	"sync/atomic"

	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/request"
	"github.com/miekg/dns"
)

type Funnel struct {
	destinations []net.IP
	zones        []string
	ttl          uint32

	idx               uint64
	destinationsCount int

	next plugin.Handler
}

func (f *Funnel) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	req := request.Request{W: w, Req: r}
	if req.QType() != dns.TypeA {
		return plugin.NextOrFailure(f.Name(), f.next, ctx, w, r)
	}

	qname := req.Name()
	zone := plugin.Zones(f.zones).Matches(qname)

	if zone == "" {
		return plugin.NextOrFailure(f.Name(), f.next, ctx, w, r)
	}

	atomic.AddUint64(&f.idx, 1)

	m := new(dns.Msg)
	m.SetReply(r)
	m.Authoritative = true
	m.Answer = []dns.RR{
		&dns.A{
			Hdr: dns.RR_Header{Name: qname, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: f.ttl},
			A:   f.destinations[(int(f.idx)-1)%f.destinationsCount],
		},
	}

	w.WriteMsg(m)

	return dns.RcodeSuccess, nil
}

func (f *Funnel) Name() string { return "funnel" }
