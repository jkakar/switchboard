package switchboard

import (
	"bytes"
	"errors"
	"math/rand"
	"net/http"
	"sync"
)

// ExchangeServeMux is an HTTP request multiplexer.  It matches the URL of
// each incoming request against a list of registered patterns to find the
// service that can respond to it and proxies the request to the appropriate
// backend.
type ExchangeServeMux struct {
	rw     sync.RWMutex                 // Synchronize access to routes map.
	routes map[string][]*patternHandler // Patterns mapped to backend services.
}

// NewExchangeServeMux allocates and returns a new ExchangeServeMux.
func NewExchangeServeMux() *ExchangeServeMux {
	return &ExchangeServeMux{routes: make(map[string][]*patternHandler)}
}

// Add registers the address for a backend service as a handler for an HTTP
// method and URL pattern.
func (mux *ExchangeServeMux) Add(method, pattern, address string) {
	mux.rw.Lock()
	defer mux.rw.Unlock()

	handlers, ok := mux.routes[method]
	if !ok {
		handlers = make([]*patternHandler, 0)
	}

	addresses := []string{address}
	handler := patternHandler{pattern: pattern, addresses: addresses}
	mux.routes[method] = append(handlers, &handler)
}

// Remove unregisters the address for a backend service as a handler for an
// HTTP method and URL pattern.
func (mux *ExchangeServeMux) Remove(method, pattern, address string) {
	mux.rw.Lock()
	defer mux.rw.Unlock()
}

// ServeHTTP dispatches the request to the backend service whose pattern most
// closely matches the request URL.
func (mux *ExchangeServeMux) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	mux.rw.RLock()
	defer mux.rw.RUnlock()

	// Attempt to match the request against registered patterns and addresses.
	addresses, err := mux.match(request)
	if err != nil {
		writer.WriteHeader(http.StatusNotFound)
		return
	}

	// Make a request to a random backend service.
	index := rand.Intn(len(*addresses))
	address := (*addresses)[index]
	url := address + request.URL.Path
	if len(request.URL.Query()) > 0 {
		url = url + "?" + request.URL.RawQuery
	}
	innerRequest, err := http.NewRequest(request.Method, url, request.Body)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	for header, values := range request.Header {
		for _, value := range values {
			innerRequest.Header.Add(header, value)
		}
	}
	response, err := http.DefaultClient.Do(innerRequest)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Relay the response from the backend service back to the client.
	for header, values := range response.Header {
		for _, value := range values {
			writer.Header().Add(header, value)
		}
	}
	writer.WriteHeader(response.StatusCode)
	body := bytes.NewBufferString("")
	body.ReadFrom(response.Body)
	writer.Write(body.Bytes())
}

// Match finds backend service addresses capable of handling the request.
func (mux *ExchangeServeMux) match(request *http.Request) (*[]string, error) {
	handlers, ok := mux.routes[request.Method]
	if ok {
		for _, handler := range handlers {
			if handler.match(request.URL.Path) {
				return &handler.addresses, nil
			}
		}
	}
	return nil, errors.New("No matching address")
}

// Handler keeps track of backend service addresses that are registered to
// handle a URL pattern.
type patternHandler struct {
	pattern   string
	addresses []string
}

// Match returns true if this handler is a match for path.
func (handler *patternHandler) match(path string) bool {
	var i, j int
	for i < len(path) {
		switch {
		case j == len(handler.pattern) && handler.pattern[j-1] == '/':
			return true
		case j >= len(handler.pattern):
			return false
		case handler.pattern[j] == ':':
			j = handler.find(handler.pattern, '/', j)
			i = handler.find(path, '/', i)
		case path[i] == handler.pattern[j]:
			i++
			j++
		default:
			return false
		}
	}
	if j != len(handler.pattern) {
		return false
	}
	return true
}

// Find searches text for char, starting at startIndex, and returns the index
// of the next instance of char.  startIndex is returned if no instance of
// char is found.
func (handler *patternHandler) find(text string, char byte, startIndex int) int {
	j := startIndex
	for j < len(text) && text[j] != char {
		j++
	}
	return j
}
