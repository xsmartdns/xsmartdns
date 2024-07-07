package speedcheck

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"math"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/miekg/dns"
	probing "github.com/prometheus-community/pro-bing"
	"github.com/xsmartdns/xsmartdns/config"
	"github.com/xsmartdns/xsmartdns/log"
	"github.com/xsmartdns/xsmartdns/model"
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
func SpeedCheckSync(ctx context.Context, msg *model.Message, rr dns.RR, cfg *config.Group) (rtMs int64) {
	if len(cfg.SpeedChecks) == 0 {
		return -1
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	ch := make(chan *sppedTestResault, len(cfg.SpeedChecks))
	timers := make([]*time.Timer, 0, len(cfg.SpeedChecks))
	for i, speedConfig := range cfg.SpeedChecks {
		speedCheckTimes := msg.InvokeConfig.SpeedCheckTimes
		speedCheckInterval := SPEED_CHECK_INTERVAL
		if speedCheckTimes > 1 {
			speedCheckInterval += time.Duration(speedCheckTimes) * time.Second
		}
		ti := time.AfterFunc(time.Duration(i)*speedCheckInterval, func() {
			select {
			case <-ctx.Done():
				return
			default:
				rtMs, err := speedCheckSyncWithTimes(ctx, rr, speedConfig, speedCheckTimes)
				ch <- &sppedTestResault{rtMs: rtMs, err: err}
				if err == nil {
					log.Infof("speed %s:%d check:[%s] avg %dms", speedConfig.SpeedCheckType, speedConfig.Port, rr.String(), rtMs)
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

func speedCheckSyncWithTimes(ctx context.Context, rr dns.RR, cfg *config.SpeedCheckConfig, times int32) (rtMsAvg int64, err error) {
	sum := int64(0)
	succeedOnce := false
	for i := int32(0); i < times; i++ {
		// sleep 1s
		if i > 0 {
			timer := time.NewTimer(time.Second)
			select {
			case <-ctx.Done():
				timer.Stop()
				return math.MaxInt64, ctx.Err()
			case <-timer.C:
			}

		}

		var rtMs int64
		rtMs, err = speedCheckSyncOne(ctx, rr, cfg)
		if err != nil {
			rtMs = math.MaxInt32
			log.Infof("speed[%d] %s:%d check:[%s] error:%v", i, cfg.SpeedCheckType, cfg.Port, rr.String(), err)
		} else {
			succeedOnce = true
			log.Infof("speed[%d] %s:%d check:[%s] %dms", i, cfg.SpeedCheckType, cfg.Port, rr.String(), rtMs)
		}
		sum += rtMs
	}
	if !succeedOnce {
		return 0, err
	}
	return int64(math.Ceil(float64(sum) / float64(times))), nil
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
		ip = rr.(*dns.AAAA).AAAA.String()
	default:
		return 0, fmt.Errorf("fail ip type:%s to speed check", dns.Type(rr.Header().Rrtype).String())
	}

	switch cfg.SpeedCheckType {
	case config.PING_SPEED_CHECK_TYPE:
		return ping(ctx, ip)
	case config.HTTP_SPEED_CHECK_TYPE:
		return httping(ctx, ip+":"+strconv.FormatInt(cfg.Port, 10))
	case config.TCP_SPEED_CHECK_TYPE:
		return tcping(ctx, ip+":"+strconv.FormatInt(cfg.Port, 10))
	}

	return math.MaxInt32, fmt.Errorf("unkonw type")
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
	dialer := net.Dialer{
		Timeout:   PING_TIMEOUT,
		KeepAlive: -1 * time.Second,
	}

	conn, err := dialer.DialContext(ctx, "tcp", address)
	if err != nil {
		return 0, fmt.Errorf("tcping connection:%s error:%v", address, err)
	}
	defer conn.Close()

	// cost time
	elapsed := time.Since(start)

	return elapsed.Milliseconds(), nil
}

func httping(ctx context.Context, address string) (rtMs int64, err error) {
	ctx, cancel := context.WithTimeout(ctx, PING_TIMEOUT)
	defer cancel()

	dialer := &net.Dialer{
		Timeout:   PING_TIMEOUT,
		KeepAlive: -1 * time.Second,
	}
	client := http.Client{
		Transport: &http.Transport{
			Proxy:                 http.ProxyFromEnvironment,
			DialContext:           dialer.DialContext,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          1,
			IdleConnTimeout:       PING_TIMEOUT,
			TLSHandshakeTimeout:   PING_TIMEOUT,
			ExpectContinueTimeout: 1 * time.Second,
			TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
		},
		Timeout: PING_TIMEOUT,
	}

	// start time
	start := time.Now()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://"+address, nil)
	if err != nil {
		return 0, fmt.Errorf("httping create request:%s error:%v", address, err)
	}
	// req.Host = host
	resp, err := client.Do(req)
	if err != nil && !errors.Is(err, io.EOF) {
		return 0, fmt.Errorf("httping connection:%s error:%v", address, err)
	}
	if resp != nil {
		resp.Body.Close()
	}
	elapsed := time.Since(start)

	return elapsed.Milliseconds(), nil
}
