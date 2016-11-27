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
	cluster.Start("127.0.0.1:0", "127.0.0.1:7711")
	sync()
	async()

	console.ReadLine()
}

func sync() {
	hello := shared.GetHelloGrain("abc")
	res, err := hello.SayHello(&shared.HelloRequest{Name: "GAM"}, timeout)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Message from SayHello: %v", res.Message)
}

func async() {
	hello := shared.GetHelloGrain("abc")
	c, e := hello.AddChan2(&shared.AddRequest{A: 123, B: 456}, timeout)
	t := time.NewTicker(100 * time.Millisecond)

	for {
		select {
		case <-t.C:
			log.Println("Tick..") //this might not happen if res returns fast enough
		case err := <-e:
			log.Fatal(err)
		case res := <-c:
			log.Printf("Result is %v", res.Result)
			return
		}
	}
}
