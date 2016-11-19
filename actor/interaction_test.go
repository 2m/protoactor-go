package actor

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type DummyMessage struct{}
type BlackHoleActor struct{}

var testTimeout = 1 * time.Second

func (state *BlackHoleActor) Receive(context Context) {}

func NewBlackHoleActor() Actor {
	return &BlackHoleActor{}
}

func TestSpawnProducesActorRef(t *testing.T) {
	actor := Spawn(FromProducer(NewBlackHoleActor))
	defer actor.Stop()
	assert.NotNil(t, actor)
}

type EchoMessage struct{}

type EchoReplyMessage struct{}

type EchoActor struct{}

func NewEchoActor() Actor {
	return &EchoActor{}
}

func (*EchoActor) Receive(context Context) {
	switch context.Message().(type) {
	case EchoMessage:
		context.Sender().Tell(EchoReplyMessage{})
	}
}

func TestActorCanReplyToMessage(t *testing.T) {
	actor := Spawn(FromProducer(NewEchoActor))
	defer actor.Stop()
	result, _ := actor.AskFuture(EchoMessage{}, testTimeout)
	if _, err := result.Result(); err != nil {
		assert.Fail(t, "timed out")
		return
	}
}

type BecomeMessage struct{}

type EchoBecomeActor struct{}

func NewEchoBecomeActor() Actor {
	return &EchoBecomeActor{}
}

func (state *EchoBecomeActor) Receive(context Context) {
	switch context.Message().(type) {
	case BecomeMessage:
		context.Become(state.Other)
	}
}

func (EchoBecomeActor) Other(context Context) {
	switch context.Message().(type) {
	case EchoMessage:
		context.Sender().Tell(EchoReplyMessage{})
	}
}

func TestActorCanBecome(t *testing.T) {
	actor := Spawn(FromProducer(NewEchoBecomeActor))
	defer actor.Stop()
	actor.Tell(BecomeMessage{})
	result, _ := actor.AskFuture(EchoMessage{}, testTimeout)
	if _, err := result.Result(); err != nil {
		assert.Fail(t, "timed out")
		return
	}
}

type UnbecomeMessage struct{}

type EchoUnbecomeActor struct{}

func NewEchoUnbecomeActor() Actor {
	return &EchoBecomeActor{}
}

func (state *EchoUnbecomeActor) Receive(context Context) {
	switch context.Message().(type) {
	case BecomeMessage:
		context.BecomeStacked(state.Other)
	case EchoMessage:
		context.Sender().Tell(EchoReplyMessage{})
	}
}

func (*EchoUnbecomeActor) Other(context Context) {
	switch context.Message().(type) {
	case UnbecomeMessage:
		context.UnbecomeStacked()
	}
}

func TestActorCanUnbecome(t *testing.T) {
	actor := Spawn(FromProducer(NewEchoUnbecomeActor))
	actor.Tell(BecomeMessage{})
	actor.Tell(UnbecomeMessage{})
	result, _ := actor.AskFuture(EchoMessage{}, testTimeout)
	if _, err := result.Result(); err != nil {
		assert.Fail(t, "timed out")
		return
	}
}

type EchoOnStartActor struct{ replyTo *PID }

func (state *EchoOnStartActor) Receive(context Context) {
	switch context.Message().(type) {
	case *Started:
		state.replyTo.Tell(EchoReplyMessage{})
	}
}

func NewEchoOnStartActor(replyTo *PID) func() Actor {
	return func() Actor {
		return &EchoOnStartActor{replyTo: replyTo}
	}
}

func TestActorCanReplyOnStarting(t *testing.T) {
	future := NewFuture(testTimeout)
	actor := Spawn(FromProducer(NewEchoOnStartActor(future.PID())))
	defer actor.Stop()
	if _, err := future.Result(); err != nil {
		assert.Fail(t, "timed out")
		return
	}
}

type EchoOnStoppingActor struct{ replyTo *PID }

func (state *EchoOnStoppingActor) Receive(context Context) {
	switch context.Message().(type) {
	case *Stopping:
		state.replyTo.Tell(EchoReplyMessage{})
	}
}

func NewEchoOnStoppingActor(replyTo *PID) func() Actor {
	return func() Actor {
		return &EchoOnStoppingActor{replyTo: replyTo}
	}
}

func TestActorCanReplyOnStopping(t *testing.T) {
	future := NewFuture(testTimeout)
	actor := Spawn(FromProducer(NewEchoOnStoppingActor(future.PID())))
	actor.Stop()
	if _, err := future.Result(); err != nil {
		assert.Fail(t, "timed out")
		return
	}
}

type CreateChildMessage struct{}
type GetChildCountMessage struct{ ReplyTo *PID }
type GetChildCountReplyMessage struct{ ChildCount int }
type CreateChildActor struct{}

func (*CreateChildActor) Receive(context Context) {
	switch msg := context.Message().(type) {
	case CreateChildMessage:
		context.Spawn(FromProducer(NewBlackHoleActor))
	case GetChildCountMessage:
		reply := GetChildCountReplyMessage{ChildCount: len(context.Children())}
		msg.ReplyTo.Tell(reply)
	}
}

func NewCreateChildActor() Actor {
	return &CreateChildActor{}
}

func TestActorCanCreateChildren(t *testing.T) {
	future := NewFuture(testTimeout)
	actor := Spawn(FromProducer(NewCreateChildActor))
	defer actor.Stop()
	expected := 10
	for i := 0; i < expected; i++ {
		actor.Tell(CreateChildMessage{})
	}
	actor.Tell(GetChildCountMessage{ReplyTo: future.PID()})
	response, err := future.Result()
	if err != nil {
		assert.Fail(t, "timed out")
		return
	}
	assert.Equal(t, expected, response.(GetChildCountReplyMessage).ChildCount)
}

type CreateChildThenStopActor struct {
	replyTo *PID
}

type GetChildCountMessage2 struct {
	ReplyDirectly  *PID
	ReplyAfterStop *PID
}

func (state *CreateChildThenStopActor) Receive(context Context) {
	switch msg := context.Message().(type) {
	case CreateChildMessage:
		context.Spawn(FromProducer(NewBlackHoleActor))
	case GetChildCountMessage2:
		msg.ReplyDirectly.Tell(true)
		state.replyTo = msg.ReplyAfterStop
	case *Stopped:
		reply := GetChildCountReplyMessage{ChildCount: len(context.Children())}
		state.replyTo.Tell(reply)
	}
}

func NewCreateChildThenStopActor() Actor {
	return &CreateChildThenStopActor{}
}

func TestActorCanStopChildren(t *testing.T) {

	actor := Spawn(FromProducer(NewCreateChildThenStopActor))
	count := 10
	for i := 0; i < count; i++ {
		actor.Tell(CreateChildMessage{})
	}

	future := NewFuture(testTimeout)
	future2 := NewFuture(testTimeout)
	actor.Tell(GetChildCountMessage2{ReplyDirectly: future.PID(), ReplyAfterStop: future2.PID()})

	//wait for the actor to reply to the first responsePID
	_, err := future.Result()
	if err != nil {
		assert.Fail(t, "timed out")
		return
	}

	//then send a stop command
	actor.Stop()

	//wait for the actor to stop and get the result from the stopped handler
	response, err := future2.Result()
	if err != nil {
		assert.Fail(t, "timed out")
		return
	}
	//we should have 0 children when the actor is stopped
	assert.Equal(t, 0, response.(GetChildCountReplyMessage).ChildCount)
}

func TestDeadLetterAfterStop(t *testing.T) {
	actor := Spawn(FromProducer(NewBlackHoleActor))
	done := false
	sub := EventStream.Subscribe(func(msg interface{}) {
		if deadLetter, ok := msg.(*DeadLetter); ok {
			if deadLetter.PID == actor {
				done = true
			}
		}
	})
	defer EventStream.Unsubscribe(sub)

	actor.Stop()
	//stop is async, no clean way to wait for complete termination yet
	time.Sleep(100 * time.Millisecond)
	//send a message to the dead actor
	actor.Tell("hello")

	assert.True(t, done)
}
