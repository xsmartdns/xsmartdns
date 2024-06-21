package main

import (
	"flag"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/xsmartdns/xsmartdns/config"
	"github.com/xsmartdns/xsmartdns/log"
	"github.com/xsmartdns/xsmartdns/router"
	"github.com/xsmartdns/xsmartdns/server"
)

var (
	configFile string
)

func init() {
	flag.StringVar(&configFile, "c", "", "config file")
}

func main() {
	// init config
	flag.Parse()
	if len(configFile) == 0 {
		flag.Usage()
		return
	}
	cfg, err := parseConfig()
	if err != nil {
		panic("parse config err:" + err.Error())
	}

	// init log
	log.Init(&cfg.Log)
	// init router
	router := router.NewGroupRouter(cfg)
	// init inbounds
	srvs := initInbounds(cfg, router)
	// start and block to wait shutdown
	startInbounds(srvs)
}

func parseConfig() (*config.Config, error) {
	b, err := os.ReadFile(configFile)
	if err != nil {
		return nil, err
	}
	return config.Parse(b)
}

func initInbounds(cfg *config.Config, router router.Router) []server.Server {
	srvs := make([]server.Server, 0, len(cfg.Inbounds))
	for _, inbound := range cfg.Inbounds {
		srv := server.NewDnsServer(*inbound, router)
		if err := srv.Init(); err != nil {
			log.Fatalf("inbound:%v init err:%v", inbound, err)
		}
		srvs = append(srvs, srv)
	}
	return srvs
}

func startInbounds(srvs []server.Server) {
	wg := sync.WaitGroup{}
	for _, srv := range srvs {
		wg.Add(1)
		go func(srv server.Server) {
			defer wg.Done()
			if err := srv.Start(); err != nil {
				log.Errorf("inbound start err:%v", err)
			}
		}(srv)
	}
	go func() {
		wg.Wait()
		log.Fatalf("all inbound is stopped")
	}()
	// wait signal to shutdown
	waitShtdown(&wg, srvs)
}

func waitShtdown(wg *sync.WaitGroup, srvs []server.Server) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigChan
	log.Infof("Received signal: %s", sig)
	chStoped := make(chan struct{}, 1)
	// shutdown and wait
	go func() {
		for _, srv := range srvs {
			srv.Shutdown()
		}
		wg.Wait()
		chStoped <- struct{}{}
	}()
	// shutdown or wait timeout
	select {
	case <-chStoped:
		log.Infof("inbounds all shutdown")
	case <-time.After(10 * time.Second):
		log.Errorf("wait shutdown timeout")
	}
}
