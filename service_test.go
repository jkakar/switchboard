package switchboard

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"github.com/coreos/go-etcd/etcd"
	. "launchpad.net/gocheck"
)

type ServiceTest struct {
	client *etcd.Client
}

var _ = Suite(&ServiceTest{})

func (s *ServiceTest) SetUpTest(c *C) {
	s.client = etcd.NewClient([]string{"http://127.0.0.1:4001"})
	s.client.Delete("test", true)
}

// Register creates a service record to represent the service and registers it
// in etcd.
func (s *ServiceTest) TestRegister(c *C) {
	address := "http://localhost:8080"
	routes := Routes{"GET": []string{"/users", "/user/:id"}}
	handler := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})
	service := NewService("test", s.client, address, routes, handler)
	record, err := service.Register(0)
	c.Assert(err, IsNil)

	key := "test/" + service.ID()
	sort := false
	recursive := true
	response, err := s.client.Get(key, sort, recursive)
	c.Assert(err, IsNil)
	recordJSON, _ := json.Marshal(record)
	c.Assert(response.Node.Value, Equals, bytes.NewBuffer(recordJSON).String())
}

// Register is effectively a no-op if the service record already exists in
// etcd.
func (s *ServiceTest) TestRegisterDuplicate(c *C) {
	address := "http://localhost:8080"
	routes := Routes{"GET": []string{"/users", "/user/:id"}}
	handler := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})
	service := NewService("test", s.client, address, routes, handler)
	_, err := service.Register(0)
	c.Assert(err, IsNil)
	_, err = service.Register(0)
	c.Assert(err, IsNil)
}

// Unregister returns an error if the service isn't registered in etcd.
func (s *ServiceTest) TestUnregisterUnregisteredService(c *C) {
	address := "http://localhost:8080"
	routes := Routes{"GET": []string{"/users", "/user/:id"}}
	handler := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})
	service := NewService("test", s.client, address, routes, handler)
	err := service.Unregister()
	c.Assert(err.(*etcd.EtcdError).ErrorCode, Equals, 100)
}

// Unregister deletes the service record in etcd.
func (s *ServiceTest) TestUnregister(c *C) {
	address := "http://localhost:8080"
	routes := Routes{"GET": []string{"/users", "/user/:id"}}
	handler := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})
	service := NewService("test", s.client, address, routes, handler)
	_, err := service.Register(0)
	c.Assert(err, IsNil)
	err = service.Unregister()
	c.Assert(err, IsNil)

	key := "test/" + service.ID()
	sort := false
	recursive := true
	_, err = s.client.Get(key, sort, recursive)
	c.Assert(err.(*etcd.EtcdError).ErrorCode, Equals, 100)
}

// Broadcast stops pushing changes to etcd when when a bool value is sent to
// the stop channel.
func (s *ServiceTest) TestBroadcastStops(c *C) {
	address := "http://localhost:8080"
	routes := Routes{"GET": []string{"/users", "/user/:id"}}
	handler := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})
	service := NewService("test", s.client, address, routes, handler)

	stop := make(chan bool)
	stopped := make(chan bool)
	go func() {
		service.Broadcast(2, 1, stop)
		stopped <- true
	}()

	stop <- true
	c.Assert(<-stopped, Equals, true)
}

// Broadcast registers a service with etcd.
func (s *ServiceTest) TestBroadcast(c *C) {
	// Create a new service.
	address := "http://localhost:8080"
	routes := Routes{"GET": []string{"/users", "/user/:id"}}
	handler := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})
	service := NewService("test", s.client, address, routes, handler)

	// Start broadcasting service changes to etcd.
	stop := make(chan bool)
	stopped := make(chan bool)
	go func() {
		service.Broadcast(2, 1, stop)
		stopped <- true
	}()

	// Cleanly shutdown the broadcaster no matter how the test finishes.
	defer func() {
		stop <- true
		c.Assert(<-stopped, Equals, true)
	}()

	// Janky logic to wait for the broadcaster to write a value to etcd that
	// can be read back will fail when service registration doesn't complete
	// within 500ms.
	receivedUpdate := false
	for i := 0; i < 500; i++ {
		key := "test/" + service.ID()
		sort := false
		recursive := true
		response, err := s.client.Get(key, sort, recursive)
		if err != nil {
			time.Sleep(time.Duration(1) * time.Millisecond)
			continue
		} else {
			record := ServiceRecord{
				ID:      service.id,
				Address: service.address,
				Routes:  service.routes}
			recordJSON, _ := json.Marshal(record)
			c.Assert(response.Node.Value, Equals, bytes.NewBuffer(recordJSON).String())
			receivedUpdate = true
			break
		}
	}
	c.Assert(receivedUpdate, Equals, true)
}
