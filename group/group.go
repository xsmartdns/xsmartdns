package group

import "github.com/miekg/dns"

type GroupInvoker interface {
	Invoke(*dns.Msg) (*dns.Msg, error)
	Shutdown()
}
