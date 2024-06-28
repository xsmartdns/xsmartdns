package model

import "github.com/miekg/dns"

type Message struct {
	*dns.Msg

	InvokeConfig *InvokeConfig
}

type InvokeConfig struct {
	// Number of consecutive speed tests,user can not change, default 1 (the value is 10 when cache update)
	SpeedCheckTimes int32
}

func WrapDnsMsg(m *dns.Msg) *Message {
	invokeConfig := &InvokeConfig{}
	invokeConfig.SpeedCheckTimes = 1
	return &Message{Msg: m, InvokeConfig: invokeConfig}
}
