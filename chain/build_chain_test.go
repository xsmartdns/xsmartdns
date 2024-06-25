package chain

import (
	"testing"

	"github.com/miekg/dns"
	. "github.com/smartystreets/goconvey/convey"
)

type mockChain struct {
	id uint16
}

func (c *mockChain) HandleRequest(r *dns.Msg, nextChain HandleInvoke) (*dns.Msg, error) {
	r.Id = c.id
	return nextChain(r)
}
func (c *mockChain) Shutdown() {
}

type mockChainStop struct {
	id uint16
}

func (c *mockChainStop) HandleRequest(r *dns.Msg, nextChain HandleInvoke) (*dns.Msg, error) {
	r.Id = c.id
	return r, nil
}
func (c *mockChainStop) Shutdown() {
}
func TestChainList(t *testing.T) {
	Convey("TestChainList", t, func() {
		handleInvoke, _ := BuildChain(&mockChain{id: 1}, &mockChain{id: 2}, &mockChain{id: 3})
		ret, err := handleInvoke(&dns.Msg{})
		So(err, ShouldBeNil)
		So(ret.Id, ShouldEqual, 3)

		handleInvoke, _ = BuildChain()
		ret, err = handleInvoke(&dns.Msg{})
		So(err, ShouldBeNil)
		So(ret.Id, ShouldEqual, 0)

		handleInvoke, _ = BuildChain(&mockChain{id: 1})
		ret, err = handleInvoke(&dns.Msg{})
		So(err, ShouldBeNil)
		So(ret.Id, ShouldEqual, 1)
	})

	Convey("TestChainList", t, func() {
		handleInvoke, _ := BuildChain(&mockChainStop{id: 1}, &mockChainStop{id: 2}, &mockChainStop{id: 3})
		ret, err := handleInvoke(&dns.Msg{})
		So(err, ShouldBeNil)
		So(ret.Id, ShouldEqual, 1)
	})
}
