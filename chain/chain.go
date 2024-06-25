package chain

import "github.com/miekg/dns"

type Chain interface {
	// process msg
	HandleRequest(r *dns.Msg, nextChain HandleInvoke) (*dns.Msg, error)
	// Shutdown
	Shutdown()
}

type HandleInvoke func(*dns.Msg) (*dns.Msg, error)
