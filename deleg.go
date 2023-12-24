package deleg

import (
	"context"

	"github.com/coredns/coredns/plugin"

	"github.com/miekg/dns"
)

// Deleg is a plugin that implements https://github.com/fl1ger/deleg/blob/main/draft-dnsop-deleg.md
type Deleg struct {
	Next plugin.Handler

	zones     []string
	responses []dns.RR
}

// ServeDNS implements the plugin.Handler interface.
func (d Deleg) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {

	pw := NewResponsePrinter(w, d)

	return plugin.NextOrFailure(d.Name(), d.Next, ctx, pw, r)
}

// Name implements the Handler interface.
func (d Deleg) Name() string { return "deleg" }

// ResponsePrinter wrap a dns.ResponseWriter and will write example to standard output when WriteMsg is called.
type ResponsePrinter struct {
	dns.ResponseWriter
	d Deleg
}

// NewResponsePrinter returns ResponseWriter.
func NewResponsePrinter(w dns.ResponseWriter, d Deleg) *ResponsePrinter {
	return &ResponsePrinter{w, d}
}

func (d Deleg) matches(owner string) string {
	for _, z := range d.zones {
		if dns.CountLabel(owner) == dns.CountLabel(z) && dns.IsSubDomain(z, owner) {
			return z
		}
	}
	return ""
}

// WriteMsg calls the underlying ResponseWriter's WriteMsg method and prints "example" to standard output.
func (r *ResponsePrinter) WriteMsg(res *dns.Msg) error {
	// Nothing to be done here.
	if len(res.Ns) == 0 {
		return r.ResponseWriter.WriteMsg(res)
	}

	for _, auth := range res.Ns {
		owner := auth.Header().Name
		rtype := auth.Header().Rrtype
		// not an NS record, tentatively try the next records
		if rtype != dns.TypeNS {
			continue
		}
		zone := r.d.matches(owner)
		//Let's assume that if there is a NS record, then there are all for the same owner name
		if zone == "" {
			return r.ResponseWriter.WriteMsg(res)
		}
		// We have a matching zone, adding the RRs to the Auth section
		responses := make([]dns.RR, 0)
		for _, rr := range r.d.responses {
			rr := dns.Copy(rr)
			rr.Header().Name = owner
			responses = append(responses, rr)
		}
		res.Ns = append(res.Ns, responses...)
		// and we are done.
		break
	}

	// Following back to writing the original response
	return r.ResponseWriter.WriteMsg(res)
}
