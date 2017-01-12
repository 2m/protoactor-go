package actor

//Spawn an actor with an auto generated id
func Spawn(props Props) *PID {
	id := ProcessRegistry.getAutoId()
	pid := spawn(id, props, nil)
	return pid
}

//SpawnNamed spawns a named actor
func SpawnNamed(props Props, name string) *PID {
	pid := spawn(name, props, nil)
	return pid
}

func spawn(id string, props Props, parent *PID) *PID {
	if props.RouterConfig() != nil {
		return spawnRouter(id, props.RouterConfig(), props, parent)
	}

	cell := newActorCell(props, parent)
	mailbox := props.ProduceMailbox()
	ref := newLocalProcess(mailbox)
	pid, absent := ProcessRegistry.Add(ref, id)

	if absent {
		mailbox.RegisterHandlers(cell, props.Dispatcher())
		cell.self = pid
		cell.InvokeUserMessage(startedMessage)
	}

	return pid
}
