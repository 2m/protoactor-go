package actor

import (
	"fmt"
	"log"
	"sync"
	"time"
)

func NewFuture(timeout time.Duration) *Future {
	ref := &FutureActorRef{
		channel: make(chan interface{}),
	}
	id := ProcessRegistry.getAutoId()

	pid, ok := ProcessRegistry.add(ref, id)

	if !ok {
		log.Printf("[ACTOR] Failed to register future actorref '%v'", id)
		log.Println(id)
	}
	ref.pid = pid

	fut := &Future{
		ref:     ref,
		timeout: timeout,
	}
	fut.wg.Add(1)
	go func() {
		select {
		case res := <-fut.ref.channel:
			fut.result = res
		case <-time.After(fut.timeout):
			fut.err = fmt.Errorf("Timeout")
		}
		fut.wg.Done()
		fut.ref.Stop(fut.PID())
	}()

	return fut
}

type Future struct {
	result  interface{}
	err     error
	wg      sync.WaitGroup
	ref     *FutureActorRef
	timeout time.Duration
}

//PID to the backing actor for the Future result
func (fut *Future) PID() *PID {
	return fut.ref.pid
}

//PipeTo starts a go routine and waits for the `Future.Result()`, then sends the result to the given `PID`
func (ref *Future) PipeTo(pid *PID) {
	go func() {
		res, err := ref.Result()
		if err != nil {
			pid.Tell(err)
		} else {
			pid.Tell(res)
		}
	}()
}

func (fut *Future) Result() (interface{}, error) {
	fut.wg.Wait()
	return fut.result, fut.err
}

func (fut *Future) Wait() error {
	fut.wg.Wait()
	return fut.err
}

//FutureActorRef is a struct carrying a response PID and a channel where the response is placed
type FutureActorRef struct {
	channel chan interface{}
	pid     *PID
}

func (ref *FutureActorRef) SendUserMessage(pid *PID, message interface{}, sender *PID) {
	ref.channel <- message
}

func (ref *FutureActorRef) SendSystemMessage(pid *PID, message SystemMessage) {
	ref.channel <- message
}

func (ref *FutureActorRef) Stop(pid *PID) {
	ProcessRegistry.remove(ref.pid)
	close(ref.channel)
}
func (ref *FutureActorRef) Watch(pid *PID)   {}
func (ref *FutureActorRef) UnWatch(pid *PID) {}
