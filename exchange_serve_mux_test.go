package switchboard

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	. "launchpad.net/gocheck"
)

type ExchangeServeMuxTest struct{}

var _ = Suite(&ExchangeServeMuxTest{})

// ServeHTTP returns a 404 Not Found when no pattern matches the requested
// route.
func (s *ExchangeServeMuxTest) TestServeHTTPWithUnknownRoute(c *C) {
	writer := httptest.NewRecorder()
	request, err := http.NewRequest("GET", "http://example.com/resource", nil)
	c.Assert(err, IsNil)

	mux := NewExchangeServeMux()
	mux.ServeHTTP(writer, request)
	c.Assert(writer.Code, Equals, http.StatusNotFound)
}

// ServeHTTP only proxies requests that match registered HTTP methods.
func (s *ExchangeServeMuxTest) TestServeHTTPConsidersHTTPMethod(c *C) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	server := httptest.NewServer(handler)
	defer server.Close()
	writer := httptest.NewRecorder()
	request, err := http.NewRequest("HEAD", "http://example.com/resource", nil)
	c.Assert(err, IsNil)

	mux := NewExchangeServeMux()
	mux.Add("GET", "/resource", server.URL)
	mux.ServeHTTP(writer, request)
	c.Assert(writer.Code, Equals, http.StatusNotFound)
}

// ServeHTTP proxies requests to static routes registered with Add.
func (s *ExchangeServeMuxTest) TestServeHTTPWithStaticRoute(c *C) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, world!")
	})
	server := httptest.NewServer(handler)
	defer server.Close()
	writer := httptest.NewRecorder()
	request, err := http.NewRequest("GET", "http://example.com/resource", nil)
	c.Assert(err, IsNil)

	mux := NewExchangeServeMux()
	mux.Add("GET", "/resource", server.URL)
	mux.ServeHTTP(writer, request)
	c.Assert(writer.Code, Equals, http.StatusOK)
	c.Assert(writer.Body.String(), Equals, "Hello, world!\n")
}

// ServeHTTP passes query string arguments received from clients to service
// backends.
func (s *ExchangeServeMuxTest) TestServeHTTPWithQueryString(c *C) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, r.URL.RawQuery)
	})
	server := httptest.NewServer(handler)
	defer server.Close()
	writer := httptest.NewRecorder()
	url := "http://example.com/resource?key=value&key1=value1&key1=value2"
	request, err := http.NewRequest("GET", url, nil)
	c.Assert(err, IsNil)

	mux := NewExchangeServeMux()
	mux.Add("GET", "/resource", server.URL)
	mux.ServeHTTP(writer, request)
	c.Assert(writer.Code, Equals, http.StatusOK)
	c.Assert(writer.Body.String(), Equals, "key=value&key1=value1&key1=value2\n")
}

// ServeHTTP passes headers received from clients to service backends and
// returns headers received from service backends back to clients.
func (s *ExchangeServeMuxTest) TestServeHTTPWithHeaders(c *C) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.Assert(r.Header, DeepEquals, http.Header{
			"User-Agent":      []string{"Go 1.1 package http"},
			"Accept-Encoding": []string{"gzip"},
			"X-From-Client":   []string{"Client"}})
		w.Header().Add("X-From-Service", "Service")
	})
	server := httptest.NewServer(handler)
	defer server.Close()
	writer := httptest.NewRecorder()
	url := "http://example.com/resource?key=value&key1=value1&key1=value2"
	request, err := http.NewRequest("GET", url, nil)
	request.Header.Add("X-From-Client", "Client")
	c.Assert(err, IsNil)

	mux := NewExchangeServeMux()
	mux.Add("GET", "/resource", server.URL)
	mux.ServeHTTP(writer, request)
	c.Assert(writer.Code, Equals, http.StatusOK)
	writer.Header().Del("Date")
	c.Assert(writer.Header(), DeepEquals, http.Header{
		"Content-Length": []string{"0"},
		"Content-Type":   []string{"text/plain; charset=utf-8"},
		"X-From-Service": []string{"Service"}})
}

// ServeHTTP proxies requests to dynamic routes registered with Add.
func (s *ExchangeServeMuxTest) TestServeHTTPWithDynamicRoute(c *C) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, world!")
	})
	server := httptest.NewServer(handler)
	defer server.Close()
	writer := httptest.NewRecorder()
	request, err := http.NewRequest("GET", "http://example.com/resource/1", nil)
	c.Assert(err, IsNil)

	mux := NewExchangeServeMux()
	mux.Add("GET", "/resource/:id", server.URL)
	mux.ServeHTTP(writer, request)
	c.Assert(writer.Code, Equals, http.StatusOK)
	c.Assert(writer.Body.String(), Equals, "Hello, world!\n")
}
