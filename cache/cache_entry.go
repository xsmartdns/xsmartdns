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
	MIN_UPDATE_DELAY = 5 * time.Second
)

type CacheEntry struct {
	// no update
	cfg          *config.CacheConfig
	updateinvoke *updateinvoke.UpdateInvoker
	request      *dns.Msg

	// update by RWMutex
	sync.RWMutex
	resp             *dns.Msg
	ttl              uint32
	storeTimeSecond  int64
	updateTimeSecond int64

	// update by atomic
	vistiedTimeSecond int64
	cleared           int32
}

func newCacheEntry(request, resp *dns.Msg, updateinvoke *updateinvoke.UpdateInvoker, cfg *config.CacheConfig) *CacheEntry {
	e := &CacheEntry{
		cfg:               cfg,
		request:           request,
		resp:              resp.Copy(),
		ttl:               util.GetAnswerTTL(resp),
		storeTimeSecond:   timeutil.NowSecond(),
		updateTimeSecond:  timeutil.NowSecond(),
		vistiedTimeSecond: timeutil.NowSecond(),
		updateinvoke:      updateinvoke,
	}
	if *cfg.PrefetchDomain {
		e.startUpdate()
	}
	return e
}

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
	if e.ttlExpired() {
		// 访问到过期缓存则返回 ttl 为 3
		nowTtl = 3
	}
	// TODO: 关闭过期缓存时，删除已过期的缓存
	return util.RewriteMsgTTL(resp, uint32(nowTtl))
}

func (e *CacheEntry) ttlExpired() bool {
	e.RLock()
	updateTimeSecond := e.updateTimeSecond
	ttl := e.ttl
	e.RUnlock()
	return timeutil.NowSecond()-updateTimeSecond > int64(ttl)
}

func (e *CacheEntry) clear() {
	atomic.StoreInt32(&e.cleared, 1)
}

func (e *CacheEntry) stoped() bool {
	return atomic.LoadInt32(&e.cleared) == 1
}

func (e *CacheEntry) startUpdate() {
	if e.stoped() {
		return
	}

	// next update
	e.RLock()
	ttl := e.ttl
	resp := e.resp
	e.RUnlock()
	var afterUpdateTime time.Duration
	if *e.cfg.CacheExpired {
		afterUpdateTime = time.Duration(ttl) * time.Second
	} else {
		afterUpdateTime = time.Duration(ttl-5) * time.Second
	}
	if afterUpdateTime < MIN_UPDATE_DELAY {
		afterUpdateTime = MIN_UPDATE_DELAY
	}

	if log.IsLevelEnabled(logrus.DebugLevel) {
		q, _ := util.GetQuestion(resp)
		log.Debuf("[%s]next update after:%s", q.Name, afterUpdateTime)
	}
	time.AfterFunc(afterUpdateTime, func() {
		if e.stoped() {
			return
		}
		e.updateResp()
		e.startUpdate()
	})
}

func (e *CacheEntry) updateResp() error {
	e.RLock()
	updateTimeSecond := e.updateTimeSecond
	e.RUnlock()
	if timeutil.NowSecond()-updateTimeSecond < int64(MIN_UPDATE_DELAY)/int64(time.Second) {
		return nil
	}

	resp, err := e.updateinvoke.Invoke(e.request)
	if err != nil {
		return fmt.Errorf("update invoke req:%s error:%v", e.request, err)
	}
	if log.IsLevelEnabled(logrus.DebugLevel) {
		q, _ := util.GetQuestion(resp)
		log.Debuf("[%s] updated:%s", q.Name, resp.String())
	}

	e.Lock()
	defer e.Unlock()
	e.resp = resp
	e.ttl = util.GetAnswerTTL(resp)
	e.updateTimeSecond = timeutil.NowSecond()
	return nil
}
