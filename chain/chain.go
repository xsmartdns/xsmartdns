package chain

import (
	"github.com/miekg/dns"
	"github.com/xsmartdns/xsmartdns/model"
)

type Chain interface {
	// process msg
	HandleRequest(r *model.Message, nextChain HandleInvoke) (*dns.Msg, error)
	// Shutdown
	Shutdown()
}

type HandleInvoke func(*model.Message) (*dns.Msg, error)
