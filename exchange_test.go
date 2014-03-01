package sb

import (
	"github.com/coreos/go-etcd/etcd"
	. "launchpad.net/gocheck"
	"net/http"
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
	serviceRecord, err := service.Notify()
	c.Assert(err, IsNil)

	exchange := NewExchange("test", s.client)
	manifest, err := exchange.ServiceManifest()
	c.Assert(len(manifest.Services), Equals, 1)
	c.Assert(manifest.Services[0], DeepEquals, serviceRecord)
}

// Watch is disabled when when a message is sent on the stop channel.
func (s *ExchangeTest) TestWatchStops(c *C) {
	service := NewService("test", s.client, "http://localhost:8080", make(Routes), http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	_, err := service.Notify()
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
	_, err := service.Notify()
	c.Assert(err, IsNil)

}
