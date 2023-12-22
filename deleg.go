package deleg

import (
	"context"

	"github.com/coredns/coredns/plugin"

	"github.com/miekg/dns"
)

// Deleg is a plugin that implements https://github.com/fl1ger/deleg/blob/main/draft-dnsop-deleg.md
type Deleg struct {
	Next plugin.Handler
}

// ServeDNS implements the plugin.Handler interface.
func (d Deleg) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	return plugin.NextOrFailure(d.Name(), d.Next, ctx, w, r)
}

// Name implements the Handler interface.
func (d Deleg) Name() string { return "deleg" }
