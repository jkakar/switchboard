package sb

import (
	"bytes"
	"code.google.com/p/go-uuid/uuid"
	"encoding/json"
	"github.com/coreos/go-etcd/etcd"
	// "log"
	"net/http"
	// "os"
	// "os/signal"
	// "time"
)

// Map HTTP methods to URLs.
type Routes map[string][]string

// The representation of a service stored in etcd and used by exchanges.
type ServiceRecord struct {
	Id      string `json:"id"`
	Address string `json:"address"`
	Routes  Routes `json:"routes"`
}

// Service responds to HTTP requests for a set of endpoints described by a
// JSON schema.
type Service struct {
	id        string       // A unique ID representing this service.
	namespace string       // The root directory in etcd for config files.
	client    *etcd.Client // The etcd client.
	address   string       // The public address for this service.
	routes    Routes       // The routes handled by this service.
	handler   http.Handler // An HTTP handler compatible with the JSON schema.
}

// NewService creates a service that can be registered with etcd to handle
// requests from an exchange.
func NewService(namespace string, client *etcd.Client, address string, routes Routes, handler http.Handler) *Service {
	return &Service{
		id:        uuid.NewRandom().String(),
		namespace: namespace,
		client:    client,
		address:   address,
		routes:    routes,
		handler:   handler}
}

// Id returns the UUID that identifies this service.
func (service *Service) Id() string {
	return service.id
}

// Notify registers this service with etcd.
func (service *Service) Notify() (*ServiceRecord, error) {
	key := service.namespace + "/" + service.id
	record := ServiceRecord{
		Id:      service.id,
		Address: service.address,
		Routes:  service.routes}
	recordJSON, err := json.Marshal(record)
	if err != nil {
		return nil, err
	}
	value := bytes.NewBuffer(recordJSON).String()
	_, err = service.client.Set(key, value, 0)
	if err != nil {
		return nil, err
	}
	return &record, nil
}

// // Register the service and listen for connections.
// func (service *Service) ListenAndServe(address string, routes Routes) error {
// 	// Listen on address and wait a second to see if an error occurs.
// 	// TODO(jkakar): This is pretty craptastic.
// 	listenResult := make(chan error)
// 	go func() {
// 		listenResult <- http.ListenAndServe(address, service.handler)
// 	}()
// 	select {
// 	case err := <-listenResult:
// 		return err
// 	case <-time.After(time.Second * 1):
// 	}

// 	// Assume that we're listening correctly and setup the service in etcd.
// 	key := service.namespace + "/" + service.id
// 	entry := ServiceRecord{Address: address, Routes: routes}
// 	entryJSON, err := json.Marshal(entry)
// 	if err != nil {
// 		log.Panic(err)
// 	}
// 	value := bytes.NewBuffer(entryJSON).String()
// 	_, err = service.client.Set(key, value, 0)
// 	if err != nil {
// 		log.Panic(err)
// 	}
// 	log.Printf("Registered service with key: %v", key)

// 	// Remove the key when the service receives a SIGTERM and shuts down.
// 	interrupt := make(chan os.Signal, 1)
// 	signal.Notify(interrupt, os.Interrupt)
// 	go func() {
// 		for sig := range interrupt {
// 			service.client.Delete(key, false)
// 			log.Fatal(sig)
// 		}
// 	}()

// 	// Wait forever.
// 	select {}
// 	return nil
// }
