# Switchboard

Compose an HTTP API out of a collection of HTTP service backends.

Backend services register themselves with frontend exchanges by pushing their
configuration into etcd.  This configuration describes HTTP methods and URL
patterns that the service is capable of responding to.  Exchanges respond to
API consumers and route their requests to service backends.  They watch for
configuration changes in etcd and update their routing rules to add and remove
routes as services come and go.

![Switchboard operators](http://newdeal.feri.org/images/ae15.gif)

## Install

```
go get github.com/jkakar/switchboard
```

API documentation is at [godoc.org/github.com/jkakar/switchboard](http://godoc.org/github.com/jkakar/switchboard).

## Example

The `examples/exchange.go` and `examples/service.go` programs demonstrate
Switchboard.  You need to run [etcd](https://github.com/coreos/etcd) on
`http://127.0.0.1:4001` before running the example programs.  If you're using
a git checkout you can build and run etcd with `./build && bin/etcd -v`.
Next, start the exchange and a handful of service processes in their own
terminals.

```bash
PORT=5000 go run examples/exchange.go
PORT=5001 go run examples/service.go
PORT=5002 go run examples/service.go
PORT=5003 go run examples/service.go
```

Now you can make requests and watch the exchange pass them to service
backends:

```bash
curl http://localhost:5000/hello/jane
```

## License

Copyright 2014, [Jamshed Kakar <jkakar@kakar.ca>](mailto:jkakar@kakar.ca)

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at:

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
