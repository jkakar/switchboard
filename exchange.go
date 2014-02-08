package sb

import (
	"bytes"
	"encoding/json"
	"github.com/coreos/go-etcd/etcd"
)

// A representation of services registered with an exchange.
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

// ServiceManifest returns information about the currently registered
// services.
func (exchange *Exchange) ServiceManifest() (*ServiceManifest, error) {
	sort := false
	recursive := true
	response, err := exchange.client.Get(exchange.namespace, sort, recursive)
	if err != nil {
		return nil, err
	}

	serviceRecords := []*ServiceRecord{}
	for _, node := range response.Node.Nodes {
		var serviceRecord ServiceRecord
		json.Unmarshal(bytes.NewBufferString(node.Value).Bytes(), &serviceRecord)
		serviceRecords = append(serviceRecords, &serviceRecord)
	}
	manifest := &ServiceManifest{
		Services: serviceRecords,
		index:    response.EtcdIndex}
	return manifest, nil
}

// Watch for service changes.
func (exchange *Exchange) Watch(stop chan bool) error {
	return nil
}
