package actor

import (
	"github.com/AsynkronIT/protoactor-go/eventstream"
)

//goland:noinspection GoNameStartsWithPackageName
type ActorSystem struct {
	ProcessRegistry *ProcessRegistryValue
	Root            *RootContext
	EventStream     *eventstream.EventStream
	Guardians       *guardiansValue
}

func New() *ActorSystem {
	system := &ActorSystem{}

	system.ProcessRegistry = NewProcessRegistry(system)
	system.Root = NewRootContext(system, EmptyMessageHeader)
	system.Guardians = NewGuardians(system)
	deadletterSubscribe(system)

	return system
}
