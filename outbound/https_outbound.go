package outbound

import (
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/miekg/dns"
	"github.com/xsmartdns/xsmartdns/config"
)

type httpsOutbound struct {
	cfg          *config.HttpsSetting
	client       http.Client
	upstreamAddr string
}

func NewHttpsOutbound(cfg *config.HttpsSetting) Outbound {
	dialer := &net.Dialer{
		Timeout: 5 * time.Second,
	}
	return &httpsOutbound{
		cfg: cfg,
		client: http.Client{
			Transport: &http.Transport{
				Proxy:                 http.ProxyFromEnvironment,
				DialContext:           dialer.DialContext,
				ForceAttemptHTTP2:     true,
				MaxIdleConns:          100,
				IdleConnTimeout:       90 * time.Second,
				TLSHandshakeTimeout:   5 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
			},
			Timeout: 5 * time.Second,
		},
		upstreamAddr: cfg.Addr,
	}
}

func (o *httpsOutbound) Invoke(r *dns.Msg) (*dns.Msg, error) {
	msgBytes, err := r.Pack()
	if err != nil {
		log.Fatalf("Failed to pack DNS message: %v", err)
	}
	// to base64
	encodedMsg := base64.RawURLEncoding.EncodeToString(msgBytes)

	// create request
	dohURL := fmt.Sprintf(o.cfg.Addr+"?dns=%s", encodedMsg)
	req, err := http.NewRequest(http.MethodGet, dohURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request error:%v", err)
	}
	req.Header.Set("Accept", "application/dns-message")

	// send request
	resp, err := o.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send DoH request: %v error:%v", resp, err)
	}
	defer resp.Body.Close()

	// response
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to send DoH request: %v code:%d", resp, resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read HTTP response body error:%v", err)
	}

	// parse
	responseMsg := new(dns.Msg)
	err = responseMsg.Unpack(body)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack DNS response message error:%v", err)
	}
	return responseMsg, nil
}
