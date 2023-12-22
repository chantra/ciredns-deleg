package deleg

import (
	"github.com/coredns/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
)

func init() { plugin.Register("deleg", setup) }

func setup(c *caddy.Controller) error {
	d := Deleg{}

	c.Next() // `deleg`
	if c.NextArg() {
		return plugin.Error("deleg", c.ArgErr())
	}

	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		d.Next = next
		return d
	})

	return nil
}
