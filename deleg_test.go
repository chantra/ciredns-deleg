package deleg

import (
	"context"
	"reflect"
	"testing"

	"github.com/coredns/coredns/plugin"
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

// TestDelegResponse tests the response of the deleg plugin.
// checking that it adds the responses to the NS section of the response
// when expected.
func TestDelegResponse(t *testing.T) {

	testCases := []struct {
		name        string
		auth        []dns.RR
		zones       []string
		responses   []dns.RR
		shouldMatch bool
		expectedNs  []dns.RR // if nil, we expect auth + responses, otherwise, we expect expectedNs
	}{
		// Standard response with an NS record matching one of the zones
		{
			"Matching zone",
			[]dns.RR{test.NS("example.org. 3600 IN NS ns1.example.org.")},
			[]string{"example.org."},
			[]dns.RR{test.A("example.org. 3600 IN A 127.0.0.1")},
			true,
			nil,
		},
		// Check that we only add the records once.
		{
			"Matching zone multiple times",
			[]dns.RR{test.NS("example.org. 3600 IN NS ns1.example.org."), test.NS("example.org. 3600 IN NS ns2.example.org.")},
			[]string{"example.org."},
			[]dns.RR{test.A("example.org. 3600 IN A 127.0.0.1")},
			true,
			nil,
		},
		// Standard response with an NS record not matching any of the zones
		{
			"Not matching zone",
			[]dns.RR{test.NS("example.com. 3600 IN NS ns1.example.org.")},
			[]string{"example.org."},
			[]dns.RR{test.A("example.org. 3600 IN A 127.0.0.1")},
			false,
			nil,
		},
		// Standard response with an NS record matching one of the zones, the matching zone is not the first in the list.
		{
			"Multizone match",
			[]dns.RR{test.NS("example.org. 3600 IN NS ns1.example.org.")},
			[]string{"example.com.", "example.org."},
			[]dns.RR{test.A("example.org. 3600 IN A 127.0.0.1")},
			true,
			nil,
		},
		// Standard response with an NS record matching one of the zones, the matching zone is not the first in the list.
		// and we have other records along the NS record.
		{
			"Multizone match",
			[]dns.RR{test.A("example.com. 3600 IN A 127.0.0.1"), test.NS("example.org. 3600 IN NS ns1.example.org.")},
			[]string{"example.com.", "example.org."},
			[]dns.RR{test.A("example.org. 3600 IN A 127.0.0.1")},
			true,
			nil,
		},
		// Standard response with an NS record matching one of the zones, the aded record is not matching the owner of the NS record. Expected as we do not validate this currently.
		{
			"Multizone match wrong owner in response",
			[]dns.RR{test.NS("example.com. 3600 IN NS ns1.example.org.")},
			[]string{"example.com.", "example.org."},
			[]dns.RR{test.A("example.org. 3600 IN A 127.0.0.1")},
			true,
			[]dns.RR{test.NS("example.com. 3600 IN NS ns1.example.org."), test.A("example.com. 3600 IN A 127.0.0.1")},
		},
		// Standard response with a record matching one of the zones BUT it is not an NS record so we do not add anything.
		{
			"Multizone match wrong type in response",
			[]dns.RR{test.A("example.com. 3600 IN A 127.0.0.1")},
			[]string{"example.com.", "example.org."},
			[]dns.RR{test.AAAA("example.org. 3600 IN AAAA ::1")},
			false,
			nil,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := new(dns.Msg)
			req.SetQuestion("www.example.org.", dns.TypeA)
			d := &Deleg{
				zones:     tc.zones,
				responses: tc.responses,
			}

			d.Next = BackendHandler(tc.auth)

			ctx := context.TODO()

			rec := dnstest.NewRecorder(&test.ResponseWriter{})

			d.ServeDNS(ctx, rec, req)

			expectedNs := tc.expectedNs
			if expectedNs == nil {
				expectedNs = append(tc.auth, tc.responses...)
			}

			if tc.shouldMatch {
				if !reflect.DeepEqual(rec.Msg.Ns, expectedNs) {
					t.Errorf("Expecting %s got %s", expectedNs, rec.Msg.Ns)
				}
			} else {
				if !reflect.DeepEqual(rec.Msg.Ns, tc.auth) {
					t.Errorf("Expecting %s got %s", tc.auth, rec.Msg.Ns)
				}
			}

		})
	}

}

// A backend handler that will add a list of records to the NS section of the response.
func BackendHandler(auth []dns.RR) plugin.Handler {
	return plugin.HandlerFunc(func(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
		m := new(dns.Msg)
		m.SetReply(r)
		m.Response = true

		m.Ns = append(m.Ns, auth...)

		w.WriteMsg(m)
		return dns.RcodeSuccess, nil
	})
}
