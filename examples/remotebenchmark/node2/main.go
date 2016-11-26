package main

import (
	"log"
	"runtime"

	"github.com/AsynkronIT/gam/actor"
	"github.com/AsynkronIT/gam/examples/remotebenchmark/messages"
	"github.com/AsynkronIT/gam/remoting"
	"github.com/AsynkronIT/goconsole"
)

type remoteActor struct{}

func (*remoteActor) Receive(context actor.Context) {
	switch context.Message().(type) {
	case *messages.StartRemote:
		log.Println("Starting")
		context.Respond(&messages.Start{})
	case *messages.Ping:
		context.Respond(&messages.Pong{})
	}
}

func newRemoteActor() actor.Producer {
	return func() actor.Actor {
		return &remoteActor{}
	}
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU() * 1)
	runtime.GC()

	remoting.Start("127.0.0.1:8080")
	props := actor.
		FromProducer(newRemoteActor()).
		WithMailbox(actor.NewBoundedMailbox(1000, 1000))

	actor.SpawnNamed(props, "remote")

	console.ReadLine()
}
