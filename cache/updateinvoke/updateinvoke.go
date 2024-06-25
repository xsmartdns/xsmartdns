package updateinvoke

import (
	"github.com/miekg/dns"
	"github.com/xsmartdns/xsmartdns/chain"
	"github.com/xsmartdns/xsmartdns/chain/chains"
	"github.com/xsmartdns/xsmartdns/config"
	"github.com/xsmartdns/xsmartdns/util"
)

type UpdateInvoker struct {
	handleInvoke chain.HandleInvoke
	chains       []chain.Chain
}

func NewUpdateInvoker(cfg *config.Group) (*UpdateInvoker, error) {
	cfg, err := util.Copy(cfg)
	if err != nil {
		return nil, err
	}
	cfg.FillDefault()
	if err := cfg.Verify(); err != nil {
		return nil, err
	}
	// force use fastest-ip in cache async update
	cfg.CacheMissResponseMode = config.FASTEST_IP_RESPONSEMODE

	handleInvoke, chains := chain.BuildChain(
		chains.NewSpeedSortChain(cfg),
		chains.NewRemoveruplicateChain(cfg),
		chains.NewResloveCnameChain(cfg),
		chains.NewRequestSettingChain(),
		chains.NewInvokeOutboundChain(cfg),
	)

	return &UpdateInvoker{handleInvoke: handleInvoke, chains: chains}, nil
}

func (i *UpdateInvoker) Invoke(r *dns.Msg) (*dns.Msg, error) {
	return i.handleInvoke(r)
}

func (i *UpdateInvoker) Shutdown() {
	for _, c := range i.chains {
		c.Shutdown()
	}
}
