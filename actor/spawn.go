package actor

import "github.com/AsynkronIT/protoactor-go/process"

type Spawner func(id string, props Props, parent *process.ID) *process.ID

var DefaultSpawner Spawner = spawn

//Spawn an actor with an auto generated id
func Spawn(props Props) *process.ID {
	return props.spawn(process.Registry.NextId(), nil)
}

//SpawnNamed spawns a named actor
func SpawnNamed(props Props, name string) *process.ID {
	return props.spawn(name, nil)
}

func spawn(id string, props Props, parent *process.ID) *process.ID {
	cell := newActorCell(props, parent)
	mailbox := props.ProduceMailbox()
	ref := newLocalProcess(mailbox)
	pid, absent := process.Registry.Add(ref, id)

	if absent {
		mailbox.RegisterHandlers(cell, props.Dispatcher())
		cell.self = pid
		cell.InvokeUserMessage(startedMessage)
	}

	return pid
}
