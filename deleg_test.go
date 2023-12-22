package deleg

import (
	"context"
	"testing"

	"github.com/coredns/coredns/plugin/pkg/dnstest"
	"github.com/coredns/coredns/plugin/test"

	"github.com/miekg/dns"
)

func TestDeleg(t *testing.T) {
	req := new(dns.Msg)
	req.SetQuestion("example.org.", dns.TypeA)
	d := &Deleg{}

	rec := dnstest.NewRecorder(&test.ResponseWriter{})
	_, err := d.ServeDNS(context.TODO(), rec, req)

	if err == nil {
		// request not handle, should be passed to next (non-existent) plugin.
		t.Errorf("Expected an error")
	}
}
