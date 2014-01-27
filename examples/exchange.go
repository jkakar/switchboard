package main

import (
	"github.com/jkakar/switchboard"
	"log"
)

func main() {
	client := etcd.NewClient([]string{"http://127.0.0.1:4001"})
	exchange := sb.NewExchange(client)
	err := exchange.ListenAndServe(":9000")
	if err != nil {
		log.Fatal(err)
	}
}
