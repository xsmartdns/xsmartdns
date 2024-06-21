package outbound

import (
	"github.com/miekg/dns"
	"github.com/xsmartdns/xsmartdns/config"
)

type dnsOutbound struct {
	client       *dns.Client
	upstreamAddr string
}

func NewDnsOutbound(cfg *config.DnsSetting) Outbound {
	client := &dns.Client{}
	client.Net = string(cfg.Net)
	return &dnsOutbound{client: client, upstreamAddr: cfg.Addr}
}

func (o *dnsOutbound) Invoke(r *dns.Msg) (*dns.Msg, error) {
	resp, _, err := o.client.Exchange(r, o.upstreamAddr)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
