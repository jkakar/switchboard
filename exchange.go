package sb

import (
	"github.com/coreos/go-etcd/etcd"
)

type Exchange struct {
	namespace string
	client    *etcd.Client
}

func NewExchange(namespace string, client *etcd.Client) *Exchange {
	return &Exchange{namespace: namespace, client: client}
}

func (exchange *Exchange) ListenAndServe(address string) error {
	return nil
}
