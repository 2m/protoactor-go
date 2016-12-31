package main

import (
	"flag"
	"runtime"

	"log"

	"github.com/AsynkronIT/gam/languages/golang/examples/remoterouting/messages"
	"github.com/AsynkronIT/gam/languages/golang/src/actor"
	"github.com/AsynkronIT/gam/languages/golang/src/remoting"
	console "github.com/AsynkronIT/goconsole"
)

var (
	flagBind = flag.String("bind", "localhost:8100", "Bind to address")
	flagName = flag.String("name", "node1", "Name")
)

type remoteActor struct {
	name  string
	count int
}

func (a *remoteActor) Receive(context actor.Context) {
	switch context.Message().(type) {
	case *messages.Ping:
		context.Respond(&messages.Pong{})
	}
}

func newRemoteActor(name string) actor.Producer {
	return func() actor.Actor {
		return &remoteActor{
			name: name,
		}
	}
}

func NewRemote(bind, name string) {
	remoting.Start(bind)
	props := actor.
		FromProducer(newRemoteActor(name)).
		WithMailbox(actor.NewBoundedMailbox(10000))

	actor.SpawnNamed(props, "remote")

	log.Println(name, "Ready")
}

func main() {

	runtime.GOMAXPROCS(runtime.NumCPU())
	runtime.GC()

	flag.Parse()

	NewRemote(*flagBind, *flagName)

	console.ReadLine()
}
