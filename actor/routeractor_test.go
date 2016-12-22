package actor

import (
	"testing"

	"github.com/stretchr/testify/mock"
)

func TestRouterActor_Receive_AddRoute(t *testing.T) {
	state := new(testRouterState)

	a := routerActor{state: state}

	p1 := NewLocalPID("p1")
	c := new(mockContext)
	c.On("Message").Return(&RouterAddRoutee{p1})
	c.On("Watch", p1).Once()

	state.On("GetRoutees").Return([]*PID{})
	state.On("SetRoutees", []*PID{p1}).Once()

	a.Receive(c)
	mock.AssertExpectationsForObjects(t, state, c)
}

func TestRouterActor_Receive_AddRoute_NoDuplicates(t *testing.T) {
	state := new(testRouterState)

	a := routerActor{state: state}

	p1 := NewLocalPID("p1")
	c := new(mockContext)
	c.On("Message").Return(&RouterAddRoutee{p1})

	state.On("GetRoutees").Return([]*PID{p1})

	a.Receive(c)
	mock.AssertExpectationsForObjects(t, state, c)
}

func TestRouterActor_Receive_RemoveRoute(t *testing.T) {
	state := new(testRouterState)

	a := routerActor{state: state}

	p1 := NewLocalPID("p1")
	p2 := NewLocalPID("p2")
	c := new(mockContext)
	c.On("Message").Return(&RouterRemoveRoutee{p1})
	c.On("Unwatch", p1).Once()

	state.On("GetRoutees").Return([]*PID{p1, p2})
	state.On("SetRoutees", []*PID{p2}).Once()

	a.Receive(c)
	mock.AssertExpectationsForObjects(t, state, c)
}

func TestRouterActor_Receive_BroadcastMessage(t *testing.T) {
	state := new(testRouterState)
	a := routerActor{state: state}

	p1 := NewLocalPID("p1")
	p2 := NewLocalPID("p2")

	child := new(mockActorRef)
	child.On("SendUserMessage", mock.Anything, mock.Anything, mock.Anything).Times(2)

	ProcessRegistry.add(child, "p1")
	ProcessRegistry.add(child, "p2")

	c := new(mockContext)
	c.On("Message").Return(&RouterBroadcastMessage{"hi"})
	c.On("Sender").Return((*PID)(nil))

	state.On("GetRoutees").Return([]*PID{p1, p2})

	a.Receive(c)
	mock.AssertExpectationsForObjects(t, state, c, child)
}
