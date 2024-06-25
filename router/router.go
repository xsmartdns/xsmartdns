package router

import (
	"github.com/miekg/dns"
	"github.com/xsmartdns/xsmartdns/group"
)

// Router used to match and find group
type Router interface {
	FindGroupInvoker(*dns.Msg) (group.GroupInvoker, error)
	Shutdown()
}
