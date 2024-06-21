package config

import (
	"encoding/json"
	"fmt"
	"net/url"
)

// Config
func (c *Config) FillDefault() {
	for _, inbound := range c.Inbounds {
		inbound.FillDefault()
	}
	for _, group := range c.Groups {
		group.FillDefault()
	}
	for _, rule := range c.Routing {
		rule.FillDefault()
	}
}
func (c *Config) Verify() error {
	if len(c.Inbounds) == 0 {
		return fmt.Errorf("inbounds is empty")
	}
	if len(c.Groups) == 0 {
		return fmt.Errorf("groups is empty")
	}
	if len(c.Routing) == 0 && len(c.Groups) != 1 {
		return fmt.Errorf("routing can be empty only when the number of groups is 1")
	}
	for i, inbound := range c.Inbounds {
		if err := inbound.Verify(); err != nil {
			return fmt.Errorf("parse inbounds[%d] error:%v", i, err)
		}
	}
	for i, group := range c.Groups {
		if err := group.Verify(); err != nil {
			return fmt.Errorf("parse groups[%d] error:%v", i, err)
		}
	}
	for i, rule := range c.Routing {
		if err := rule.Verify(); err != nil {
			return fmt.Errorf("parse routing[%d] error:%v", i, err)
		}
	}
	return nil
}

// Inbound
func (c *Inbound) FillDefault() {
	if len(c.Protocol) == 0 {
		c.Protocol = DEFAULT_PROTOCOL
	}
	if len(c.Net) == 0 {
		c.Net = DEFAULT_NET
	}
}
func (c *Inbound) Verify() error {
	if c.Protocol != DNS_PROTOCOL {
		return fmt.Errorf("unknow protocol:%s", c.Protocol)
	}
	if len(c.Listen) == 0 {
		return fmt.Errorf("listen is empty")
	}
	switch c.Net {
	case UDP_NET:
	case TCP_NET:
	case TLS_NET:
	default:
		return fmt.Errorf("unknow net:%s", c.Net)
	}
	return nil
}

// Group
func (c *Group) FillDefault() {
	if len(c.Tag) == 0 {
		c.Tag = DEFAULT_TAG
	}
	if len(c.CacheMissResponseMode) == 0 {
		c.CacheMissResponseMode = FIRST_PING_RESPONSEMODE
	}
	for _, outbound := range c.Outbounds {
		outbound.FillDefault()
	}
	if c.CacheConfig == nil {
		c.CacheConfig = &CacheConfig{}
		c.CacheConfig.FillDefault()
	}
	if len(c.SpeedChecks) == 0 {
		c.SpeedChecks = append(c.SpeedChecks,
			&SpeedCheckConfig{SpeedCheckType: PING_SPEED_CHECK_TYPE},
			&SpeedCheckConfig{SpeedCheckType: TCP_SPEED_CHECK_TYPE, Port: 80},
			&SpeedCheckConfig{SpeedCheckType: TCP_SPEED_CHECK_TYPE, Port: 443},
			// &SpeedCheckConfig{SpeedCheckType: UDP_SPEED_CHECK_TYPE, Port: 443},
		)
	}
}
func (c *Group) Verify() error {
	if len(c.Outbounds) == 0 {
		return fmt.Errorf("outbounds is empty")
	}
	for i, outbound := range c.Outbounds {
		if err := outbound.Verify(); err != nil {
			return fmt.Errorf("parse outbounds[%d] error:%v", i, err)
		}
	}
	switch c.CacheMissResponseMode {
	case FIRST_PING_RESPONSEMODE:
	case FASTEST_IP_RESPONSEMODE:
	case FASTEST_RESPONSE_RESPONSEMODE:
	default:
		return fmt.Errorf("unkown cacheMissResponseMode:%s", c.CacheMissResponseMode)
	}

	for i, speedCheck := range c.SpeedChecks {
		if err := speedCheck.Verify(); err != nil {
			return fmt.Errorf("parse speedCheck[%d] error:%v", i, err)
		}
	}
	return nil
}

// SpeedCheckConfig
func (c *SpeedCheckConfig) Verify() error {
	switch c.SpeedCheckType {
	case PING_SPEED_CHECK_TYPE:
	case TCP_SPEED_CHECK_TYPE:
		if c.Port == 0 {
			return fmt.Errorf("port is empty")
		}
	default:
		return fmt.Errorf("unkown speedTestType:%s", c.SpeedCheckType)
	}
	return nil
}

// Outbound
func (c *Outbound) FillDefault() {
	if len(c.Protocol) == 0 {
		c.Protocol = DEFAULT_PROTOCOL
	}
}
func (c *Outbound) Verify() error {
	switch c.Protocol {
	case DNS_PROTOCOL:
		if err := json.Unmarshal(c.Setting, &c.DnsSetting); err != nil {
			return err
		}
		c.DnsSetting.FillDefault()
		if err := c.DnsSetting.Verify(); err != nil {
			return fmt.Errorf("DnsSetting verify error:%v", err)
		}
	case SOCK5_PROTOCOL:
		if err := json.Unmarshal(c.Setting, &c.Sock5Setting); err != nil {
			return err
		}
		c.Sock5Setting.FillDefault()
		if err := c.Sock5Setting.Verify(); err != nil {
			return fmt.Errorf("Sock5Setting verify error:%v", err)
		}
	case HTTPS_PROTOCOL:
		if err := json.Unmarshal(c.Setting, &c.HttpsSetting); err != nil {
			return err
		}
		c.HttpsSetting.FillDefault()
		if err := c.HttpsSetting.Verify(); err != nil {
			return fmt.Errorf("HttpsSetting verify error:%v", err)
		}
	default:
		return fmt.Errorf("unknow protocol:%s", c.Protocol)
	}
	return nil
}

// CacheConfig
func (c *CacheConfig) FillDefault() {
	if c.CacheSize == nil {
		tmp := int32(10240)
		// TODO: auto size
		c.CacheSize = &tmp
	}
	if c.PrefetchDomain == nil {
		tmp := true
		c.PrefetchDomain = &tmp
	}
	if c.CacheExpired == nil {
		tmp := true
		c.CacheExpired = &tmp
	}
}

// DnsSetting
func (c *DnsSetting) FillDefault() {
	if len(c.Net) == 0 {
		c.Net = UDP_NET
	}
	// 解析 URL
	parsedURL, _ := url.Parse(c.Addr)
	if parsedURL != nil {
		if len(parsedURL.Port()) == 0 {
			c.Addr = fmt.Sprintf("%s:53", parsedURL.Hostname())
		}
	}
}
func (c *DnsSetting) Verify() error {
	switch c.Net {
	case UDP_NET:
	case TCP_NET:
	case TLS_NET:
	default:
		return fmt.Errorf("unknow net:%s", c.Net)
	}
	return nil
}

// Rule
func (c *Rule) FillDefault() {
}
func (c *Rule) Verify() error {
	if len(c.Domain) == 0 {
		return fmt.Errorf("domain is empty")
	}
	if len(c.GroupTag) == 0 {
		return fmt.Errorf("groupTag is empty")
	}
	// TODO: check rule mapping
	return nil
}
