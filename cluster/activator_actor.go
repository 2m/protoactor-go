package cluster

import (
	"log"
	"math/rand"

	"github.com/AsynkronIT/gam/actor"
)

var (
	activatorPid = actor.SpawnNamed(actor.FromProducer(newActivatorActor()), "activator")
)

type activator struct {
}

func activatorForHost(host string) *actor.PID {
	pid := actor.NewPID(host, "activator")
	return pid
}

func getRandomActivator() *actor.PID {
	r := rand.Int()
	members := list.Members()
	i := r % len(members)
	member := members[i]
	return activatorForHost(member.Name)
}

func newActivatorActor() actor.Producer {
	return func() actor.Actor {
		return &activator{}
	}
}

func (*activator) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	case *actor.Started:
		log.Println("[CLUSTER] Activator started")
	case *ActorPidRequest:
		props := nameLookup[msg.Kind]
		pid := actor.SpawnNamed(props, msg.Name)
		response := &ActorPidResponse{
			Pid: pid,
		}
		context.Respond(response)
	default:
		log.Printf("[CLUSTER] Activator got unknown message %+v", msg)
	}
}
