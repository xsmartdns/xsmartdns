package chains

import (
	"github.com/miekg/dns"
	"github.com/xsmartdns/xsmartdns/chain"
	"github.com/xsmartdns/xsmartdns/config"
	"github.com/xsmartdns/xsmartdns/model"
)

type resloveCnameChain struct {
}

func NewResloveCnameChain(cfg *config.Group) chain.Chain {
	return &resloveCnameChain{}
}

func (c *resloveCnameChain) HandleRequest(r *model.Message, nextChain chain.HandleInvoke) (*dns.Msg, error) {
	resp, err := nextChain(r)
	if err != nil {
		return nil, err
	}
	haveIp := false
	others := make([]dns.RR, 0, len(r.Answer))
	for _, answer := range resp.Answer {
		switch answer.Header().Rrtype {
		case dns.TypeCNAME:
			continue
		case dns.TypeA, dns.TypeAAAA:
			haveIp = true
			fallthrough
		default:
			others = append(others, answer)
		}
	}
	// remove all cname answers if have any ip answer
	if haveIp {
		resp.Answer = others
	}
	return resp, nil
}

func (c *resloveCnameChain) Shutdown() {
}
