package switchboard

import (
	"bytes"
	"math/rand"
	"net/http"
	"sync"
)

// addressArray is a set of addresses that may satisfy a particular pattern.
type addressArray []string

// patternMap maps URL patterns to their addresses.
type patternMap map[string]addressArray

// methodMap maps HTTP methods to their URL patterns.
type methodMap map[string]patternMap

// ExchangeServeMux is an HTTP request multiplexer.  It matches the URL of
// each incoming request against a list of registered patterns to find the
// service that can respond to it and proxies the request to the appropriate
// backend.
type ExchangeServeMux struct {
	rw     sync.RWMutex // Synchronize access to routes map.
	routes methodMap    // Map of methods and patterns to backend services.
}

// NewExchangeServeMux allocates and returns a new ExchangeServeMux.
func NewExchangeServeMux() *ExchangeServeMux {
	return &ExchangeServeMux{routes: make(methodMap)}
}

// Add registers the address as a backend service for the given HTTP method
// and URL pattern.
func (mux *ExchangeServeMux) Add(method string, pattern string, address string) {
	mux.rw.Lock()
	defer mux.rw.Unlock()

	patterns, ok := mux.routes[method]
	if !ok {
		patterns = make(patternMap)
		mux.routes[method] = patterns
	}
	addresses, ok := patterns[pattern]
	if !ok {
		addresses = addressArray{}
		patterns[pattern] = addresses
	}
	patterns[pattern] = append(addresses, address)
}

// Add deregisters the address as a backend service for the given HTTP method
// and URL pattern.
func (mux *ExchangeServeMux) Remove(method string, pattern string, address string) {
	mux.rw.Lock()
	defer mux.rw.Unlock()
}

// ServeHTTP dispatches the request to the backend service whose pattern most
// closely matches the request URL.
func (mux *ExchangeServeMux) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	mux.rw.RLock()
	defer mux.rw.RUnlock()

	patterns, ok := mux.routes[request.Method]
	if ok {
		addresses, ok := patterns[request.URL.Path]
		if ok {
			// Make a request to a random backend service.
			index := rand.Intn(len(addresses))
			address := addresses[index]
			url := address + request.URL.Path
			if len(request.URL.Query()) > 0 {
				url = url + "?" + request.URL.Query().Encode()
			}
			innerRequest, err := http.NewRequest(request.Method, url, request.Body)
			if err != nil {
				writer.WriteHeader(http.StatusInternalServerError)
				return
			}
			response, err := http.DefaultClient.Do(innerRequest)
			if err != nil {
				writer.WriteHeader(http.StatusInternalServerError)
				return
			}

			// Relay the response from the backend service back to the client.
			writer.WriteHeader(response.StatusCode)
			body := bytes.NewBufferString("")
			body.ReadFrom(response.Body)
			writer.Write(body.Bytes())
			return
		}
	}

	writer.WriteHeader(http.StatusNotFound)
}
