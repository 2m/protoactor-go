package routing

import (
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/AsynkronIT/protoactor-go/process"
)

// SpawnPool spawns a pool router with an auto generated id
func SpawnPool(config PoolRouterConfig, props actor.Props) *process.PID {
	id := process.ProcessRegistry.NextId()
	pid := spawn(id, config, props, nil)
	return pid
}

// SpawnGroup spawns a pool router with an auto generated id
func SpawnGroup(config GroupRouterConfig) *process.PID {
	id := process.ProcessRegistry.NextId()
	pid := spawn(id, config, actor.Props{}, nil)
	return pid
}

// SpawnNamedPool spawns a named actor
func SpawnNamedPool(config RouterConfig, props actor.Props, name string) *process.PID {
	pid := spawn(name, config, props, nil)
	return pid
}

// SpawnNamedPool spawns a named actor
func SpawnNamedGroup(config RouterConfig, name string) *process.PID {
	pid := spawn(name, config, actor.Props{}, nil)
	return pid
}

func spawn(id string, config RouterConfig, props actor.Props, parent *process.PID) *process.PID {
	props = props.WithSpawn(nil)
	routerState := config.CreateRouterState()

	routerProps := actor.FromInstance(&routerActor{
		props:  props,
		config: config,
		state:  routerState,
	})

	routerID := process.ProcessRegistry.NextId()
	router := actor.DefaultSpawner(routerID, routerProps, parent)

	ref := &routerProcess{
		router: router,
		state:  routerState,
	}
	proxy, _ := process.ProcessRegistry.Add(ref, id)
	return proxy
}
