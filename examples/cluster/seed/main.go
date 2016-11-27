package main

import (
	"log"
	"time"

	"github.com/AsynkronIT/gam/cluster"
	"github.com/AsynkronIT/gam/examples/cluster/shared"
	console "github.com/AsynkronIT/goconsole"
)

const (
	timeout = 1 * time.Second
)

func main() {
	cluster.Start("127.0.0.1:7711")
	hello := shared.GetHelloGrain("abc")

	res, err := hello.SayHello(&shared.HelloRequest{Name: "Roger"}, timeout)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Message from grain %v", res.Message)
	console.ReadLine()
}

