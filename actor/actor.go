package actor

type Actor interface {
	Receive(message *Context)
}

func spawn(props PropsValue) ActorRef {
	cell := NewActorCell(props)
	mailbox := NewDefaultMailbox(cell)
	ref := ChannelActorRef{
		mailbox: mailbox,
	}
	cell.Self = &ref
	ref.Tell(Starting{})
	return &ref
}
