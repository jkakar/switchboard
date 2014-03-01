package sb

import (
	"bytes"
	"encoding/json"
	"github.com/coreos/go-etcd/etcd"
)

// ServiceManifest is a representation of services registered with an exchange.
type ServiceManifest struct {
	Services []*ServiceRecord
	index    uint64
}

// Exchange responds to HTTP requests and proxies them to services that are
// capable to responding to them.
type Exchange struct {
	namespace string
	client    *etcd.Client
}

// NewExchange creates an exchange that can fetch information about and watch
// for services changes in etcd.
func NewExchange(namespace string, client *etcd.Client) *Exchange {
	return &Exchange{namespace: namespace, client: client}
}

// ServiceManifest returns information from etcd about the currently
// registered services.
func (exchange *Exchange) ServiceManifest() (*ServiceManifest, error) {
	sort := false
	recursive := true
	response, err := exchange.client.Get(exchange.namespace, sort, recursive)
	if err != nil {
		return nil, err
	}

	return exchange.buildManifest(response), nil
}

// buildManifest reads a response from etcd and converts it to a service
// manifest.
func (exchange *Exchange) buildManifest(response *etcd.Response) *ServiceManifest {
	serviceRecords := []*ServiceRecord{}
	for _, node := range response.Node.Nodes {
		var serviceRecord ServiceRecord
		json.Unmarshal(bytes.NewBufferString(node.Value).Bytes(), &serviceRecord)
		serviceRecords = append(serviceRecords, &serviceRecord)
	}
	return &ServiceManifest{
		Services: serviceRecords,
		index:    response.EtcdIndex}
}

// Watch for updates in etcd and send new service manifests to the watcher
// channel.  Send on the stop channel to stop watching.
func (exchange *Exchange) Watch(watcher chan *ServiceManifest, stop chan bool) (err error) {
	receiver := make(chan *etcd.Response)
	stopped := make(chan bool)
	go func() {
		_, err = exchange.client.Watch(exchange.namespace, 0, true, receiver, stop)
		stopped <- true
	}()
	select {
	case response := <-receiver:
		watcher <- exchange.buildManifest(response)
	case <-stopped:
		return nil
	}
	return err
}
