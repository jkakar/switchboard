package switchboard

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"code.google.com/p/go-uuid/uuid"
	"github.com/coreos/go-etcd/etcd"
)

// Routes maps HTTP methods to URLs.
type Routes map[string][]string

// ServiceRecord is a representation of a service stored in etcd and used by
// exchanges.
type ServiceRecord struct {
	ID      string `json:"id"`
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

// ID returns the UUID that identifies this service.
func (service *Service) ID() string {
	return service.id
}

// Address returns the root URL configured for this service.
func (service *Service) Address() string {
	return service.address
}

// Routes returns the map of HTTP methods to URL patterns configured for this
// service.
func (service *Service) Routes() Routes {
	return service.routes
}

// Register adds a service record to etcd.  The ttl is the time to live for
// the service record, in seconds.  A ttl of 0 registers a service record that
// never expires.
func (service *Service) Register(ttl uint64) (*ServiceRecord, error) {
	key := service.namespace + "/" + service.id
	record := ServiceRecord{
		ID:      service.id,
		Address: service.address,
		Routes:  service.routes}
	recordJSON, err := json.Marshal(record)
	if err != nil {
		return nil, err
	}
	value := bytes.NewBuffer(recordJSON).String()
	_, err = service.client.Set(key, value, ttl)
	if err != nil {
		return nil, err
	}
	return &record, nil
}

// Unregister destroys the service record in etcd.  An error is returned if
// the service isn't registered.
func (service *Service) Unregister() error {
	key := service.namespace + "/" + service.id
	recursive := false
	_, err := service.client.Delete(key, recursive)
	return err
}

// Broadcast registers this service with etcd every interval seconds.  The ttl
// is the time to live for the service record, in seconds.  This blocking call
// will terminate when a value is received on the stop channel.
func (service *Service) Broadcast(interval uint64, ttl uint64, stop chan bool) {
	// TODO(jkakar) Check for errors.
	service.Register(ttl)
	for {
		select {
		case <-time.After(time.Duration(interval) * time.Second):
			// TODO(jkakar) Check for errors.
			service.Register(ttl)
		case <-stop:
			return
		}
	}
}
