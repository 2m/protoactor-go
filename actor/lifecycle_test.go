package actor

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestActorCanReplyOnStarting(t *testing.T) {
	future := NewFuture(testTimeout)
	a, err := rootContext.Spawn(PropsFromFunc(func(context Context) {
		switch context.Message().(type) {
		case *Started:
			context.Send(future.PID(), EchoResponse{})
		}
	}))
	assert.NoError(t, err)
	rootContext.StopFuture(a).Wait()
	assertFutureSuccess(future, t)
}

func TestActorCanReplyOnStopping(t *testing.T) {
	future := NewFuture(testTimeout)
	a, err := rootContext.Spawn(PropsFromFunc(func(context Context) {
		switch context.Message().(type) {
		case *Stopping:
			context.Send(future.PID(), EchoResponse{})
		}
	}))
	assert.NoError(t, err)
	rootContext.StopFuture(a).Wait()
	assertFutureSuccess(future, t)
}
