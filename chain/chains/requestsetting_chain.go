package chains

import (
	"github.com/miekg/dns"
	"github.com/xsmartdns/xsmartdns/chain"
	"github.com/xsmartdns/xsmartdns/model"
)

type requestSettingChain struct {
}

func NewRequestSettingChain() chain.Chain {
	return &requestSettingChain{}
}

func (c *requestSettingChain) HandleRequest(r *model.Message, nextChain chain.HandleInvoke) (*dns.Msg, error) {
	// Enable recursive query
	r.RecursionDesired = true
	return nextChain(r)
}

func (c *requestSettingChain) Shutdown() {
}
