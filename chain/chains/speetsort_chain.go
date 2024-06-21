package chains

import (
	"context"
	"math"
	"sort"
	"sync"

	"github.com/miekg/dns"
	"github.com/xsmartdns/xsmartdns/chain"
	"github.com/xsmartdns/xsmartdns/config"
	"github.com/xsmartdns/xsmartdns/util"
	"github.com/xsmartdns/xsmartdns/util/speedcheck"
)

type speedSortChain struct {
	cfg *config.Group
}

func NewSpeedSortChain(cfg *config.Group) chain.Chain {
	return &speedSortChain{cfg: cfg}
}

func (c *speedSortChain) HandleRequest(r *dns.Msg, nextChain chain.HandleInvoke) (*dns.Msg, error) {
	// FASTEST_RESPONSE_RESPONSEMODE not should speedtest
	if c.cfg.CacheMissResponseMode == config.FASTEST_RESPONSE_RESPONSEMODE {
		return nextChain(r)
	}
	resp, err := nextChain(r)
	if err != nil {
		return resp, err
	}
	ipRRs, other := util.FilterIpRR(resp.Answer)
	// only check ip rr
	if len(ipRRs) == 0 {
		return resp, nil
	}

	// speed check
	resaults := c.speedCheck(ipRRs, c.cfg.CacheMissResponseMode == config.FIRST_PING_RESPONSEMODE)
	// sort rr by rt
	sort.Slice(resaults, func(i, j int) bool {
		return resaults[i].rtMs < resaults[j].rtMs
	})
	// TODO: 过滤掉耗时过大的ip
	for i := 0; i < len(resaults); i++ {
		ipRRs[i] = resaults[i].rr
	}

	resp.Answer = append(ipRRs, other...)
	return resp, nil
}

func (c *speedSortChain) speedCheck(ipRRs []dns.RR, onlyfirstResponse bool) []*sppedTestResault {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	wg := sync.WaitGroup{}
	ch := make(chan *sppedTestResault, len(ipRRs))
	for _, rr := range ipRRs {
		wg.Add(1)
		go func(rr dns.RR) {
			defer wg.Done()
			rtMs := speedcheck.SpeedCheckSync(ctx, rr, c.cfg.SpeedChecks)
			ch <- &sppedTestResault{
				rr:   rr,
				rtMs: rtMs,
			}
		}(rr)
	}
	resaults := make([]*sppedTestResault, 0, len(ipRRs))
	if onlyfirstResponse {
		firstResp := <-ch
		for _, rr := range ipRRs {
			rtMs := int64(math.MaxInt64)
			if firstResp.rr == rr {
				rtMs = firstResp.rtMs
			}
			resaults = append(resaults, &sppedTestResault{
				rr:   rr,
				rtMs: rtMs,
			})
		}
	} else {
		wg.Wait()
		close(ch)
		for ret := range ch {
			resaults = append(resaults, ret)
		}
	}

	return resaults
}

type sppedTestResault struct {
	rr   dns.RR
	rtMs int64
}
