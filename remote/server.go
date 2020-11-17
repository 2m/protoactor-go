package remote

import (
	"fmt"
	"io/ioutil"
	"net"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/AsynkronIT/protoactor-go/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
)

type Remote struct {
	actorSystem  *actor.ActorSystem
	s            *grpc.Server
	edpReader    *endpointReader
	config       *Config
	nameLookup   map[string]actor.Props
	activatorPid *actor.PID
}

func NewRemote(actorSystem *actor.ActorSystem, config Config) *Remote {
	return &Remote{
		actorSystem: actorSystem,
		config:      &config,
		nameLookup:  make(map[string]actor.Props),
	}
}

// Start the remote server
func (r *Remote) Start() {
	grpclog.SetLoggerV2(grpclog.NewLoggerV2(ioutil.Discard, ioutil.Discard, ioutil.Discard))
	lis, err := net.Listen("tcp", r.config.Address())
	if err != nil {
		panic(fmt.Errorf("failed to listen: %v", err))
	}

	var address string
	if r.config.AdvertisedHost != "" {
		address = r.config.AdvertisedHost
	} else {
		address = lis.Addr().String()
	}
	r.actorSystem.ProcessRegistry.RegisterAddressResolver(remoteHandler)
	r.actorSystem.ProcessRegistry.Address = address

	r.spawnActivatorActor()
	r.startEndpointManager()

	r.s = grpc.NewServer(r.config.ServerOptions...)
	r.edpReader = newEndpointReader(r)
	RegisterRemotingServer(r.s, r.edpReader)
	plog.Info("Starting Proto.Actor server", log.String("address", address))
	go r.s.Serve(lis)
}

func (r *Remote) Shutdown(graceful bool) {
	if graceful {
		r.edpReader.suspend(true)
		r.stopEndpointManager()
		r.stopActivatorActor()

		// For some reason GRPC doesn't want to stop
		// Setup timeout as workaround but need to figure out in the future.
		// TODO: grpc not stopping
		c := make(chan bool, 1)
		go func() {
			r.s.GracefulStop()
			c <- true
		}()

		select {
		case <-c:
			plog.Info("Stopped Proto.Actor server")
		case <-time.After(time.Second * 10):
			r.s.Stop()
			plog.Info("Stopped Proto.Actor server", log.String("err", "timeout"))
		}
	} else {
		r.s.Stop()
		plog.Info("Killed Proto.Actor server")
	}
}
