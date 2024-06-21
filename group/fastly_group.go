package group

import (
	"github.com/miekg/dns"
	"github.com/xsmartdns/xsmartdns/chain"
	"github.com/xsmartdns/xsmartdns/chain/chains"
	"github.com/xsmartdns/xsmartdns/chain/chains/cachechain"
	"github.com/xsmartdns/xsmartdns/config"
)

// dns upstream group
// one dns request will invoke all outbound in the group
type fastlyGroupInvoker struct {
	handleInvoke chain.HandleInvoke
}

func NewFastlyGroupInvoker(cfg *config.Group) GroupInvoker {
	handleInvoke := chain.BuildChain(
		cachechain.NewCacheChain(cfg),
		chains.NewSpeedSortChain(cfg),
		chains.NewRemoveruplicateChain(cfg),
		chains.NewResloveCnameChain(cfg),
		chains.NewRequestSettingChain(),
		chains.NewInvokeOutboundChain(cfg),
	)
	return &fastlyGroupInvoker{handleInvoke: handleInvoke}
}

func (p *fastlyGroupInvoker) Invoke(r *dns.Msg) (*dns.Msg, error) {
	return p.handleInvoke(r)
}
