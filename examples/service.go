package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/coreos/go-etcd/etcd"
	"github.com/jkakar/switchboard"
	"github.com/pat-go/pat.go"
)

func main() {
	// Setup HTTP routes and initialize the service.
	port := os.Getenv("PORT")
	address := "http://localhost:" + port
	client := etcd.NewClient([]string{"http://127.0.0.1:4001"})
	routes := switchboard.Routes{"GET": []string{"/hello/:name"}}
	service := switchboard.NewService("example", client, address, routes)

	// Broadcast service presence to etcd every 5 seconds (with a TTL of 10
	// seconds).
	go func() {
		log.Print("Broadcasting service configuration to etcd")
		stop := make(chan bool)
		service.Broadcast(5, 10, stop)
	}()

	// Unregister the service when we receive a SIGTERM and shuts down.
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	go func() {
		for sig := range interrupt {
			log.Print("Unregistering service")
			service.Unregister()
			log.Fatal(sig)
		}
	}()

	// Setup a handler to process requests for our registered routes and
	// listen for HTTP requests from the exchange.
	handler := pat.New()
	handler.Get("/hello/:name", Log(http.HandlerFunc(Hello)))
	log.Printf("Listening for HTTP requests on port %v", port)
	err := http.ListenAndServe("localhost:"+port, handler)
	if err != nil {
		log.Print(err)
	}

	// Unregister the service because we couldn't listen for requests.
	log.Print("Unregistering service")
	service.Unregister()
	log.Print("Shutting down")
}

func Hello(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get(":name")
	io.WriteString(w, "Hello, "+name)
}

func Log(handler http.Handler) http.Handler {
	wrapper := func(writer http.ResponseWriter, request *http.Request) {
		log.Printf("%s %s %s", request.RemoteAddr, request.Method, request.URL.Path)
		handler.ServeHTTP(writer, request)
	}
	return http.HandlerFunc(wrapper)
}
