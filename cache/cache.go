package cache

import (
	"sync"

	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/miekg/dns"
	"github.com/xsmartdns/xsmartdns/cache/updateinvoke"
	"github.com/xsmartdns/xsmartdns/config"
	"github.com/xsmartdns/xsmartdns/log"
	"github.com/xsmartdns/xsmartdns/util"
)

type DnsQueryCache struct {
	sync.Mutex
	cfg          *config.CacheConfig
	cache        *lru.Cache[string, *CacheEntry]
	updateinvoke *updateinvoke.UpdateInvoker
}

func NewDnsQueryCache(cfg *config.Group) (*DnsQueryCache, error) {
	dc := &DnsQueryCache{cfg: cfg.CacheConfig}
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
