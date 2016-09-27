package main

import (
	"log"

	//"github.com/gogo/protobuf/proto"

	"github.com/AsynkronIT/gam/actor"
	"github.com/AsynkronIT/gam/couchbase_persistence"
	"github.com/AsynkronIT/gam/examples/persistence/messages"
	"github.com/AsynkronIT/gam/persistence"
	"github.com/AsynkronIT/goconsole"
)

type persistentActor struct {
	name  string
	items []string
}

//CQRS style messages
func (state *persistentActor) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	case *messages.RenameCommand: //command handler, you can have side effects here
		event := &messages.RenamedEvent{Name: msg.Name}
		log.Printf("Rename %v\n", msg.Name)
		context.Receive(event)
	case *messages.RenamedEvent: //event handler, only mutate state here
		state.name = msg.Name
	case *messages.AddItemCommand:
		event := &messages.AddedItemEvent{Item: msg.Item}
		log.Printf("Add item %v", msg.Item)
		context.Receive(event)
	case *messages.AddedItemEvent:
		state.items = append(state.items, msg.Item)
	case *messages.DumpCommand: //just so we can manually trigger a console dump of state
		log.Printf("%+v", state)
	case *persistence.ReplayComplete: //will be triggered once the persistence plugin have replayed all events
		log.Println("Replay Complete")
		context.Receive(&messages.DumpCommand{})
	}
}

func newPersistentActor() actor.Actor {
	return &persistentActor{
		name: "Initial Name",
	}
}

func main() {

	props := actor.
		FromProducer(newPersistentActor).
		WithReceivers(persistence.Using(couchbase_persistence.New("labb", "couchbase://localhost")))

	pid := actor.Spawn(props)
	pid.Tell(&messages.AddItemCommand{Item: "Banana"})
	pid.Tell(&messages.AddItemCommand{Item: "Apple"})
	pid.Tell(&messages.AddItemCommand{Item: "Orange"})
	pid.Tell(&messages.RenameCommand{Name: "Acme Inc"})
	pid.Tell(&messages.DumpCommand{})
	pid.Tell(&actor.PoisonPill{})
	pid.Tell(&messages.DumpCommand{})
	console.ReadLine()
}
