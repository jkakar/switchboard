package switchboard

import (
	"net/http"

	"github.com/coreos/go-etcd/etcd"
	. "launchpad.net/gocheck"
)

type ExchangeTest struct {
	client   *etcd.Client
	exchange *Exchange
	mux      *ExchangeServeMux
}

var _ = Suite(&ExchangeTest{})

func (s *ExchangeTest) SetUpTest(c *C) {
	s.client = etcd.NewClient([]string{"http://127.0.0.1:4001"})
	s.client.Delete("test", true)
	s.mux = NewExchangeServeMux()
	s.exchange = NewExchange("test", s.client, s.mux)
}

// Init returns an error if the specified namespace doesn't exist in etcd.
func (s *ExchangeTest) TestInitWithoutServices(c *C) {
	err := s.exchange.Init()
	c.Assert(err.(*etcd.EtcdError).ErrorCode, Equals, 100)
}

// Init returns an error if the specified namespace doesn't exist in etcd.
func (s *ExchangeTest) TestInit(c *C) {
	address := "http://localhost:8080"
	routes := Routes{"GET": []string{"/users", "/user/:id"}}
	handler := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})
	service := NewService("test", s.client, address, routes, handler)
	_, err := service.Notify(0)
	c.Assert(err, IsNil)

	err = s.exchange.Init()
	c.Assert(err, IsNil)
	addresses, err := s.mux.Match("GET", "/users")
	c.Assert(err, IsNil)
	c.Assert(addresses, DeepEquals, &[]string{"http://localhost:8080"})
	addresses, err = s.mux.Match("GET", "/user/123")
	c.Assert(err, IsNil)
	c.Assert(addresses, DeepEquals, &[]string{"http://localhost:8080"})
}

// // ServiceManifest returns an error if the specified namespace doesn't exist
// // in etcd.
// func (s *ExchangeTest) TestServiceManifestWithoutData(c *C) {
// 	exchange := NewExchange("test", s.client)
// 	_, err := exchange.ServiceManifest()
// 	c.Assert(err.(*etcd.EtcdError).ErrorCode, Equals, 100)
// }

/*
import (
	"net/http"

	"github.com/coreos/go-etcd/etcd"
	. "launchpad.net/gocheck"
)

type ExchangeTest struct {
	client *etcd.Client
}

var _ = Suite(&ExchangeTest{})

func (s *ExchangeTest) SetUpTest(c *C) {
	s.client = etcd.NewClient([]string{"http://127.0.0.1:4001"})
	s.client.Delete("test", true)
}

// ServiceManifest returns an error if the specified namespace doesn't exist
// in etcd.
func (s *ExchangeTest) TestServiceManifestWithoutData(c *C) {
	exchange := NewExchange("test", s.client)
	_, err := exchange.ServiceManifest()
	c.Assert(err.(*etcd.EtcdError).ErrorCode, Equals, 100)
}

// ServiceManifest connects to etcd and discovers currently registered
// services.
func (s *ExchangeTest) TestServiceManifest(c *C) {
	service := NewService("test", s.client, "http://localhost:8080", make(Routes), http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	serviceRecord, err := service.Notify(0)
	c.Assert(err, IsNil)

	exchange := NewExchange("test", s.client)
	manifest, err := exchange.ServiceManifest()
	c.Assert(len(manifest.Services), Equals, 1)
	c.Assert(manifest.Services[0], DeepEquals, serviceRecord)
}

// Watch is disabled when when a message is sent on the stop channel.
func (s *ExchangeTest) TestWatchStops(c *C) {
	service := NewService("test", s.client, "http://localhost:8080", make(Routes), http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	_, err := service.Notify(0)
	c.Assert(err, IsNil)

	exchange := NewExchange("test", s.client)
	stop := make(chan bool)
	stopped := make(chan bool)
	go func() {
		update := make(chan *ServiceManifest)
		err = exchange.Watch(update, stop)
		c.Assert(err, IsNil)
		stopped <- true
	}()
	stop <- true
	c.Assert(<-stopped, Equals, true)
}

func (s *ExchangeTest) TestWatch(c *C) {
	service := NewService("test", s.client, "http://localhost:8080", make(Routes), http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	_, err := service.Notify(0)
	c.Assert(err, IsNil)

}
*/
