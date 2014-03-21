// Package switchboard composes an HTTP API out of a collection of HTTP
// service backends.
//
// Backend services register themselves with frontend exchanges by pushing
// their configuration into etcd.  This configuration describes HTTP methods
// and URL patterns that the service is capable of responding to.  Exchanges
// respond to API consumers and route their requests to service backends.
// They watch for configuration changes in etcd and update their routing rules
// to add and remove routes as services come and go.
package switchboard
