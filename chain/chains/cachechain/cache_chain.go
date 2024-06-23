package cachechain

import (
	"fmt"

	"github.com/miekg/dns"
	"github.com/xsmartdns/xsmartdns/cache"
	"github.com/xsmartdns/xsmartdns/chain"
	"github.com/xsmartdns/xsmartdns/config"
	"github.com/xsmartdns/xsmartdns/util"
)

type cacheChain struct {
	cfg *config.Group

	cache *cache.DnsQueryCache
}

func NewCacheChain(cfg *config.Group) chain.Chain {
	dc, err := cache.NewDnsQueryCache(cfg)
	if err != nil {
		panic(fmt.Sprintf("create DnsQueryCache error:%v", err))
	}
	return &cacheChain{cfg: cfg, cache: dc}
}

func (c *cacheChain) HandleRequest(r *dns.Msg, nextChain chain.HandleInvoke) (*dns.Msg, error) {
	// find cache
	resp := c.cache.FindCacheResp(r)
	if resp != nil {
		resp.Id = r.Id
		return resp, nil
	}

	// miss cache
	resp, err := nextChain(r)
	if err != nil {
		return nil, err
	}
	c.cache.StoreCache(r, resp)
	util.RewriteMsgTTL(resp, 3)
	return resp, nil
}
