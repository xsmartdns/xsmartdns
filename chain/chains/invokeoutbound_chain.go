package chains

import (
	"fmt"
	"sync"

	"github.com/miekg/dns"
	"github.com/xsmartdns/xsmartdns/chain"
	"github.com/xsmartdns/xsmartdns/config"
	"github.com/xsmartdns/xsmartdns/log"
	"github.com/xsmartdns/xsmartdns/outbound"
	"github.com/xsmartdns/xsmartdns/util"
)

type invokeOutboundChain struct {
	cfg       *config.Group
	outbounds []outbound.Outbound
}

func NewInvokeOutboundChain(cfg *config.Group) chain.Chain {
	outbounds := make([]outbound.Outbound, 0, len(cfg.Outbounds))
	for _, c := range cfg.Outbounds {
		outbounds = append(outbounds, initOutbound(c))
	}
	return &invokeOutboundChain{cfg: cfg, outbounds: outbounds}
}

func (c *invokeOutboundChain) HandleRequest(r *dns.Msg, nextChain chain.HandleInvoke) (*dns.Msg, error) {
	wg := sync.WaitGroup{}
	ch := make(chan *invokeResp, len(c.outbounds))
	for i, o := range c.outbounds {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			resp, err := o.Invoke(r)
			ch <- &invokeResp{
				outboundIdx: idx,
				resp:        resp,
				err:         err,
			}
		}(i)
	}

	switch c.cfg.CacheMissResponseMode {
	case config.FIRST_PING_RESPONSEMODE:
		firstResp := <-ch
		if firstResp.err != nil {
			return nil, fmt.Errorf("[FIRST_PING_RESPONSEMODE]invoke outbound[%d] error:%v", firstResp.outboundIdx, firstResp.err)
		}
		// must last chain
		return firstResp.resp, nil
	case config.FASTEST_IP_RESPONSEMODE:
		wg.Wait()
		close(ch)
		msgs := make([]*dns.Msg, 0, len(c.outbounds))
		for resp := range ch {
			if resp.err != nil {
				log.Warnf("[FASTEST_IP_RESPONSEMODE]invoke outbound[%d] error:%v", resp.outboundIdx, resp.err)
				continue
			}
			msgs = append(msgs, resp.resp)
		}
		// not any resp succeed
		if len(msgs) == 0 {
			return nil, fmt.Errorf("invoke outbounds all failed")
		}
		// merge all msgs
		if len(msgs) > 1 {
			util.MergeAllAnswer(msgs[0], msgs[1:]...)
		}
		// must last chain
		return msgs[0], nil
	case config.FASTEST_RESPONSE_RESPONSEMODE:
		firstResp := <-ch
		if firstResp.err != nil {
			return nil, fmt.Errorf("[FASTEST_RESPONSE_RESPONSEMODE]invoke outbound[%d] error:%v", firstResp.outboundIdx, firstResp.err)
		}
		// must last chain
		return firstResp.resp, nil
	default:
		return nil, fmt.Errorf("unkown cacheMissResponseMode:%s", c.cfg.CacheMissResponseMode)
	}
}

type invokeResp struct {
	outboundIdx int
	resp        *dns.Msg
	err         error
}

func initOutbound(c *config.Outbound) outbound.Outbound {
	switch c.Protocol {
	case config.DNS_PROTOCOL:
		return outbound.NewDnsOutbound(c.DnsSetting)
	case config.SOCK5_PROTOCOL:
		return outbound.NewSock5Outbound(c.Sock5Setting)
	case config.HTTPS_PROTOCOL:
		return outbound.NewHttpsOutbound(c.HttpsSetting)
	default:
		panic("unkown protocol:" + c.Protocol)
	}
}
