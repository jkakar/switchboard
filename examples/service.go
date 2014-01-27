package main

import (
	"bytes"
	"fmt"
	"github.com/coreos/go-etcd/etcd"
	"github.com/jkakar/switchboard"
	"log"
	"net/http"
	"os"
)

func main() {
	port := os.Getenv("PORT")
	client := etcd.NewClient([]string{"http://127.0.0.1:4001"})
	handler := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		buffer := bytes.NewBufferString(fmt.Sprintf("Hello from service on port %v.", port))
		writer.Write(buffer.Bytes())
	})
	service := sb.NewService("example", client, "json-schema", handler)
	routes := sb.Routes{"get": []string{"/hello"}}
	err := service.ListenAndServe(fmt.Sprintf(":%v", port), routes)
	if err != nil {
		log.Fatal(err)
	}
}
