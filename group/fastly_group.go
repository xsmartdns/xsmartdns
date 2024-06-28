package group

import (
	"github.com/miekg/dns"
	"github.com/xsmartdns/xsmartdns/chain"
	"github.com/xsmartdns/xsmartdns/chain/chains"
	"github.com/xsmartdns/xsmartdns/chain/chains/cachechain"
	"github.com/xsmartdns/xsmartdns/config"
	"github.com/xsmartdns/xsmartdns/model"
)

// dns upstream group
// one dns request will invoke all outbound in the group
type fastlyGroupInvoker struct {
	handleInvoke chain.HandleInvoke
	chains       []chain.Chain
}

func NewFastlyGroupInvoker(cfg *config.Group) GroupInvoker {
	handleInvoke, chains := chain.BuildChain(
		cachechain.NewCacheChain(cfg),
		chains.NewSpeedSortChain(cfg),
		chains.NewRemoveruplicateChain(cfg),
		chains.NewResloveCnameChain(cfg),
		chains.NewRequestSettingChain(),
		chains.NewInvokeOutboundChain(cfg),
	)
	return &fastlyGroupInvoker{handleInvoke: handleInvoke, chains: chains}
}

func (p *fastlyGroupInvoker) Invoke(r *dns.Msg) (*dns.Msg, error) {
	return p.handleInvoke(model.WrapDnsMsg(r))
}

func (p *fastlyGroupInvoker) Shutdown() {
	for _, c := range p.chains {
		c.Shutdown()
	}
}
