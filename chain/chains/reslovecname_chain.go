package chains

import (
	"github.com/miekg/dns"
	"github.com/xsmartdns/xsmartdns/chain"
	"github.com/xsmartdns/xsmartdns/config"
)

type resloveCnameChain struct {
}

func NewResloveCnameChain(cfg *config.Group) chain.Chain {
	return &resloveCnameChain{}
}

func (c *resloveCnameChain) HandleRequest(r *dns.Msg, nextChain chain.HandleInvoke) (*dns.Msg, error) {
	resp, err := nextChain(r)
	if err != nil {
		return nil, err
	}
	// TODO: reslove cname to A/AAAA
	for _, answer := range r.Answer {
		switch answer.Header().Rrtype {
		case dns.TypeCNAME:
			// reslove cname domain
		}
	}
	return resp, nil
}

func (c *resloveCnameChain) Shutdown() {
}
