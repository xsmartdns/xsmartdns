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
	"github.com/xsmartdns/xsmartdns/util"
	"github.com/xsmartdns/xsmartdns/util/timeutil"
)

const (
	MIN_UPDATE_DELAY = 30 * time.Second
)

type CacheEntry struct {
	// no update
	cfg          *config.CacheConfig
	updateinvoke *updateinvoke.UpdateInvoker
	request      *dns.Msg
	host         string

	// update by RWMutex
	sync.RWMutex
	resp             *dns.Msg
	ttl              uint32
	storeTimeSecond  int64
	updateTimeSecond int64
	ti               *time.Timer

	// update by atomic
	vistiedTimeSecond int64
	cleared           int32
	updating          int32
}

func newCacheEntry(request, resp *dns.Msg, updateinvoke *updateinvoke.UpdateInvoker, cfg *config.CacheConfig) *CacheEntry {
	host, _ := util.GetHost(resp)
	e := &CacheEntry{
		cfg:               cfg,
		request:           request,
		host:              host,
		resp:              resp.Copy(),
		ttl:               util.GetAnswerTTL(resp),
		storeTimeSecond:   timeutil.NowSecond(),
		updateTimeSecond:  timeutil.NowSecond(),
		vistiedTimeSecond: timeutil.NowSecond(),
		updateinvoke:      updateinvoke,
	}
	if *cfg.PrefetchDomain {
		e.startUpdate(func() time.Duration {
			e.RLock()
			ttl := e.ttl
			e.RUnlock()
			afterUpdateTime := (time.Duration(ttl) - 5) * time.Second
			return afterUpdateTime
		})
	} else if *cfg.CacheExpired {
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
	if nowTtl < 0 {
		nowTtl = 0
	}
	if e.vistiedExpired() {
		return nil
	}
	if e.ttlExpired() {
		if !*e.cfg.CacheExpired {
			return nil
		}
		// Accessing the expired cache returns ttl of 3
		if !*e.cfg.PrefetchDomain {
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
	if *e.cfg.CacheExpiredTimeout <= 0 {
		return false
	}
	e.RLock()
	vistiedTimeSecond := e.vistiedTimeSecond
	e.RUnlock()
	return timeutil.NowSecond()-vistiedTimeSecond > int64(*e.cfg.CacheExpiredTimeout)
}

func (e *CacheEntry) clear() {
	atomic.StoreInt32(&e.cleared, 1)
	e.RLock()
	ti := e.ti
	e.RUnlock()
	if ti != nil {
		ti.Stop()
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
	afterUpdateTime := getAfterTime()
	if afterUpdateTime < MIN_UPDATE_DELAY {
		afterUpdateTime = MIN_UPDATE_DELAY
	}
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

	resp, err := e.updateinvoke.Invoke(e.request)
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
