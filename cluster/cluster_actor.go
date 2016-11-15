package cluster

import (
	"log"
	"time"

	"github.com/AsynkronIT/gam/actor"
	"github.com/AsynkronIT/gam/cluster/messages"
)

var clusterPid = actor.SpawnNamed(actor.FromProducer(newClusterActor()), "cluster")

func newClusterActor() actor.ActorProducer {
	return func() actor.Actor {
		return &clusterActor{
			partition: make(map[string]*actor.PID),
		}
	}
}

type clusterActor struct {
	partition map[string]*actor.PID
}

func (state *clusterActor) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	case *actor.Started:
		log.Println("Cluster actor started")
	case *messages.ActorPidRequest:
		state.actorPidRequest(msg)
	case *clusterStatusJoin:
		state.clusterStatusJoin(msg)
	case *clusterStatusLeave:
		log.Printf("[STATUS] Node left %v", msg.node.Name)
	case *messages.TakeOwnership:
		log.Printf("Took ownerhip of %v", msg.Pid)
		state.partition[msg.Id] = msg.Pid
	default:
		log.Printf("Cluster got unknown message %+v", msg)
	}
}

func (state *clusterActor) actorPidRequest(msg *messages.ActorPidRequest) {
	pid := state.partition[msg.Id]
	if pid == nil {

		x, resp := actor.RequestResponsePID()
		//get a random node
		random := getRandom()

		//send request
		random.Tell(&messages.ActorActivateRequest{
			Id:     msg.Id,
			Kind:   msg.Kind,
			Sender: x,
		})

		tmp, _ := resp.ResultOrTimeout(5 * time.Second)
		typed := tmp.(*messages.ActorActivateResponse)
		pid = typed.Pid
		state.partition[msg.Id] = pid
	}
	response := &messages.ActorPidResponse{
		Pid: pid,
	}
	msg.Sender.Tell(response)
}

func (state *clusterActor) clusterStatusJoin(msg *clusterStatusJoin) {
	log.Printf("[STATUS] Node joined %v", msg.node.Name)
	selfName := list.LocalNode().Name
	for key := range state.partition {
		c := findClosest(key)
		if c.Name != selfName {
			log.Printf("Node %v should take ownership of %v", c.Name, key)
			pid := state.partition[key]
			owner := clusterForNode(c)
			owner.Tell(&messages.TakeOwnership{
				Pid: pid,
				Id:  key,
			})
		}
	}
}
