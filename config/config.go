package config

import (
	"encoding/json"
)

type Config struct {
	Inbounds []*Inbound `json:"inbounds"`
	Groups   []*Group   `json:"groups"`
	Routing  []*Rule    `json:"routing"`
	Log      Log        `json:"log"`
}

type Inbound struct {
	// only "dns", "https" and more reserved, default is dns
	Protocol Protocol `json:"protocol"`
	// Address to listen on, ":dns" if empty.
	Listen string `json:"listen"`
	// if "tcp" or "tcp-tls" (DNS over TLS) it will listen as TCP, default an UDP one
	Net Net `json:"net"`
	// if use "tcp-tls" Net or "https" protocol, should set tls cert and tls key
	TlsCert string `json:"tls_cert"`
	// if use "tcp-tls" Net or "https" protocol, should set tls cert and tls key
	TlsKey string `json:"tls_key"`
}

type Group struct {
	// the tag index of Group, default is "default"
	Tag       string      `json:"tag"`
	Outbounds []*Outbound `json:"outbounds"`
	// like smartdns response-mode. default is first-ping
	CacheMissResponseMode CacheMissResponseMode `json:"cacheMissResponseMode"`
	// speedChecks like xmartdns speed-check-mode, default is ping,tcp:80,tcp:443,udp:443
	// execute in order until success
	SpeedChecks []*SpeedCheckConfig `json:"speedChecks"`
	// max number of ips to answer,default no limit
	MaxIpsNumber *int64 `json:"maxIpsNumber"`
	// cache config
	CacheConfig *CacheConfig `json:"cache"`
	// fisrt response order / rt test response
}

type SpeedCheckConfig struct {
	SpeedCheckType SpeedCheckType `json:"speedCheckType"`
	Port           int64          `json:"port"`
}

type CacheConfig struct {
	CacheSize *int32 `json:"cacheSize"`
	// domain prefetch feature default false
	PrefetchDomain *bool `json:"prefetchDomain"`
	// cache expired feature, default true
	CacheExpired *bool `json:"cacheExpired"`
	// Cache expired timeout , default 0 no timeout
	CacheExpiredTimeout *int64 `json:"cacheExpiredTimeout"`
	// TTL value to use when replying with expired data, default 5(15 when prefetchDomain enabled)
	CacheExpiredReplyTtl *int64 `json:"cacheExpiredReplyTtl"`
	// Prefetch time when serve expired, default 28800
	CacheExpiredPrefetchTimeSecond *int64 `json:"cacheExpiredPrefetchTimeSecond"`
}

type Outbound struct {
	// if "dns" or "sock5" or "https", and more reserved, default is dns
	Protocol Protocol `json:"protocol"`
	// changed by protocol
	Setting json.RawMessage `json:"setting"`

	DnsSetting   *DnsSetting   `json:"-"`
	Sock5Setting *Sock5Setting `json:"-"`
	HttpsSetting *HttpsSetting `json:"-"`
}

type DnsSetting struct {
	// Address to outbound server, eg: 8.8.8.8, 1.1.1.1:53
	Addr string `json:"addr"`
	// if "tcp" or "tcp-tls" (DNS over TLS) it will listen as TCP, default an udp one
	Net Net `json:"net"`
	// insecure_skip_verify ,default is false
	InsecureSkipVerify bool `json:"insecure_skip_verify"`
}

type Sock5Setting struct {
	// sock5://127.0.0.1:1070
	Addr string `json:"addr"`
}

func (c *Sock5Setting) FillDefault() {
}
func (c *Sock5Setting) Verify() error {
	return nil
}

type HttpsSetting struct {
	// https://doh.pub/dns-query
	Addr               string `json:"addr"`
	InsecureSkipVerify bool   `json:"insecure_skip_verify"`
}

func (c *HttpsSetting) FillDefault() {
}
func (c *HttpsSetting) Verify() error {
	return nil
}

type Rule struct {
	// dns query domain filter, eg: geosite:cn, *.taobao.com, www.taobao.com
	Domain []string `json:"domain"`
	// forward traffic to group
	GroupTag string `json:"groupTag"`
}

type Log struct {
	// log level: debug,info,warn,error,panic
	Level string `json:"level"`
	// log path
	Filename string `json:"filename"`
}
