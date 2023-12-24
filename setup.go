package deleg

import (
	"github.com/coredns/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"

	"github.com/miekg/dns"
)

func init() { plugin.Register("deleg", setup) }

func setup(c *caddy.Controller) error {
	delegs, err := delegParse(c)

	if err != nil {
		return plugin.Error("deleg", err)
	}
	d := Deleg{
		delegs: delegs,
	}

	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		d.Next = next
		return d
	})

	return nil
}

func rewriteResponsesOwner(responses []dns.RR, owner string) []dns.RR {
	rewrittenResponses := []dns.RR{}
	for _, r := range responses {
		r := dns.Copy(r)
		r.Header().Name = owner
		rewrittenResponses = append(rewrittenResponses, r)
	}
	return rewrittenResponses
}

// delegParse parses the configuration for delegations and returns a map of zones
// to their corresponding DNS resource records (RR). It expects a Caddy controller
// as input and returns the map of delegations and an error if any.
func delegParse(c *caddy.Controller) (map[string][]dns.RR, error) {
	var delegs map[string][]dns.RR = make(map[string][]dns.RR)

	i := 0
	// `deleg`
	for c.Next() {
		responses := []dns.RR{}
		if i > 0 {
			return nil, plugin.ErrOnce
		}
		i++

		zones := plugin.OriginsFromArgsOrServerBlock(c.RemainingArgs(), c.ServerBlockKeys)

		for c.NextBlock() {
			switch x := c.Val(); x {
			case "responses":
				r, e := responseParse(c)
				if e != nil {
					return nil, e
				}
				responses = append(responses, r...)

			default:
				return nil, c.Errf("unknown property '%s'", x)
			}
		}

		for _, z := range zones {
			zoneName := dns.CanonicalName(z)
			delegs[zoneName] = rewriteResponsesOwner(responses, zoneName)
		}

	}

	return delegs, nil
}

func responseParse(c *caddy.Controller) ([]dns.RR, error) {
	responses := []dns.RR{}

	resps := c.RemainingArgs()

	if len(resps) == 0 {
		return nil, c.ArgErr()
	}

	for _, resp := range resps {
		r, err := dns.NewRR(resp)
		if err != nil {
			return nil, err
		}
		responses = append(responses, r)
	}

	return responses, nil
}
