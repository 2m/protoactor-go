package actor

//Props or properties of an actor, it defines how the actor should be created
type Props struct {
	actorProducer       Producer
	mailboxProducer     MailboxProducer
	supervisionStrategy SupervisionStrategy
	routerConfig        RouterConfig
	receivePlugins      []Receive
	dispatcher          Dispatcher
}

func (props Props) Dispatcher() Dispatcher {
	if props.dispatcher == nil {
		return defaultDispatcher
	}
	return props.dispatcher
}
func (props Props) RouterConfig() RouterConfig {
	return props.routerConfig
}

func (props Props) ProduceActor() Actor {
	return props.actorProducer()
}

func (props Props) Supervisor() SupervisionStrategy {
	if props.supervisionStrategy == nil {
		return defaultSupervisionStrategy
	}
	return props.supervisionStrategy
}

func (props Props) ProduceMailbox() Mailbox {
	if props.mailboxProducer == nil {
		return NewUnboundedMailbox()()
	}
	return props.mailboxProducer()
}

func (props Props) WithReceivers(plugin ...Receive) Props {
	//pass by value, we only modify the copy
	props.receivePlugins = append(props.receivePlugins, plugin...)
	return props
}

func (props Props) WithMailbox(mailbox MailboxProducer) Props {
	//pass by value, we only modify the copy
	props.mailboxProducer = mailbox
	return props
}

func (props Props) WithSupervisor(supervisor SupervisionStrategy) Props {
	//pass by value, we only modify the copy
	props.supervisionStrategy = supervisor
	return props
}

func (props Props) WithPoolRouter(routerConfig PoolRouterConfig) Props {
	//pass by value, we only modify the copy
	props.routerConfig = routerConfig
	return props
}

func (props Props) WithDispatcher(dispatcher Dispatcher) Props {
	//pass by value, we only modify the copy
	props.dispatcher = dispatcher
	return props
}

func FromProducer(actorProducer Producer) Props {
	return Props{
		actorProducer:   actorProducer,
		mailboxProducer: nil,
		routerConfig:    nil,
	}
}

func FromFunc(receive Receive) Props {
	return FromInstance(receive)
}

func FromInstance(template Actor) Props {
	producer := func() Actor {
		return template
	}
	p := FromProducer(producer)
	return p
}

func FromGroupRouter(router GroupRouterConfig) Props {
	return Props{
		routerConfig:  router,
		actorProducer: nil,
	}
}
