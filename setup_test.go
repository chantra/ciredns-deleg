package deleg

import (
	"testing"

	"github.com/coredns/caddy"
	"github.com/coredns/coredns/plugin/test"
	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
)

func TestSetup(t *testing.T) {

	tests := []struct {
		input             string
		shouldErr         bool
		expectedZones     []string
		expectedResponses map[string][]dns.RR
	}{
		{`deleg`, false, nil, nil},
		{`deleg example.org`, false, []string{"example.org."}, nil},
		{`dnssec 10.0.0.0/8`, false, []string{"10.in-addr.arpa."}, nil},
		{
			`deleg example.org {
				responses "example.org. 3600 IN TXT aaaaa"
			}`, false, []string{"example.org."}, map[string][]dns.RR{"example.org.": {test.TXT("example.org. 3600 IN TXT aaaaa")}},
		},
		{
			`deleg example.org {
				responses "example.org. 3600 IN TXT aaaaa" "example.org. 3600 IN TXT bbbbbb"
			}`, false, []string{"example.org."}, map[string][]dns.RR{"example.org.": {test.TXT("example.org. 3600 IN TXT aaaaa"), test.TXT("example.org. 3600 IN TXT bbbbbb")}},
		},
		{
			`deleg example.org {
				responses "example.org. 3600 IN TXT \"aaaaa\" \"bbbbbb\""
			}`, false, []string{"example.org."}, map[string][]dns.RR{"example.org.": {test.TXT("example.org. 3600 IN TXT \"aaaaa\" \"bbbbbb\"")}},
		},
		{
			`deleg example.org {
				responses "example.org. 3600 IN TXT \"spf1 -all\""
			}`, false, []string{"example.org."}, map[string][]dns.RR{"example.org.": {test.TXT("example.org. 3600 IN TXT \"spf1 -all\"")}},
		},
		// multiple zones associated with the same block.
		{
			`deleg example.org example.com {
				responses "example.org. 3600 IN TXT spf1 -all"
			}`, false, []string{"example.org.", "example.com."}, map[string][]dns.RR{"example.org.": {test.TXT("example.org. 3600 IN TXT spf1 -all")}, "example.com.": {test.TXT("example.com. 3600 IN TXT spf1 -all")}},
		},

		// multiple zones associated with different records.
		{
			`deleg example.org example.com {
				responses "example.org. 3600 IN TXT org"
			}
			deleg example.net  {
				responses "example.net. 3600 IN TXT net"
			}`, false, []string{"example.org.", "example.com.", "example.net."},
			map[string][]dns.RR{
				"example.org.": {test.TXT("example.org. 3600 IN TXT org")},
				"example.com.": {test.TXT("example.com. 3600 IN TXT org")},
				"example.net.": {test.TXT("example.net. 3600 IN TXT net")},
			},
		},
	}

	for i, test := range tests {
		c := caddy.NewTestController("dns", test.input)
		delegs, err := delegParse(c)

		if test.shouldErr && err == nil {
			t.Errorf("Test %d: Expected error but found %s for input %s", i, err, test.input)
		}

		if err != nil {
			if !test.shouldErr {
				t.Errorf("Test %d: Expected no error but found one for input %s. Error was: %v", i, test.input, err)
			}

		}
		if !test.shouldErr {
			var zones []string
			for k := range delegs {
				zones = append(zones, k)
			}
			assert.ElementsMatch(t, test.expectedZones, zones, "Zones mismatch. Expected %s, actual %s", test.expectedResponses, zones)

			for k := range delegs {
				assert.ElementsMatch(t, test.expectedResponses[k], delegs[k], "Responses mismatch. Expected %s, actual %s", test.expectedResponses[k], delegs[k])
			}

		}
	}
}
