package cache

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/miekg/dns"
	"github.com/sirupsen/logrus"
	"github.com/xsmartdns/xsmartdns/cache/updateinvoke"
	"github.com/xsmartdns/xsmartdns/config"
	"github.com/xsmartdns/xsmartdns/log"
	"github.com/xsmartdns/xsmartdns/model"
	"github.com/xsmartdns/xsmartdns/util"
	"github.com/xsmartdns/xsmartdns/util/timeutil"
)

const (
	MIN_UPDATE_DELAY_SECOND = 15
	MIN_UPDATE_DELAY        = MIN_UPDATE_DELAY_SECOND * time.Second

	CACHE_SPEED_CHECK_TIMES = 10
)

type CacheEntry struct {
	// no update
	cfg             *config.CacheConfig
	storeTimeSecond int64
	updateinvoke    *updateinvoke.UpdateInvoker
	request         *dns.Msg
	host            string

	// update by RWMutex
	sync.RWMutex
	resp             *dns.Msg
	ttl              uint32
	updateTimeSecond int64
	ti               *time.Timer

	// update by atomic
	vistiedTimeSecond int64
	cleared           int32
	updating          int32
}

func newCacheEntry(request, resp *dns.Msg, updateinvoke *updateinvoke.UpdateInvoker, cfg *config.CacheConfig) *CacheEntry {
	host, _ := util.GetHost(resp)
	now := timeutil.NowSecond()
	e := &CacheEntry{
		cfg:               cfg,
		request:           request,
		host:              host,
		resp:              resp.Copy(),
		ttl:               util.GetAnswerTTL(resp),
		storeTimeSecond:   now,
		updateTimeSecond:  now,
		vistiedTimeSecond: now,
		updateinvoke:      updateinvoke,
	}
	if cfg.PrefetchDomain {
		e.startUpdate(func() time.Duration {
			e.RLock()
			ttl := e.ttl
			e.RUnlock()
			afterUpdateTime := (time.Duration(ttl) - MIN_UPDATE_DELAY_SECOND) * time.Second
			return afterUpdateTime
		})
	} else if !cfg.DisableCacheExpired {
		e.startUpdate(func() time.Duration {
			return time.Duration(*cfg.CacheExpiredPrefetchTimeSecond) * time.Second
		})
	}
	go e.updateResp()
	return e
}

// get cache resp, if return nil the cache will be delete
func (e *CacheEntry) getResp() *dns.Msg {
	e.RLock()
	resp := e.resp.Copy()
	ttl := e.ttl
	updateTimeSecond := e.updateTimeSecond
	e.RUnlock()
	atomic.StoreInt64(&e.vistiedTimeSecond, timeutil.NowSecond())
	// rewrite ttl
	nowTtl := int64(ttl) - (timeutil.NowSecond() - updateTimeSecond)
	if nowTtl < MIN_UPDATE_DELAY_SECOND {
		nowTtl = MIN_UPDATE_DELAY_SECOND
	}
	if e.vistiedExpired() {
		return nil
	}
	if e.ttlExpired() {
		if e.cfg.DisableCacheExpired {
			return nil
		}
		// Accessing the expired cache returns ttl of 3
		if !e.cfg.PrefetchDomain {
			go e.updateResp()
		}
		nowTtl = *e.cfg.CacheExpiredReplyTtl
	}
	return util.RewriteMsgTTL(resp, uint32(nowTtl))
}

func (e *CacheEntry) ttlExpired() bool {
	e.RLock()
	updateTimeSecond := e.updateTimeSecond
	ttl := e.ttl
	e.RUnlock()
	return timeutil.NowSecond()-updateTimeSecond > int64(ttl)
}

func (e *CacheEntry) vistiedExpired() bool {
	if e.cfg.CacheExpiredTimeout <= 0 {
		return false
	}
	e.RLock()
	vistiedTimeSecond := e.vistiedTimeSecond
	e.RUnlock()
	return timeutil.NowSecond()-vistiedTimeSecond > e.cfg.CacheExpiredTimeout
}

func (e *CacheEntry) clear() {
	atomic.StoreInt32(&e.cleared, 1)
	e.RLock()
	ti := e.ti
	e.RUnlock()
	if ti != nil {
		ti.Stop()
	}
	if log.IsLevelEnabled(logrus.DebugLevel) {
		log.Debuf("[%s]cleaned", e.host)
	}
}

func (e *CacheEntry) stoped() bool {
	return atomic.LoadInt32(&e.cleared) == 1
}

func (e *CacheEntry) startUpdate(getAfterTime func() time.Duration) {
	if e.stoped() {
		return
	}

	// next update
	afterUpdateTime := util.Max(getAfterTime(), MIN_UPDATE_DELAY)
	if log.IsLevelEnabled(logrus.DebugLevel) {
		log.Debuf("[%s]next update after:%s", e.host, afterUpdateTime)
	}
	e.Lock()
	e.ti = time.AfterFunc(afterUpdateTime, func() {
		if e.stoped() {
			return
		}
		e.updateResp()
		e.startUpdate(getAfterTime)
	})
	e.Unlock()
}

func (e *CacheEntry) updateResp() error {
	if !atomic.CompareAndSwapInt32(&e.updating, 0, 1) {
		if log.IsLevelEnabled(logrus.DebugLevel) {
			log.Debuf("[%s] is updating", e.host)
		}
		return nil
	}
	defer atomic.StoreInt32(&e.updating, 0)

	req := model.WrapDnsMsg(e.request)
	// not first update and multiPrefetchSpeedCheck enabled
	if e.cfg.MultiPrefetchSpeedCheck && e.storeTimeSecond != e.updateTimeSecond {
		req.InvokeConfig.SpeedCheckTimes = CACHE_SPEED_CHECK_TIMES
	}
	resp, err := e.updateinvoke.Invoke(req)
	if err != nil {
		return fmt.Errorf("update invoke req:%s error:%v", e.request, err)
	}
	if log.IsLevelEnabled(logrus.DebugLevel) {
		host, _ := util.GetHost(resp)
		log.Debuf("[%s] updated:%s", host, resp.String())
	}

	e.Lock()
	defer e.Unlock()
	e.resp = resp
	e.ttl = util.GetAnswerTTL(resp)
	e.updateTimeSecond = timeutil.NowSecond()
	return nil
}
