package router

import (
	"fmt"

	"github.com/miekg/dns"
	"github.com/xsmartdns/xsmartdns/config"
	"github.com/xsmartdns/xsmartdns/group"
)

// match and find group
type groupRouter struct {
	cfg          *config.Config
	defaultGroup *config.Group
	// key:group tag
	groupMap map[string]group.GroupInvoker
}

func NewGroupRouter(cfg *config.Config) Router {
	groupMap := make(map[string]group.GroupInvoker)
	for _, g := range cfg.Groups {
		groupMap[g.Tag] = group.NewFastlyGroupInvoker(g)
	}
	return &groupRouter{cfg: cfg, groupMap: groupMap, defaultGroup: cfg.Groups[0]}
}

func (router *groupRouter) FindGroupInvoker(r *dns.Msg) (group.GroupInvoker, error) {
	groupCfg := router.findGroupCfg(r)
	g := router.groupMap[groupCfg.Tag]
	if g == nil {
		return nil, fmt.Errorf("group:%s not found", groupCfg.Tag)
	}
	return g, nil
}

func (router *groupRouter) findGroupCfg(r *dns.Msg) *config.Group {
	if len(router.cfg.Routing) == 0 {
		return router.defaultGroup
	}
	// TODO: geosite and domain match
	return router.defaultGroup
}
