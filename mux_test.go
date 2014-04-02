package switchboard

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	. "gopkg.in/check.v1"
)

type ExchangeServeMuxTest struct{}

var _ = Suite(&ExchangeServeMuxTest{})

// Add registers a new method, pattern and address.  It creates data
// structures as needed.
func (s *ExchangeServeMuxTest) TestAdd(c *C) {
	mux := NewExchangeServeMux()
	mux.Add("GET", "/resource", "http://example.com")
	c.Assert(len(mux.routes), Equals, 1)
	handlers := mux.routes["GET"]
	c.Assert(len(handlers), Equals, 1)
	c.Assert(handlers[0].pattern, Equals, "/resource")
	c.Assert(handlers[0].addresses, DeepEquals, []string{"http://example.com"})
}

// Add is a effectively a no-op if a duplicate method, pattern and address are
// provided.
func (s *ExchangeServeMuxTest) TestAddDuplicatePatternAndAddress(c *C) {
	mux := NewExchangeServeMux()
	mux.Add("GET", "/resource", "http://example.com")
	mux.Add("GET", "/resource", "http://example.com")
	c.Assert(len(mux.routes), Equals, 1)
	handlers := mux.routes["GET"]
	c.Assert(len(handlers), Equals, 1)
	c.Assert(handlers[0].pattern, Equals, "/resource")
	c.Assert(handlers[0].addresses, DeepEquals, []string{"http://example.com"})
}

// Add appends new addresses to an existing pattern handler registered for a
// method and pattern.
func (s *ExchangeServeMuxTest) TestAddAddressToExistingHandler(c *C) {
	mux := NewExchangeServeMux()
	mux.Add("GET", "/resource", "http://example.com:8080")
	mux.Add("GET", "/resource", "http://example.com:8081")
	c.Assert(len(mux.routes), Equals, 1)
	handlers := mux.routes["GET"]
	c.Assert(len(handlers), Equals, 1)
	c.Assert(handlers[0].pattern, Equals, "/resource")
	expected := []string{"http://example.com:8080", "http://example.com:8081"}
	c.Assert(handlers[0].addresses, DeepEquals, expected)
}

// Add creates a new pattern handler for each new pattern.
func (s *ExchangeServeMuxTest) TestAddMultiplePatterns(c *C) {
	mux := NewExchangeServeMux()
	mux.Add("GET", "/resource0", "http://example.com")
	mux.Add("GET", "/resource1", "http://example.com")
	c.Assert(len(mux.routes), Equals, 1)
	handlers := mux.routes["GET"]
	c.Assert(len(handlers), Equals, 2)
	c.Assert(handlers[0].pattern, Equals, "/resource0")
	c.Assert(handlers[0].addresses, DeepEquals, []string{"http://example.com"})
	c.Assert(handlers[1].pattern, Equals, "/resource1")
	c.Assert(handlers[1].addresses, DeepEquals, []string{"http://example.com"})
}

// Remove is a effectively a no-op if the requested method doesn't exist.
func (s *ExchangeServeMuxTest) TestRemoveWithoutMatchingMethod(c *C) {
	mux := NewExchangeServeMux()
	mux.Remove("GET", "/resource", "http://example.com")
	c.Assert(len(mux.routes), Equals, 0)
}

// Remove removes a pattern handler when it no longer contains addresses.
func (s *ExchangeServeMuxTest) TestRemovePatternHandler(c *C) {
	mux := NewExchangeServeMux()
	mux.Add("GET", "/resource", "http://example.com")
	mux.Remove("GET", "/resource", "http://example.com")
	c.Assert(len(mux.routes), Equals, 1)
	handlers := mux.routes["GET"]
	c.Assert(len(handlers), Equals, 0)
}

// Remove removes a registered address from a pattern handler.
func (s *ExchangeServeMuxTest) TestRemoveFirstAddress(c *C) {
	mux := NewExchangeServeMux()
	mux.Add("GET", "/resource", "http://example.com:8080")
	mux.Add("GET", "/resource", "http://example.com:8081")
	mux.Remove("GET", "/resource", "http://example.com:8080")
	c.Assert(len(mux.routes), Equals, 1)
	handlers := mux.routes["GET"]
	c.Assert(len(handlers), Equals, 1)
	c.Assert(handlers[0].pattern, Equals, "/resource")
	c.Assert(handlers[0].addresses, DeepEquals, []string{"http://example.com:8081"})
}

// Remove removes a registered address from a pattern handler.
func (s *ExchangeServeMuxTest) TestRemoveLastAddress(c *C) {
	mux := NewExchangeServeMux()
	mux.Add("GET", "/resource", "http://example.com:8080")
	mux.Add("GET", "/resource", "http://example.com:8081")
	mux.Remove("GET", "/resource", "http://example.com:8081")
	c.Assert(len(mux.routes), Equals, 1)
	handlers := mux.routes["GET"]
	c.Assert(len(handlers), Equals, 1)
	c.Assert(handlers[0].pattern, Equals, "/resource")
	c.Assert(handlers[0].addresses, DeepEquals, []string{"http://example.com:8080"})
}

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

// ServeHTTP proxies requests to dynamic routes registered with Add.
func (s *ExchangeServeMuxTest) TestServeHTTPWithBestMatch(c *C) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, world!")
	})
	server := httptest.NewServer(handler)
	defer server.Close()
	writer := httptest.NewRecorder()
	request, err := http.NewRequest("GET", "http://example.com/resource/1/2", nil)
	c.Assert(err, IsNil)

	mux := NewExchangeServeMux()
	mux.Add("GET", "/resource/:one", "broken")
	mux.Add("GET", "/resource/:one/:two", server.URL)
	mux.ServeHTTP(writer, request)
	c.Assert(writer.Code, Equals, http.StatusOK)
	c.Assert(writer.Body.String(), Equals, "Hello, world!\n")
}

type PatternHandlerTest struct{}

var _ = Suite(&PatternHandlerTest{})

// Match matches paths to patterns that don't have a placeholder.
func (s *PatternHandlerTest) TestMatchWithoutPlaceholder(c *C) {
	handler := patternHandler{pattern: "/foo"}
	c.Assert(handler.Match("/foo"), Equals, true)
	c.Assert(handler.Match("/foo/bar"), Equals, false)
}

// Match matches paths to patterns that have a placeholder at then end of the
// pattern.
func (s *PatternHandlerTest) TestMatchWithPlaceholder(c *C) {
	handler := patternHandler{pattern: "/foo/:name"}
	c.Assert(handler.Match("/foo/bar"), Equals, true)
	c.Assert(handler.Match("/foo"), Equals, false)
}

// Match matches paths to patterns that have a placeholder in the middle of
// the pattern.
func (s *PatternHandlerTest) TestMatchWithEmbeddedPlaceholder(c *C) {
	handler := patternHandler{pattern: "/foo/:name/baz"}
	c.Assert(handler.Match("/foo/bar/baz"), Equals, true)
}

// Match matches paths to patterns that have multiple placeholders.
func (s *PatternHandlerTest) TestMatchWithMultiplePlaceholders(c *C) {
	handler := patternHandler{pattern: "/foo/:name/baz/:id"}
	c.Assert(handler.Match("/foo/bar/baz"), Equals, false)
	c.Assert(handler.Match("/foo/bar/baz/123"), Equals, true)
}

// Match matches paths to patterns that have multiple placeholders with the
// same name.
func (s *PatternHandlerTest) TestMatchWithDuplicatePlaceholders(c *C) {
	handler := patternHandler{pattern: "/foo/:name/baz/:name"}
	c.Assert(handler.Match("/foo/bar/baz"), Equals, false)
	c.Assert(handler.Match("/foo/bar/baz/123"), Equals, true)
}

// Match matches paths to patterns that have placeholders with colons in their
// name.
func (s *PatternHandlerTest) TestMatchWithDoubleColonPlaceholder(c *C) {
	handler := patternHandler{pattern: "/foo/::name"}
	c.Assert(handler.Match("/foo/bar"), Equals, true)
}

// Match matches paths to patterns that have placeholders with a constant
// prefix string.
func (s *PatternHandlerTest) TestMatchWithPrefixedPlaceholder(c *C) {
	handler := patternHandler{pattern: "/foo/x:name"}
	c.Assert(handler.Match("/foo/xbar"), Equals, true)
	c.Assert(handler.Match("/foo/bar"), Equals, false)
}

// Match treats patterns that end in a trailing slash as ending in a splat.
// That is, anything after the trailing slash in the path is considered a
// match.
func (s *PatternHandlerTest) TestMatchWithSplat(c *C) {
	handler := patternHandler{pattern: "/foo/"}
	c.Assert(handler.Match("/foo/bar/baz"), Equals, true)
	c.Assert(handler.Match("/foo/bar"), Equals, true)
}

// Match matches paths to patterns that have placeholders and end in a splat.
func (s *PatternHandlerTest) TestMatchWithPrefixAndSplat(c *C) {
	handler := patternHandler{pattern: "/foo/:name/bar/"}
	c.Assert(handler.Match("/foo/name/bar/baz"), Equals, true)
	c.Assert(handler.Match("/foo/name/bar/baz/quux"), Equals, true)
}
