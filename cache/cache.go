package cache

import (
	"sync"
	"time"

	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/miekg/dns"
	"github.com/xsmartdns/xsmartdns/cache/updateinvoke"
	"github.com/xsmartdns/xsmartdns/config"
	"github.com/xsmartdns/xsmartdns/log"
	"github.com/xsmartdns/xsmartdns/util"
)

const (
	CLEAR_EXPIRED_CACHE_INTERVAL = 30 * time.Minute
)

type DnsQueryCache struct {
	sync.RWMutex
	cfg          *config.CacheConfig
	checkTimer   *time.Ticker
	cache        *lru.Cache[string, *CacheEntry]
	updateinvoke *updateinvoke.UpdateInvoker
}

func NewDnsQueryCache(cfg *config.Group) (*DnsQueryCache, error) {
	dc := &DnsQueryCache{cfg: cfg.CacheConfig, checkTimer: time.NewTicker(CLEAR_EXPIRED_CACHE_INTERVAL)}
	c, err := lru.NewWithEvict(int(*cfg.CacheConfig.CacheSize), dc.onEvicted)
	if err != nil {
		return nil, err
	}
	dc.cache = c
	updateinvoke, err := updateinvoke.NewUpdateInvoker(cfg)
	if err != nil {
		return nil, err
	}
	dc.updateinvoke = updateinvoke
	go dc.checkExpiredCacheLoop()
	return dc, nil
}

func (c *DnsQueryCache) FindCacheResp(r *dns.Msg) *dns.Msg {
	key, err := getKey(r)
	if err != nil {
		return nil
	}

	c.Lock()
	defer c.Unlock()
	if !c.cache.Contains(key) {
		return nil
	}
	value, ok := c.cache.Get(key)
	if !ok {
		return nil
	}
	resp := value.getResp()
	if resp == nil {
		c.cache.Remove(key)
	}
	return resp
}

func (c *DnsQueryCache) StoreCache(r, resp *dns.Msg) {
	key, err := getKey(r)
	if err != nil {
		return
	}
	c.Lock()
	defer c.Unlock()
	if c.cache.Contains(key) {
		return
	}
	c.cache.Add(key, newCacheEntry(r.Copy(), resp.Copy(), c.updateinvoke, c.cfg))
}

func (c *DnsQueryCache) Shutdown() {
	c.updateinvoke.Shutdown()
	c.checkTimer.Stop()
}

func (c *DnsQueryCache) checkExpiredCacheLoop() {
	for range c.checkTimer.C {
		c.clearExpiredCache()
	}
}

func (c *DnsQueryCache) clearExpiredCache() {
	// quick path
	c.RLock()
	_, value, ok := c.cache.GetOldest()
	if !ok || !value.vistiedExpired() {
		c.RUnlock()
		return
	}
	c.RUnlock()
	// slow path
	c.Lock()
	defer c.Unlock()
	for {
		key, value, ok := c.cache.GetOldest()
		if !ok || !value.vistiedExpired() {
			break
		}
		c.cache.Remove(key)
	}
}

func (c *DnsQueryCache) onEvicted(key string, value *CacheEntry) {
	value.clear()
}

func getKey(r *dns.Msg) (string, error) {
	// Make sure msg has been decompressed
	r.Len()
	question, err := util.GetQuestion(r)
	if err != nil {
		log.Errorf("cache get msg:%v question error:%v", r, err)
		return "", err
	}
	return util.GetQuestionKey(question), nil
}
