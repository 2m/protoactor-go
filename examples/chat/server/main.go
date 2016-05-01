package main

import (
	"bufio"
	"log"
	"os"
	"runtime"

	"github.com/emirpasic/gods/sets/hashset"
	"github.com/rogeralsing/gam/actor"
	"github.com/rogeralsing/gam/examples/chat/messages"
	"github.com/rogeralsing/gam/remoting"
)

func notifyAll(clients *hashset.Set, message interface{}) {
	for _, tmp := range clients.Values() {
		client := tmp.(*actor.PID)
		client.Tell(message)
	}
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	remoting.StartServer("127.0.0.1:8080")
	clients := hashset.New()
	server := actor.SpawnReceiveFunc(func(context actor.Context) {
		switch msg := context.Message().(type) {
		case *messages.Connect:
			log.Printf("Client %v connected", msg.Sender)
			clients.Add(msg.Sender)
			msg.Sender.Tell(&messages.Connected{Message: "Welcome!"})
		case *messages.SayRequest:
			notifyAll(clients, &messages.SayResponse{
				UserName: msg.UserName,
				Message:  msg.Message,
			})
		case *messages.NickRequest:
			notifyAll(clients, &messages.NickResponse{
				OldUserName: msg.OldUserName,
				NewUserName: msg.NewUserName,
			})
		}
	})
	actor.ProcessRegistry.Register("chatserver", server)
	bufio.NewReader(os.Stdin).ReadString('\n')
}
