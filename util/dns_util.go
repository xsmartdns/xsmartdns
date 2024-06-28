package util

import (
	"fmt"
	"math"
	"strings"

	"github.com/miekg/dns"
)

// get dns message question
// Question holds a DNS question. Usually there is just one. While the
// original DNS RFCs allow multiple questions in the question section of a
// message, in practice it never works. Because most DNS servers see multiple
// questions as an error, it is recommended to only have one question per message.
func GetQuestion(m *dns.Msg) (*dns.Question, error) {
	if len(m.Question) != 1 {
		return nil, fmt.Errorf("question number:%d is illegal", len(m.Question))
	}
	return &m.Question[0], nil
}

func GetHost(m *dns.Msg) (string, error) {
	q, err := GetQuestion(m)
	if err != nil {
		return "", err
	}
	return strings.TrimRight(q.Name, "."), nil
}

// Get the unique key of a dns question
func GetQuestionKey(q *dns.Question) string {
	return q.String()
}

// Get the unique key of a dns question
func GetQuestionRR(rr dns.RR) string {
	header := rr.Header()
	return dns.TypeToString[header.Rrtype] + "\t" + rr.String()[len(header.String()):]
}

// Remove duplicate rr
func RemoveDuplicateRR(data []dns.RR) []dns.RR {
	if len(data) <= 1 {
		return data
	}
	visited := make(map[string]struct{}, len(data))
	ret := make([]dns.RR, 0, len(data))
	for _, rr := range data {
		key := GetQuestionRR(rr)
		_, exist := visited[key]
		if !exist {
			ret = append(ret, rr)
			visited[key] = struct{}{}
		}
	}
	return ret
}

// merge from other msg answer
func MergeAllAnswer(m *dns.Msg, other ...*dns.Msg) *dns.Msg {
	answer := make([]dns.RR, 0, len(m.Answer))
	ns := make([]dns.RR, 0, len(m.Ns))
	extra := make([]dns.RR, 0, len(m.Ns))
	for _, msg := range other {
		answer = append(answer, msg.Answer...)
		ns = append(ns, msg.Ns...)
		extra = append(extra, msg.Extra...)
	}
	return m
}

// check ip type for dns.RR
func IsIpsRR(rr dns.RR) bool {
	switch rr.Header().Rrtype {
	case dns.TypeA, dns.TypeAAAA:
		return true
	}
	return false
}

// filter all ip rr
func FilterIpRR(data []dns.RR) (ipRRs, other []dns.RR) {
	ipRRs = make([]dns.RR, 0, len(data))
	other = make([]dns.RR, 0)
	for _, rr := range data {
		if IsIpsRR(rr) {
			ipRRs = append(ipRRs, rr)
		} else {
			other = append(other, rr)
		}
	}
	return
}

func LenLimit(data []dns.RR, limit int64) []dns.RR {
	if len(data) > int(limit) {
		return data[:limit]
	}
	return data
}

// rewrite rr ttl
func RewriteRRTTL(data []dns.RR, ttl uint32) {
	for _, msg := range data {
		msg.Header().Ttl = ttl
	}
}

// rewrite msg ttl
func RewriteMsgTTL(msg *dns.Msg, ttl uint32) *dns.Msg {
	RewriteRRTTL(msg.Answer, ttl)
	RewriteRRTTL(msg.Ns, ttl)
	RewriteRRTTL(msg.Extra, ttl)
	return msg
}

// get the max ttl in answer
func GetAnswerTTL(msg *dns.Msg) uint32 {
	ttl := uint32(math.MaxUint32)
	for _, rr := range msg.Answer {
		if rr.Header().Ttl < ttl {
			ttl = rr.Header().Ttl
		}
	}
	return ttl
}
