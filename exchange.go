package switchboard

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/coreos/go-etcd/etcd"
)

// Exchange watches for service changes in etcd and update an
// ExchangeServeMux.
type Exchange struct {
	namespace string            // The root directory in etcd for config files.
	client    *etcd.Client      // The etcd client.
	mux       *ExchangeServeMux // The serve mux to keep in sync with etcd.
	waitIndex uint64            // Wait index to use when watching etcd.
}

// NewExchange creates a new exchange configured to watch for changes in a
// given etcd directory.
func NewExchange(namespace string, client *etcd.Client, mux *ExchangeServeMux) *Exchange {
	return &Exchange{namespace: namespace, client: client, mux: mux}
}

// Init fetches service information from etcd and initializes the exchange.
func (exchange *Exchange) Init() error {
	sort := false
	recursive := true
	response, err := exchange.client.Get(exchange.namespace, sort, recursive)
	if err != nil {
		// TODO(jkakar) We probably want to create a missing namespace if one
		// doesn't already exist.
		return err
	}

	for _, node := range response.Node.Nodes {
		service := exchange.load(&node)
		exchange.Register(service)
	}

	// We want to watch changes *after* this one.
	exchange.waitIndex = response.EtcdIndex + 1
	return nil
}

func (exchange *Exchange) load(node *etcd.Node) *ServiceRecord {
	var service ServiceRecord
	// TODO(jkakar) Check for errors.
	json.Unmarshal(bytes.NewBufferString(node.Value).Bytes(), &service)
	return &service
}

// Watch observes changes in etcd and registers and unregisters services, as
// necessary, with the ExchangeServeMux.  This blocking call will terminate
// when a value is received on the stop channel.
func (exchange *Exchange) Watch(stop chan bool) {
	receiver := make(chan *etcd.Response)
	stopped := make(chan bool)
	go func() {
		recursive := true
		// TODO(jkakar) Check for errors.
		exchange.client.Watch(exchange.namespace, exchange.waitIndex, recursive, receiver, stop)
		stopped <- true
	}()

	for {
		select {
		case response := <-receiver:
			fmt.Printf("index:  %v\n", response.EtcdIndex)
			fmt.Printf("nindex: %v\n", response.Node.ModifiedIndex)
			fmt.Printf("action: %v\n", response.Action)
			fmt.Printf("value:  %v\n", response.Node.Value)
			fmt.Printf("key:    %v\n", response.Node.Key)
			fmt.Printf("nodes:  %v\n", response.Node.Nodes)
			fmt.Print("\n")
			if response.Action == "set" {
				service := exchange.load(response.Node)
				exchange.Register(service)
			} else if response.Action == "delete" {
				service := exchange.load(response.Node)
				exchange.Unregister(service)
			}
		case <-stopped:
			return
		}
	}
}

// Register adds routes exposed by a service to the ExchangeServeMux.
func (exchange *Exchange) Register(service *ServiceRecord) {
	for method, patterns := range service.Routes {
		for _, pattern := range patterns {
			exchange.mux.Add(method, pattern, service.Address)
		}
	}
}

// Unregister removes routes exposed by a service from the ExchangeServeMux.
func (exchange *Exchange) Unregister(service *ServiceRecord) {
	for method, patterns := range service.Routes {
		for _, pattern := range patterns {
			exchange.mux.Remove(method, pattern, service.Address)
		}
	}
}
