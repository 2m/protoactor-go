package cluster

import (
	"log"

	"github.com/AsynkronIT/gam/actor"
	"github.com/AsynkronIT/gam/cluster/messages"
)

type activator struct {
}

var activatorPid = actor.SpawnNamed(actor.FromProducer(newActivatorActor()), "activator")

func newActivatorActor() actor.ActorProducer {
	return func() actor.Actor {
		return &activator{}
	}
}

func (*activator) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	case *actor.Started:
		log.Println("[CLUSTER] Activator actor started")
	case *messages.ActorActivateRequest:
		log.Printf("[CLUSTER] Cluster actor creating %v of type %v", msg.Id, msg.Kind)
		props := nameLookup[msg.Kind]
		pid := actor.SpawnNamed(props, msg.Id)
		response := &messages.ActorActivateResponse{
			Pid: pid,
		}
		msg.Sender.Tell(response)
	default:
		log.Printf("[CLUSTER] Activator got unknown message %+v", msg)
	}
}
