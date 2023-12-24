package deleg

import (
	"reflect"
	"testing"

	"github.com/coredns/caddy"
	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
)

func TestSetup(t *testing.T) {

	tests := []struct {
		input             string
		shouldErr         bool
		expectedZones     []string
		expectedResponses [][]string
	}{
		{`deleg`, false, nil, nil},
		{`deleg example.org`, false, []string{"example.org."}, nil},
		{`dnssec 10.0.0.0/8`, false, []string{"10.in-addr.arpa."}, nil},
		{
			`deleg example.org {
				responses "example.org. 3600 IN TXT aaaaa"
			}`, false, []string{"example.org."}, [][]string{{"aaaaa"}},
		},
		{
			`deleg example.org {
				responses "example.org. 3600 IN TXT aaaaa" "example.org. 3600 IN TXT bbbbbb"
			}`, false, []string{"example.org."}, [][]string{{"aaaaa"}, {"bbbbbb"}},
		},
		{
			`deleg example.org {
				responses "example.org. 3600 IN TXT \"aaaaa\" \"bbbbbb\""
			}`, false, []string{"example.org."}, [][]string{{"aaaaa", "bbbbbb"}},
		},
		{
			`deleg example.org {
				responses "example.org. 3600 IN TXT \"spf1 -all\""
			}`, false, []string{"example.org."}, [][]string{{"spf1 -all"}},
		},
		// multiple zones associated with the same block.
		{
			`deleg example.org example.com {
				responses "example.org. 3600 IN TXT spf1 -all"
			}`, false, []string{"example.org.", "example.com."}, [][]string{{"spf1", "-all"}},
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
				for i, r := range test.expectedResponses {
					response := delegs[k][i].(*dns.TXT).Txt
					if !reflect.DeepEqual(r, response) {
						t.Errorf("Deleg not correctly set for input %s. Expected: '%s', actual: '%s'", test.input, r, response)
					}
				}
			}

		}
	}
}
