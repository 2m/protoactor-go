package process

import (
	"errors"
	"log"
	"sync"
	"time"
)

var (
	ErrTimeout = errors.New("timeout")
)

func NewFuture(timeout time.Duration) *Future {
	fut := &Future{cond: sync.NewCond(&sync.Mutex{})}

	ref := &futureProcess{f: fut}
	id := Registry.NextId()

	pid, ok := Registry.Add(ref, id)
	if !ok {
		log.Printf("[ACTOR] Failed to register future actorref '%v'", id)
		log.Println(id)
	}

	fut.pid = pid
	fut.t = time.AfterFunc(timeout, func() {
		fut.err = ErrTimeout
		ref.Stop(pid)
	})

	return fut
}

type Future struct {
	pid  *ID
	cond *sync.Cond
	// protected by cond
	done   bool
	result interface{}
	err    error
	t      *time.Timer
}

// PID to the backing actor for the Future result
func (f *Future) PID() *ID {
	return f.pid
}

// PipeTo starts a go routine and waits for the `Future.Result()`, then sends the result to the given `PID`
func (f *Future) PipeTo(pid *ID) {
	go func() {
		res, err := f.Result()
		if err != nil {
			pid.Tell(err)
		} else {
			pid.Tell(res)
		}
	}()
}

func (f *Future) ContinueWith(fun func(f *Future)) {
	go func() {
		f.Wait()
		fun(f)
	}()
}

func (f *Future) wait() {
	f.cond.L.Lock()
	for !f.done {
		f.cond.Wait()
	}
	f.cond.L.Unlock()
}

func (f *Future) Result() (interface{}, error) {
	f.wait()
	return f.result, f.err
}

func (f *Future) Wait() error {
	f.wait()
	return f.err
}

// futureProcess is a struct carrying a response PID and a channel where the response is placed
type futureProcess struct {
	f *Future
}

func (ref *futureProcess) SendUserMessage(pid *ID, message interface{}, sender *ID) {
	ref.f.result = message
	ref.Stop(pid)
}

func (ref *futureProcess) SendSystemMessage(pid *ID, message SystemMessage) {
	ref.f.result = message
	ref.Stop(pid)
}

func (ref *futureProcess) Stop(pid *ID) {
	ref.f.cond.L.Lock()
	if ref.f.done {
		ref.f.cond.L.Unlock()
		return
	}

	ref.f.done = true
	ref.f.t.Stop()
	Registry.Remove(pid)

	ref.f.cond.L.Unlock()
	ref.f.cond.Signal()
}
