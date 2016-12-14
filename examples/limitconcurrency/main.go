package main

import (
	"log"

	"github.com/AsynkronIT/gam/actor"
	"github.com/AsynkronIT/gam/routing"
	"github.com/AsynkronIT/goconsole"
)

type myMessage struct{ i int }

const maxConcurrency = 5

func doWork(ctx actor.Context) {
	if msg, ok := ctx.Message().(*myMessage); ok {
		//this is guaranteed to only execute with a max concurrency level of `maxConcurrency`
		log.Printf("%v got message %d", ctx.Self(), msg.i)
	}
}

func main() {
	pid := actor.Spawn(actor.FromFunc(doWork).WithPoolRouter(routing.NewRoundRobinPool(maxConcurrency)))
	for i := 0; i < 1000; i++ {
		pid.Tell(&myMessage{i})
	}
	console.ReadLine()
}
