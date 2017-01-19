package actor

type Spawner func(id string, props Props, parent *PID) *PID

var DefaultSpawner Spawner = spawn

//Spawn an actor with an auto generated id
func Spawn(props Props) *PID {
	return props.spawn(ProcessRegistry.NextId(), nil)
}

//SpawnNamed spawns a named actor
func SpawnNamed(props Props, name string) *PID {
	return props.spawn(name, nil)
}

func spawn(id string, props Props, parent *PID) *PID {
	cell := newLocalContext(props, parent)
	mailbox := props.ProduceMailbox()
	var ref Process = &localProcess{mailbox: mailbox}
	pid, absent := ProcessRegistry.Add(ref, id)

	if absent {
		pid.p = &ref
		cell.self = pid
		mailbox.RegisterHandlers(cell, props.Dispatcher())
		mailbox.PostSystemMessage(startedMessage)
	}

	return pid
}
