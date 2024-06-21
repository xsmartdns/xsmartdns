package chain

import (
	"github.com/miekg/dns"
)

func BuildChain(chains ...Chain) HandleInvoke {
	if len(chains) == 0 {
		return func(r *dns.Msg) (*dns.Msg, error) {
			return r, nil
		}
	}

	var next HandleInvoke = func(r *dns.Msg) (*dns.Msg, error) {
		return r, nil
	}

	for i := len(chains) - 1; i >= 0; i-- {
		currentChain := chains[i]
		nextChain := next
		next = func(r *dns.Msg) (*dns.Msg, error) {
			return currentChain.HandleRequest(r, nextChain)
		}
	}

	return next
}
