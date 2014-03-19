package switchboard

import (
	"bytes"
	"encoding/json"
	"net/http"

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

// Notify creates a service record to represent the service and registers it
// in etcd.
func (s *ServiceTest) TestNotify(c *C) {
	address := "http://localhost:8080"
	routes := Routes{"GET": []string{"/users", "/user/:id"}}
	handler := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})
	service := NewService("test", s.client, address, routes, handler)
	record, err := service.Notify(0)

	key := "test/" + service.ID()
	sort := false
	recursive := true
	response, err := s.client.Get(key, sort, recursive)
	c.Assert(err, IsNil)
	recordJSON, _ := json.Marshal(record)
	c.Assert(response.Node.Value, Equals, bytes.NewBuffer(recordJSON).String())
}
