package main

import (
	"log"
	"runtime"

	"sync"

	"fmt"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/AsynkronIT/protoactor-go/examples/remoterouting/messages"
	"github.com/AsynkronIT/protoactor-go/process"
	"github.com/AsynkronIT/protoactor-go/remoting"
	"github.com/AsynkronIT/protoactor-go/routing"

	console "github.com/AsynkronIT/goconsole"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	runtime.GC()

	remoting.Start("127.0.0.1:8100")

	p1 := process.NewPID("127.0.0.1:8101", "remote")
	p2 := process.NewPID("127.0.0.1:8102", "remote")
	remote := routing.SpawnGroup(routing.NewConsistentHashGroup(p1, p2))

	messageCount := 1000000

	var wgStop sync.WaitGroup

	props := actor.
		FromProducer(newLocalActor(&wgStop, messageCount)).
		WithMailbox(actor.NewBoundedMailbox(10000))

	pid := actor.Spawn(props)

	log.Println("Starting to send")

	t := time.Now()

	for i := 0; i < messageCount; i++ {
		message := &messages.Ping{User: fmt.Sprintf("User_%d", i)}
		remote.Request(message, pid)
	}

	wgStop.Wait()

	actor.StopActor(pid)

	fmt.Printf("elapsed: %v\n", time.Since(t))

	console.ReadLine()
}
