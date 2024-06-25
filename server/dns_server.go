package server

import (
	"github.com/miekg/dns"
	"github.com/xsmartdns/xsmartdns/config"
	"github.com/xsmartdns/xsmartdns/log"
	"github.com/xsmartdns/xsmartdns/router"
)

type dnsServer struct {
	cfg       config.Inbound
	dnsServer *dns.Server

	router router.Router
}

func NewDnsServer(cfg config.Inbound, router router.Router) Server {
	return &dnsServer{cfg: cfg, router: router}
}

func (srv *dnsServer) Init() error {
	// dns handler
	dns.HandleFunc(".", srv.handleDNSRequest)

	// create dns server
	srv.dnsServer = &dns.Server{Addr: srv.cfg.Listen, Net: string(srv.cfg.Net)}
	return nil
}

func (srv *dnsServer) Start() error {
	log.Infof("Starting %s server on %s", srv.cfg.Net, srv.cfg.Listen)
	return srv.dnsServer.ListenAndServe()
}

func (srv *dnsServer) Shutdown() error {
	log.Infof("Shutdown %s server on %s", srv.cfg.Net, srv.cfg.Listen)
	srv.router.Shutdown()
	return srv.dnsServer.Shutdown()
}

// process dns request
func (srv *dnsServer) handleDNSRequest(w dns.ResponseWriter, r *dns.Msg) {
	logAccessRequest(r)

	// process request
	resp, err := srv.processServe(r)
	if err != nil {
		log.Errorf("request:%s processServe error:%s", r, err)
		// TODO: write empty msg to client?
		return
	}

	// response
	logResponse(resp)
	// write to client
	w.WriteMsg(resp)
}

func (srv *dnsServer) processServe(r *dns.Msg) (*dns.Msg, error) {
	// find group by router
	invoker, err := srv.router.FindGroupInvoker(r)
	if err != nil {
		return nil, err
	}
	// group invoke
	return invoker.Invoke(r)
}

// access request
func logAccessRequest(r *dns.Msg) {
	if len(r.Question) == 0 {
		return
	}
	log.Infof("Received DNS request from %s", r.Question[0].Name)
	for _, question := range r.Question {
		log.Infof("Question:%s %s %s", question.Name, dns.Class(question.Qclass).String(), dns.Type(question.Qtype).String())
	}
}

// access response
func logResponse(r *dns.Msg) {
	for _, answer := range r.Answer {
		log.Infof("%T Answer: %s", answer, answer.String())
	}
}
