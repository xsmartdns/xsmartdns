package speedcheck

import (
	"context"
	"fmt"
	"math"
	"net"
	"strconv"
	"time"

	"github.com/miekg/dns"
	probing "github.com/prometheus-community/pro-bing"
	"github.com/xsmartdns/xsmartdns/config"
	"github.com/xsmartdns/xsmartdns/log"
	"github.com/xsmartdns/xsmartdns/util"
)

const (
	PING_TIMEOUT         = 2 * time.Second
	SPEED_CHECK_INTERVAL = 200 * time.Millisecond
)

type sppedTestResault struct {
	rtMs int64
	err  error
}

// do speed check
func SpeedCheckSync(ctx context.Context, rr dns.RR, cfgs []*config.SpeedCheckConfig) (rtMs int64) {
	if len(cfgs) == 0 {
		return -1
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	ch := make(chan *sppedTestResault, len(cfgs))
	timers := make([]*time.Timer, 0, len(cfgs))
	for i, speedConfig := range cfgs {
		ti := time.AfterFunc(time.Duration(i)*SPEED_CHECK_INTERVAL, func() {
			select {
			case <-ctx.Done():
				return
			default:
				rtMs, err := speedCheckSyncOne(ctx, rr, speedConfig)
				ch <- &sppedTestResault{rtMs: rtMs, err: err}
				if err == nil {
					log.Infof("speed %s:%d check:[%s] %dms", speedConfig.SpeedCheckType, speedConfig.Port, rr.String(), rtMs)
				} else {
					log.Infof("speed %s:%d check:[%s] error:%v", speedConfig.SpeedCheckType, speedConfig.Port, rr.String(), err)
				}
			}
		})
		timers = append(timers, ti)
	}
	defer func() {
		for _, ti := range timers {
			ti.Stop()
		}
	}()
	rtMs = math.MaxInt64
	for resault := range ch {
		if resault.err == nil {
			return resault.rtMs
		}
	}
	return
}

func speedCheckSyncOne(ctx context.Context, rr dns.RR, cfg *config.SpeedCheckConfig) (rtMs int64, err error) {
	if !util.IsIpsRR(rr) {
		return 0, fmt.Errorf("rr type is not ips")
	}
	var ip string
	switch rr.Header().Rrtype {
	case dns.TypeA:
		ip = rr.(*dns.A).A.String()
	case dns.TypeAAAA:
		ip = rr.(*dns.AAAA).String()
	default:
		return 0, fmt.Errorf("fail ip type:%s to speed check", dns.Type(rr.Header().Rrtype).String())
	}

	switch cfg.SpeedCheckType {
	case config.PING_SPEED_CHECK_TYPE:
		return ping(ctx, ip)
	case config.TCP_SPEED_CHECK_TYPE:
		return tcping(ctx, ip+":"+strconv.FormatInt(cfg.Port, 10))
	}

	return math.MaxInt64, fmt.Errorf("unkonw type")
}

func ping(ctx context.Context, ip string) (rtMs int64, err error) {
	ctx, cancel := context.WithTimeout(ctx, PING_TIMEOUT)
	defer cancel()

	pinger, err := probing.NewPinger(ip)
	if err != nil {
		return 0, err
	}
	defer pinger.Stop()
	go func() {
		<-ctx.Done()
		pinger.Stop()
	}()
	pinger.Count = 1
	err = pinger.RunWithContext(ctx)
	if err != nil {
		return 0, err
	}
	stats := pinger.Statistics()

	return stats.MaxRtt.Milliseconds(), nil
}

func tcping(ctx context.Context, address string) (rtMs int64, err error) {
	ctx, cancel := context.WithTimeout(ctx, PING_TIMEOUT)
	defer cancel()

	// start time
	start := time.Now()
	dialer := net.Dialer{}

	conn, err := dialer.DialContext(ctx, "tcp", address)
	if err != nil {
		return 0, fmt.Errorf("tcping connection:%s error:%v", address, err)
	}
	defer conn.Close()

	// cost time
	elapsed := time.Since(start)

	return elapsed.Milliseconds(), nil
}
