package switchboard_test

import (
	"time"

	"github.com/coreos/go-etcd/etcd"
	"github.com/jkakar/switchboard"
	. "launchpad.net/gocheck"
)

type ExchangeTest struct {
	client   *etcd.Client
	exchange *switchboard.Exchange
	mux      *switchboard.ExchangeServeMux
}

var _ = Suite(&ExchangeTest{})

func (s *ExchangeTest) SetUpTest(c *C) {
	s.client = etcd.NewClient([]string{"http://127.0.0.1:4001"})
	s.client.Delete("test", true)
	s.mux = switchboard.NewExchangeServeMux()
	s.exchange = switchboard.NewExchange("test", s.client, s.mux)
}

// Init returns an error if the specified namespace doesn't exist in etcd.
func (s *ExchangeTest) TestInitCreatesNamespaceDirectory(c *C) {
	err := s.exchange.Init()
	c.Assert(err, IsNil)
	response, err := s.client.Get("test", false, false)
	c.Assert(err, IsNil)
	c.Assert(response.Node.Key, Equals, "/test")
}

// Init returns an error if the specified namespace doesn't exist in etcd.
func (s *ExchangeTest) TestInit(c *C) {
	address := "http://localhost:8080"
	routes := switchboard.Routes{"GET": []string{"/users", "/user/:id"}}
	service := switchboard.NewService("test", s.client, address, routes)
	_, err := service.Register(0)
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

// Watch stops observing changes in etcd when when a bool value is sent to the
// stop channel.
func (s *ExchangeTest) TestWatchStops(c *C) {
	address := "http://localhost:8080"
	routes := switchboard.Routes{"GET": []string{"/users", "/user/:id"}}
	service := switchboard.NewService("test", s.client, address, routes)
	_, err := service.Register(0)
	c.Assert(err, IsNil)

	stop := make(chan bool)
	stopped := make(chan bool)
	go func() {
		err = s.exchange.Init()
		c.Assert(err, IsNil)
		s.exchange.Watch(stop)
		stopped <- true
	}()

	stop <- true
	c.Assert(<-stopped, Equals, true)
}

// Watch observes newly registered services in etcd and adds the relevant
// patterns and addresses to the ExchangeServeMux.
func (s *ExchangeTest) TestWatchRegisteredService(c *C) {
	// Create the namespace in etcd and initialize the exchange.
	_, err := s.client.CreateDir("/test", 0)
	c.Assert(err, IsNil)
	err = s.exchange.Init()
	c.Assert(err, IsNil)

	// Start the watcher goroutine.
	stop := make(chan bool)
	stopped := make(chan bool)
	go func() {
		s.exchange.Watch(stop)
		stopped <- true
	}()

	// Cleanly shutdown the watcher no matter how the test finishes.
	defer func() {
		stop <- true
		c.Assert(<-stopped, Equals, true)
	}()

	// Register a new service.
	address := "http://localhost:8080"
	routes := switchboard.Routes{"GET": []string{"/users"}}
	service := switchboard.NewService("test", s.client, address, routes)
	_, err = service.Register(0)
	c.Assert(err, IsNil)

	// Janky logic to wait for updates from etcd will fail when updates don't
	// propagate within 500ms.
	receivedUpdate := false
	for i := 0; i < 500; i++ {
		addresses, err := s.mux.Match("GET", "/users")
		if err != nil {
			time.Sleep(time.Duration(1) * time.Millisecond)
			continue
		} else {
			c.Assert(addresses, DeepEquals, &[]string{"http://localhost:8080"})
			receivedUpdate = true
			break
		}
	}
	c.Assert(receivedUpdate, Equals, true)
}

// Watch observes unregistered services in etcd and removes the relevant
// patterns and addresses from the ExchangeServeMux.
func (s *ExchangeTest) TestWatchUnegisteredService(c *C) {
	// Create the namespace in etcd.
	_, err := s.client.CreateDir("/test", 0)
	c.Assert(err, IsNil)

	// Register a new service.
	address := "http://localhost:8080"
	routes := switchboard.Routes{"GET": []string{"/users"}}
	service := switchboard.NewService("test", s.client, address, routes)
	_, err = service.Register(0)
	c.Assert(err, IsNil)

	// Initialize the exchange.
	err = s.exchange.Init()
	c.Assert(err, IsNil)

	// Start the watcher goroutine.
	stop := make(chan bool)
	stopped := make(chan bool)
	go func() {
		s.exchange.Watch(stop)
		stopped <- true
	}()

	// Cleanly shutdown the watcher no matter how the test finishes.
	defer func() {
		stop <- true
		c.Assert(<-stopped, Equals, true)
	}()

	// Unregister the service.
	err = service.Unregister()
	c.Assert(err, IsNil)

	// Janky logic to wait for updates from etcd will fail when updates don't
	// propagate within 500ms.
	receivedUpdate := false
	for i := 0; i < 500; i++ {
		addresses, err := s.mux.Match("GET", "/users")
		if err == nil {
			time.Sleep(time.Duration(1) * time.Millisecond)
			continue
		} else {
			c.Assert(addresses, IsNil)
			receivedUpdate = true
			break
		}
	}
	c.Assert(receivedUpdate, Equals, true)
}
