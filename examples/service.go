package main

import (
	"io"
	"log"
	"net/http"
	"os"

	"github.com/coreos/go-etcd/etcd"
	"github.com/jkakar/switchboard"
	"github.com/pat-go/pat.go"
)

func main() {
	// Initialize the service.
	port := os.Getenv("PORT")
	address := "http://localhost:" + port
	client := etcd.NewClient([]string{"http://127.0.0.1:4001"})
	routes := switchboard.Routes{"GET": []string{"/hello/:name"}}
	handler := pat.New()
	handler.Get("/hello/:name", http.HandlerFunc(hello))
	service := switchboard.NewService("example", client, address, routes, handler)

	// Broadcast service presence to etcd.
	go func() {
		stop := make(chan bool)
		service.Broadcast(5, 10, stop)
	}()

	// Listen for HTTP requests from the exchange.
	err := http.ListenAndServe("localhost:"+port, handler)
	if err != nil {
		log.Fatal(err)
	}
}

func hello(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get(":name")
	io.WriteString(w, "Hello, "+name)
	log.Printf("Responding to /hello/" + name)
}
