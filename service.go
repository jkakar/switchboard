package sb

import (
	"bytes"
	"code.google.com/p/go-uuid/uuid"
	"encoding/json"
	"github.com/coreos/go-etcd/etcd"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

// Map HTTP methods to URLs.
type Routes map[string][]string

// Service responds to HTTP requests for a set of endpoints described by a
// JSON schema.
type Service struct {
	id        string       // A unique ID representing this service.
	namespace string       // The root directory in etcd for config files.
	client    *etcd.Client // The etcd client.
	schema    string       // A JSON schema describing the service API.
	handler   http.Handler // An HTTP handler compatible with the JSON schema.
}

// NewService creates a service instance to represent an API node.  The JSON
// schema is used by exchanges to determine the kinds of requests the service
// is capable of handling.  The handler must be able to handle requests to
// links defined in the schema.
func NewService(namespace string, client *etcd.Client, schema string, handler http.Handler) *Service {
	return &Service{
		id:        uuid.NewRandom().String(),
		namespace: namespace,
		client:    client,
		schema:    schema,
		handler:   handler}
}

// Register the service and listen for connections.
func (service *Service) ListenAndServe(address string, routes Routes) error {
	// Listen on address and wait a second to see if an error occurs.
	// TODO(jkakar): This is pretty craptastic.
	listenResult := make(chan error)
	go func() {
		listenResult <- http.ListenAndServe(address, service.handler)
	}()
	select {
	case err := <-listenResult:
		return err
	case <-time.After(time.Second * 1):
	}

	// Assume that we're listening correctly and setup the service in etcd.
	key := service.namespace + "/" + service.id
	entry := serviceEntry{Address: address, Routes: routes}
	entryJSON, err := json.Marshal(entry)
	if err != nil {
		log.Panic(err)
	}
	value := bytes.NewBuffer(entryJSON).String()
	_, err = service.client.Set(key, value, 0)
	if err != nil {
		log.Panic(err)
	}
	log.Printf("Registered service with key: %v", key)

	// Remove the key when the service receives a SIGTERM and shuts down.
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	go func() {
		for sig := range interrupt {
			service.client.Delete(key, false)
			log.Fatal(sig)
		}
	}()

	// Wait forever.
	<-make(chan bool)
	return nil
}

// Capture service routing information for use by an exchange.
type serviceEntry struct {
	Address string `json:"address"`
	Routes  Routes `json:"routes"`
}
