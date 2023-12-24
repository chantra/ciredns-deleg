package deleg

import (
	"testing"

	"github.com/coredns/caddy"
	"github.com/coredns/coredns/plugin/test"
	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
)

func TestSetup(t *testing.T) {

	testCases := []struct {
		name              string
		input             string
		shouldErr         bool
		expectedZones     []string
		expectedResponses map[string][]dns.RR
	}{
		{
			"empty deleg block without zone",
			`deleg`, false, nil, nil,
		},
		{"empty deleg block with zone", `deleg example.org`, false, []string{"example.org."}, nil},
		{"reverse zone infered from subnet", `dnssec 10.0.0.0/8`, false, []string{"10.in-addr.arpa."}, nil},
		{"valid basic deleg block",
			`deleg example.org {
				responses "example.org. 3600 IN TXT aaaaa"
			}`, false, []string{"example.org."}, map[string][]dns.RR{"example.org.": {test.TXT("example.org. 3600 IN TXT aaaaa")}},
		},
		{
			"valid basic deleg block with multiple responses",
			`deleg example.org {
				responses "example.org. 3600 IN TXT aaaaa" "example.org. 3600 IN TXT bbbbbb"
			}`, false, []string{"example.org."}, map[string][]dns.RR{"example.org.": {test.TXT("example.org. 3600 IN TXT aaaaa"), test.TXT("example.org. 3600 IN TXT bbbbbb")}},
		},
		// multiple zones associated with the same block.
		{
			"valid deleg with multiple zones",
			`deleg example.org example.com {
				responses "example.org. 3600 IN TXT spf1 -all"
			}`, false, []string{"example.org.", "example.com."}, map[string][]dns.RR{"example.org.": {test.TXT("example.org. 3600 IN TXT spf1 -all")}, "example.com.": {test.TXT("example.com. 3600 IN TXT spf1 -all")}},
		},
		// multiple zones associated with different records.
		{
			"multiple deleg blocks with different responses",
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

	for i, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c := caddy.NewTestController("dns", tc.input)
			delegs, err := delegParse(c)

			if tc.shouldErr && err == nil {
				t.Errorf("Test %d: Expected error but found %s for input %s", i, err, tc.input)
			}

			if err != nil {
				if !tc.shouldErr {
					t.Errorf("Test %d: Expected no error but found one for input %s. Error was: %v", i, tc.input, err)
				}

			}
			if !tc.shouldErr {
				var zones []string
				for k := range delegs {
					zones = append(zones, k)
				}
				assert.ElementsMatch(t, tc.expectedZones, zones, "Zones mismatch. Expected %s, actual %s", tc.expectedResponses, zones)

				for k := range delegs {
					assert.ElementsMatch(t, tc.expectedResponses[k], delegs[k], "Responses mismatch. Expected %s, actual %s", tc.expectedResponses[k], delegs[k])
				}

			}
		})
	}
}
