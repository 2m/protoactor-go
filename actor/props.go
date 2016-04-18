package actor

type ActorProducer func() Actor
type MailboxProducer func(func(interface{}),func(interface{})) Mailbox

type PropsValue struct {
	actorProducer   ActorProducer
	mailboxProducer MailboxProducer
	routerConfig    RouterConfig
}

func Props(actorProducer ActorProducer) PropsValue {
	return PropsValue{
		actorProducer: actorProducer,
		mailboxProducer: NewQueueMailbox,
	}
}

func (props PropsValue) WithRouter(routerConfig RouterConfig) PropsValue {
	//pass by value, we only modify the copy
	props.routerConfig = routerConfig
	return props
}

func (props PropsValue) WithMailbox(mailboxProducer MailboxProducer) PropsValue {
	//pass by value, we only modify the copy
	props.mailboxProducer = mailboxProducer
	return props
}
