package actor

import (
	"fmt"
	"log"
	"reflect"
	"time"
)

//Tell a message to a given PID
func (pid *PID) Tell(message interface{}) {
	ref, _ := ProcessRegistry.get(pid)
	ref.SendUserMessage(pid, message, nil)
}

//Ask a message to a given PID
func (pid *PID) TellWithSender(message interface{}, sender *PID) error {
	ref, _ := ProcessRegistry.get(pid)
	ref.SendUserMessage(pid, message, sender)
	return nil
}

//Ask a message to a given PID
func (pid *PID) AskFuture(message interface{}, timeout time.Duration) (*Future, error) {
	ref, found := ProcessRegistry.get(pid)
	if !found {
		return nil, fmt.Errorf("Unknown PID %s", pid)
	}
	future := NewFuture(timeout)
	ref.SendUserMessage(pid, message, future.PID())
	return future, nil
}

func (pid *PID) sendSystemMessage(message SystemMessage) {
	ref, _ := ProcessRegistry.get(pid)
	ref.SendSystemMessage(pid, message)
}

func (pid *PID) StopFuture() (*Future, error) {
	ref, found := ProcessRegistry.get(pid)

	if !found {
		return nil, fmt.Errorf("Unknown PID %s", pid)
	}
	future := NewFuture(10 * time.Second)

	ref, ok := ref.(*LocalActorRef)
	if !ok {
		log.Fatalf("[ACTOR] Trying to stop non local actorref %s", reflect.TypeOf(ref))
	}

	ref.Watch(future.PID())

	ref.Stop(pid)

	return future, nil
}

//Stop the given PID
func (pid *PID) Stop() {
	ref, _ := ProcessRegistry.get(pid)
	ref.Stop(pid)
}

func (pid *PID) suspend() {
	ref, _ := ProcessRegistry.get(pid)
	ref.(*LocalActorRef).Suspend()
}

func (pid *PID) resume() {
	ref, _ := ProcessRegistry.get(pid)
	ref.(*LocalActorRef).Resume()
}

//NewPID returns a new instance of the PID struct
func NewPID(host, id string) *PID {
	return &PID{
		Host: host,
		Id:   id,
	}
}

//NewLocalPID returns a new instance of the PID struct with the host preset
func NewLocalPID(id string) *PID {
	return &PID{
		Host: ProcessRegistry.Host,
		Id:   id,
	}
}
