package outbound

import (
	"github.com/miekg/dns"
	"github.com/xsmartdns/xsmartdns/model"
)

type Outbound interface {
	Invoke(r *model.Message) (*dns.Msg, error)
}
