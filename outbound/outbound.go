package outbound

import "github.com/miekg/dns"

type Outbound interface {
	Invoke(r *dns.Msg) (*dns.Msg, error)
}
