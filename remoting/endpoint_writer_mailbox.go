package remoting

import (
	"sync/atomic"

	"github.com/AsynkronIT/gam/actor"
	"github.com/AsynkronIT/goring"

	"runtime"
)

const (
	mailboxIdle    int32 = iota
	mailboxRunning int32 = iota
)
const (
	mailboxHasNoMessages   int32 = iota
	mailboxHasMoreMessages int32 = iota
)

type endpointWriterMailbox struct {
	userMailbox     *goring.Queue
	systemMailbox   *goring.Queue
	schedulerStatus int32
	hasMoreMessages int32
	userInvoke      func(interface{})
	systemInvoke    func(actor.SystemMessage)
	batchSize       int
}

func (mailbox *endpointWriterMailbox) PostUserMessage(message interface{}) {
	//batching mailbox only use the message part
	mailbox.userMailbox.Push(message)
	mailbox.schedule()
}

func (mailbox *endpointWriterMailbox) PostSystemMessage(message actor.SystemMessage) {
	mailbox.systemMailbox.Push(message)
	mailbox.schedule()
}

func (mailbox *endpointWriterMailbox) schedule() {
	atomic.StoreInt32(&mailbox.hasMoreMessages, mailboxHasMoreMessages) //we have more messages to process
	if atomic.CompareAndSwapInt32(&mailbox.schedulerStatus, mailboxIdle, mailboxRunning) {
		go mailbox.processMessages()
	}
}

func (mailbox *endpointWriterMailbox) Suspend() {

}

func (mailbox *endpointWriterMailbox) Resume() {

}

func (mailbox *endpointWriterMailbox) processMessages() {
	//we are about to start processing messages, we can safely reset the message flag of the mailbox
	atomic.StoreInt32(&mailbox.hasMoreMessages, mailboxHasNoMessages)
	batchSize := mailbox.batchSize
	done := false

	for !done {
		if sysMsg, ok := mailbox.systemMailbox.Pop(); ok {

			first := sysMsg.(actor.SystemMessage)
			mailbox.systemInvoke(first)
		} else if userMsg, ok := mailbox.userMailbox.PopMany(int64(batchSize)); ok {
			//pack the batch in a user message
			mailbox.userInvoke(userMsg)
		} else {
			done = true
			break
		}

		if !done {
			runtime.Gosched()
		}
	}

	//set mailbox to idle
	atomic.StoreInt32(&mailbox.schedulerStatus, mailboxIdle)
	//check if there are still messages to process (sent after the message loop ended)
	if atomic.SwapInt32(&mailbox.hasMoreMessages, mailboxHasNoMessages) == mailboxHasMoreMessages {
		mailbox.schedule()
	}

}

func newEndpointWriterMailbox(batchSize, initialSize int) actor.MailboxProducer {

	return func() actor.Mailbox {
		userMailbox := goring.New(int64(initialSize))
		systemMailbox := goring.New(10)
		mailbox := endpointWriterMailbox{
			userMailbox:     userMailbox,
			systemMailbox:   systemMailbox,
			hasMoreMessages: mailboxHasNoMessages,
			schedulerStatus: mailboxIdle,
			batchSize:       batchSize,
		}
		return &mailbox
	}
}

func (mailbox *endpointWriterMailbox) RegisterHandlers(userInvoke func(interface{}), systemInvoke func(actor.SystemMessage)) {
	mailbox.userInvoke = userInvoke
	mailbox.systemInvoke = systemInvoke
}
