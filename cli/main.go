package main

import (
	"fmt"
	"io"
	"log"
	"math"
	"strconv"
	"strings"

	"github.com/chzyer/readline"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/AsynkronIT/protoactor-go/remote"
	proto "github.com/gogo/protobuf/proto"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// Function constructor - constructs new function for listing given directory
var completer = readline.NewPrefixCompleter(
	readline.PcItem("tell"),
	readline.PcItem("exit"),
)

func filterInput(r rune) (rune, bool) {
	switch r {
	// block CtrlZ feature
	case readline.CharCtrlZ:
		return r, false
	}
	return r, true
}

func main() {
	logo := `
     ___         _         ___ _    ___
    | _ \_ _ ___| |_ ___  / __| |  |_ _|
    |  _/ '_/ _ \  _/ _ \| (__| |__ | |
    |_| |_| \___/\__\___(_)___|____|___|
`
	fmt.Println(logo)

	remote.DefaultSerializerID = 1
	remote.Start("127.0.0.1:0")
	actor.SpawnNamed(actor.FromFunc(func(ctx actor.Context) {
		log.Printf("ECHO: %+v", ctx.Message())
	}), "echo")

	vars := make(map[string]string)
	vars["%address%"] = actor.ProcessRegistry.Address
	vars["%echo%"] = fmt.Sprintf(`{"Address":"%v", "Id":"echo"}`, actor.ProcessRegistry.Address)

	l, err := readline.NewEx(&readline.Config{
		Prompt:          "\033[31m»\033[0m ",
		HistoryFile:     "/tmp/readline.tmp",
		AutoComplete:    completer,
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",

		HistorySearchFold:   true,
		FuncFilterInputRune: filterInput,
	})

	if err != nil {
		panic(err)
	}
	defer l.Close()

	log.SetOutput(l.Stderr())
	for {
		line, err := l.Readline()
		if err == readline.ErrInterrupt {
			if len(line) == 0 {
				break
			} else {
				continue
			}
		} else if err == io.EOF {
			break
		}

		line = strings.TrimSpace(line)
		for k, v := range vars {
			line = strings.Replace(line, k, v, 1000)
		}
		log.Println(line)
		switch {

		case strings.HasPrefix(line, "tell "):
			line = tell(line)

		case line == "exit":
			goto exit
		case line == "":
		default:
			log.Println("Unknown command :", strconv.Quote(line))
		}
	}
exit:
}
func tell(line string) string {
	parts := strings.SplitN(line, " ", 4)
	i := parts[1]
	x := strings.SplitN(i, "/", 2)
	address := x[0]
	id := x[1]
	m := &remote.JsonMessage{
		Json:     parts[3],
		TypeName: parts[2],
	}
	pid := actor.NewPID(address, id)
	remote.SendMessage(pid, m, nil, 1)
	return line
}
