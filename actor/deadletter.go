package actor

import "log"

type deadLetterProcess struct{}

var (
	deadLetter           Process = &deadLetterProcess{}
	deadLetterSubscriber *Subscription
)

func init() {
	deadLetterSubscriber = EventStream.Subscribe(func(msg interface{}) {
		if deadLetter, ok := msg.(*DeadLetterEvent); ok {
			log.Printf("[DeadLetter] %v got %+v from %v", deadLetter.PID, deadLetter.Message, deadLetter.Sender)
		}
	})
}

type DeadLetterEvent struct {
	// PID specifies the process ID of the dead letter process
	PID *PID

	// Message specifies the message that could not be delivered
	Message interface{}

	// Sender specifies the process that sent the original Message
	Sender *PID
}

func (*deadLetterProcess) SendUserMessage(pid *PID, message interface{}, sender *PID) {
	EventStream.Publish(&DeadLetterEvent{
		PID:     pid,
		Message: message,
		Sender:  sender,
	})
}

func (*deadLetterProcess) SendSystemMessage(pid *PID, message SystemMessage) {
	EventStream.Publish(&DeadLetterEvent{
		PID:     pid,
		Message: message,
	})
}

func (ref *deadLetterProcess) Stop(pid *PID) {
	ref.SendSystemMessage(pid, stopMessage)
}
