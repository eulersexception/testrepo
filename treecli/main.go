package main

import (
	"flag"
	"fmt"
	console "github.com/AsynkronIT/goconsole"
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/AsynkronIT/protoactor-go/remote"
	"github.com/ob-vss-ws19/blatt-3-suedachse/messages"
	"strconv"
	"sync"
	"time"
)


type Client struct {
	count int
	wg    *sync.WaitGroup
}

func (state *Client) Receive(context actor.Context) {
	debug(21, "called Receive()")
	switch msg := context.Message().(type) {
	case *messages.CreateResponse:
		fmt.Printf("Tree created! Id =  %v, token = %v\n", msg.GetId(), msg.GetToken())
	case *messages.DeleteTreeResponse:
		fmt.Printf("Response code %v - tree deletion alert. %v\n", msg.GetCode(), msg.GetMessage())
	case *messages.ForceTreeDeleteResponse:
		fmt.Printf("Response code %v - tree has been deleted. %v\n", msg.GetCode(), msg.GetMessage())
	case *messages.InsertResponse:
		fmt.Printf("Response code for insertion %v - %v\n", msg.GetCode(), msg.GetResult())
	case *messages.SearchResponse:
		fmt.Printf("Response code for search %v - value is %v\n", msg.GetCode(), msg.GetValue())
	case *messages.DeleteResponse:
		fmt.Printf("Response code for deletion %v - %v\n", msg.GetCode(), msg.GetResult())
	case *messages.TraverseResponse:
		fmt.Printf("Response code for traversion %v\n - %v\n", msg.GetCode(), msg.GetResult())

		for k, v := range msg.GetPairs() {
			fmt.Printf("{keys: %v, values: %v}\n", k, v)
		}

	default:
	}
	defer state.wg.Done()
}

const (
	newTree         = "newtree"
	deleteTree      = "deletetree"
	forceTreeDelete = "forTreeDelete"
	insert          = "insert"
	search          = "search"
	delete          = "delete"
	traverse        = "traverse"
)

func main() {

	debug(59, "Defining flags")
	flagBind := flag.String("bind", "localhost:18090", "Bind to address")
	flagRemote := flag.String("remote", "localhost:18091", "remote host:port")
	flagID := flag.Int("id", -1, "Tree id")
	flagToken := flag.String("token", "", "Tree token")
	debug(64, "Flags defined -- now parsing")
	flag.Parse()
	debug(66, "flags parsed")

	flagArgs := flag.Args()
	message := getMessage(int32(*flagID), *flagToken, flagArgs)

	if message == nil {
		printHelp()
		return
	}

	debug(76, "starting Remote")
	//remote.SetLogLevel(log.ErrorLevel)
	remote.Start(*flagBind)

	var wg sync.WaitGroup

	props := actor.PropsFromProducer(func() actor.Actor {
		wg.Add(10)
		return &Client{0, &wg}
	})
	rootContext := actor.EmptyRootContext
	pid := rootContext.Spawn(props)
	debug(87, fmt.Sprintf("created props, spawned them, got PID = %v", pid))

	pidResp, err := remote.SpawnNamed(*flagRemote, "remote", "treeservice", 5*time.Second)

	//remote.Register("treecli", props)
	debug(92, "registered Remote")

	if err != nil {
		fmt.Printf("Couldn't connect to %s\n", *flagRemote)
		return
	}

	remotePid := pidResp.Pid
	debug(100, fmt.Sprintf("got Remote PID = %v", remotePid))

	rootContext.RequestWithCustomSender(remotePid, message, pid)

	debug(104, fmt.Sprintf("Send message from PID %v to PID %v: \"%v\"", remotePid, pid, message))
	console.ReadLine()
	wg.Wait()
}

func printHelp() {
	help := "\n-----------------------------------------------------\n\n" +
		"  This is a demonstration of distributed software systems by \n" +
		"  an implementation of the \"Remote Actor Model\".\n" +
		"  By using listed commands you can create a tree to store key-value pairs. \n\n" +
		"  Keys are of type integer and values of type string. \n\n" +
		"  Create new tree:\n" +
		"    treecli newtree <max number of key-value-pairs>\n\n" +
		"  Commands on existing trees:\n" +
		"    treecli --id=<treeID> --token=<token> <command> <key> <value>\n\n" +
		"  Possible commands and parameters:\n" +
		"    " + insert + " <key> <value>\n" +
		"    " + search + " <key>\n" +
		"    " + delete + " <key>\n" +
		"    " + deleteTree + "\n" +
		"    " + forceTreeDelete + "\n" +
		"    " + traverse + "\n"
	fmt.Print(help)
}

func logError(err error) {
	fmt.Printf("An error ocured - %s", err.Error())
}

func debug(line int, info string) {
	fmt.Printf("TreeCli :: Line %v  --> %v\n", line, info)
}

func getMessage(id int32, token string, args []string) (message interface{}) {
	argsLength := len(args)
	message = &messages.ErrorResponse{Message: "too few arguments - check your command"}
	wrongCredentials := fmt.Sprintf("Id = %v or token = %v invalid", id, token)

	debug(142, fmt.Sprintf("getMessage(%v, %v) with default message \"to few arguments\"", id, token))

	if argsLength == 0 {
		return message
	}

	switch args[0] {
	case newTree:
		debug(150, "switched to case newTree")
		if argsLength == 2 {
			maxLeafSize, error := strconv.Atoi(args[1])
			if error == nil {
				debug(154, "preparing CreateRequest")
				message = &messages.CreateRequest{Code: int32(maxLeafSize)}
			}
		}
	case deleteTree:
		if argsLength == 1 {
			if id != -1 && token != "" {
				debug(161, "preparing DeleteRequest")
				message = &messages.DeleteTreeRequest{Id: id, Token: token}
			} else {
				debug(164, "preparing ErrorResponse")
				message = &messages.ErrorResponse{Message: wrongCredentials}
			}
		}
	case forceTreeDelete:
		if argsLength == 1 {
			if id != -1 && token != "" {
				debug(171, "preparing ForceTreeDeleteRequest")
				message = &messages.ForceTreeDeleteRequest{Id: id, Token: token}
			} else {
				debug(174, "preparing ErrorResponse")
				message = &messages.ErrorResponse{Message: wrongCredentials}
			}
		}
	case insert:
		if argsLength == 3 {
			key, error := strconv.Atoi(args[1])
			if error != nil {
				debug(182, "preparing ErrorResponse")
				response := fmt.Sprintf("invalid input for <key>: %s", args[1])
				message = &messages.ErrorResponse{Message: response}

				break
			}

			value := args[2]

			if id != -1 && token != "" {
				debug(192, "preparing InsertRequest")
				message = &messages.InsertRequest{Id: id, Token: token, Key: int32(key), Value: value, Success: true, }
			} else {
				debug(195, "preparing ErrorResponse")
				message = messages.ErrorResponse{Message: wrongCredentials}
			}
		}
	case search:
		if argsLength == 2 {
			key, error := strconv.Atoi(args[1])
			if error != nil {
				debug(203, "preparing ErrorResponse")
				response := fmt.Sprintf("invalid input for <key>: %s", args[1])
				message = &messages.ErrorResponse{Message: response}

				break
			}

			if id != -1 && token != "" {
				debug(211, "preparing SearchRequest")
				message = &messages.SearchRequest{Id: id, Token: token, Key: int32(key)}
			} else {
				debug(214, "preparing ErrorResponse")
				message = messages.ErrorResponse{Message: wrongCredentials}
			}
		}
	case delete:
		if argsLength == 2 {
			key, error := strconv.Atoi(args[1])

			if error != nil {
				debug(223, "preparing ErrorResponse")
				response := fmt.Sprintf("invalid input for <key>: %s", args[1])
				message = &messages.ErrorResponse{Message: response}

				break
			}

			if id != -1 && token != "" {
				debug(231, "preparing DeleteRequest")
				message = &messages.DeleteRequest{Id: id, Token: token, Key: int32(key)}
			} else {
				debug(234, "preparing ErrorResponse")
				message = messages.ErrorResponse{Message: wrongCredentials}
			}
		}
	case traverse:
		if argsLength == 1 {
			if id != -1 && token != "" {
				debug(241, "preparing TraverseRequest")
				message = &messages.TraverseRequest{Id: id, Token: token}
			} else {
				debug(244, "preparing ErrorResponse")
				message = messages.ErrorResponse{Message: wrongCredentials}
			}
		}
	default:
	}

	return message
}
