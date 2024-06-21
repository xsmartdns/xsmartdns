package outbound

import (
	"fmt"

	"github.com/miekg/dns"
	"github.com/xsmartdns/xsmartdns/config"
)

type sock5Outbound struct {
	upstreamAddr string
}

func NewSock5Outbound(cfg *config.Sock5Setting) Outbound {
	return &sock5Outbound{upstreamAddr: cfg.Addr}
}

func (o *sock5Outbound) Invoke(r *dns.Msg) (*dns.Msg, error) {
	// TODO: sock5，需要移动到 transport 模型？否则无法让递归解析生效
	return nil, fmt.Errorf("not support")
}
