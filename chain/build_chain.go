package chain

import (
	"github.com/miekg/dns"
	"github.com/xsmartdns/xsmartdns/model"
)

func BuildChain(chains ...Chain) (HandleInvoke, []Chain) {
	if len(chains) == 0 {
		return func(r *model.Message) (*dns.Msg, error) {
			return r.Msg, nil
		}, chains
	}

	var next HandleInvoke = func(r *model.Message) (*dns.Msg, error) {
		return r.Msg, nil
	}

	for i := len(chains) - 1; i >= 0; i-- {
		currentChain := chains[i]
		nextChain := next
		next = func(r *model.Message) (*dns.Msg, error) {
			return currentChain.HandleRequest(r, nextChain)
		}
	}

	return next, chains
}
