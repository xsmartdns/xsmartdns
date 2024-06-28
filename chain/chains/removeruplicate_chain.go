package chains

import (
	"github.com/miekg/dns"
	"github.com/xsmartdns/xsmartdns/chain"
	"github.com/xsmartdns/xsmartdns/config"
	"github.com/xsmartdns/xsmartdns/model"
	"github.com/xsmartdns/xsmartdns/util"
)

type removeruplicateChain struct {
	cfg *config.Group
}

func NewRemoveruplicateChain(cfg *config.Group) chain.Chain {
	return &removeruplicateChain{cfg: cfg}
}

func (c *removeruplicateChain) HandleRequest(r *model.Message, nextChain chain.HandleInvoke) (*dns.Msg, error) {
	resp, err := nextChain(r)
	if err != nil {
		return nil, err
	}
	resp.Answer = util.RemoveDuplicateRR(resp.Answer)
	if c.cfg.MaxIpsNumber != nil {
		resp.Answer = util.LenLimit(resp.Answer, *c.cfg.MaxIpsNumber)
	}
	resp.Ns = util.RemoveDuplicateRR(resp.Ns)
	resp.Extra = util.RemoveDuplicateRR(resp.Extra)
	return resp, nil
}

func (c *removeruplicateChain) Shutdown() {
}
