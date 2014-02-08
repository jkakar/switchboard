package sb_test

import (
	"github.com/coreos/go-etcd/etcd"
	"github.com/jkakar/switchboard"
	. "launchpad.net/gocheck"
	"net/http"
)

type ExchangeTest struct {
	client *etcd.Client
}

var _ = Suite(&ExchangeTest{})

// Delete the "test" namespace in etcd.
func (s *ExchangeTest) SetUpTest(c *C) {
	s.client = etcd.NewClient([]string{"http://127.0.0.1:4001"})
	s.client.Delete("test", true)
}

// ServiceManifest returns an error if the specified namespace doesn't exist
// in etcd.
func (s *ExchangeTest) TestServiceManifestWithoutData(c *C) {
	exchange := sb.NewExchange("test", s.client)
	_, err := exchange.ServiceManifest()
	c.Assert(err.(*etcd.EtcdError).ErrorCode, Equals, 100)
}

// ServiceManifest connects to etcd and discovers currently registered
// services.
func (s *ExchangeTest) TestServiceManifest(c *C) {
	service := sb.NewService("test", s.client, "http://localhost:8080", make(sb.Routes), http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	serviceRecord, err := service.Notify()
	c.Assert(err, IsNil)

	exchange := sb.NewExchange("test", s.client)
	manifest, err := exchange.ServiceManifest()
	c.Assert(len(manifest.Services), Equals, 1)
	c.Assert(manifest.Services[0], DeepEquals, serviceRecord)
}

// // MarshalExchange converts a Exchange struct into a tnetstring encoded message.
// func (s *ExchangeTest) TestMarshalExchange(c *C) {
// 	headers := http.Header{"Content-Type": []string{"text/plain"}}
// 	userData := zhttp.UserData{"Version": "1"}
// 	response := zhttp.Exchange{
// 		Id:        "message-id",
// 		Type:      "",
// 		Condition: "",
// 		Code:      201,
// 		Reason:    "Created",
// 		Headers:   headers,
// 		Body:      "Hello, world!",
// 		UserData:  userData}
// 	message, err := zhttp.MarshalExchange(&response)
// 	c.Assert(err, IsNil)
// 	c.Assert(message, Equals, "177:2:Id,10:message-id,4:Type,0:,9:Condition,0:,4:Code,3:201#6:Reason,7:Created,7:Headers,34:12:Content-Type,14:10:text/plain,]}4:Body,13:Hello, world!,8:UserData,14:7:Version,1:1,}}")
// }

// // UnmarshalExchange converts a tnetstring encoded message into a Exchange
// // struct.
// func (s *ExchangeTest) TestUnmarshalExchange(c *C) {
// 	headers := http.Header{"Content-Type": []string{"text/plain"}}
// 	userData := zhttp.UserData{"Version": "1"}
// 	response1 := zhttp.Exchange{
// 		Id:        "message-id",
// 		Type:      "error",
// 		Condition: "",
// 		Code:      404,
// 		Reason:    "Not Found",
// 		Headers:   headers,
// 		Body:      "",
// 		UserData:  userData}
// 	message, _ := zhttp.MarshalExchange(&response1)

// 	var response2 zhttp.Exchange
// 	err := zhttp.UnmarshalExchange(message, &response2)
// 	c.Assert(err, IsNil)
// 	c.Assert(response1, DeepEquals, response2)
// }
