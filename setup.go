package deleg

import (
	"github.com/coredns/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"

	"github.com/miekg/dns"
)

func init() { plugin.Register("deleg", setup) }

func setup(c *caddy.Controller) error {
	zones, responses, err := delegParse(c)

	if err != nil {
		return plugin.Error("deleg", err)
	}
	d := Deleg{
		zones:     zones,
		responses: responses,
	}

	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		d.Next = next
		return d
	})

	return nil
}

func delegParse(c *caddy.Controller) ([]string, []dns.RR, error) {
	zones := []string{}
	responses := []dns.RR{}

	i := 0
	// `deleg`
	for c.Next() {
		if i > 0 {
			return nil, nil, plugin.ErrOnce
		}
		i++

		zones = plugin.OriginsFromArgsOrServerBlock(c.RemainingArgs(), c.ServerBlockKeys)

		for c.NextBlock() {
			switch x := c.Val(); x {
			case "responses":
				r, e := responseParse(c)
				if e != nil {
					return nil, nil, e
				}
				responses = append(responses, r...)

			default:
				return nil, nil, c.Errf("unknown property '%s'", x)
			}
		}

	}

	return zones, responses, nil
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
