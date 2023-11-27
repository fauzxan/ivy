package central

import (
	"ivy/message"
	"log"

	"github.com/fatih/color"
)

const (
	WRITE_FORWARD = "write_forward"
	READ = "read"
	WRITE = "write"
	JOIN = "join"
)

/*
To maintain a list of pages, and the corresponding copies that the clients own. 
*/
type Record struct{
	pageId int // id of the page being stored
	copies []int // id of the nodes that have the copy
	ownerid int // id of the nodes that owns the copy 
}

type Central struct{
	IP string
	IsPrimary bool // Used to indicate if I am the primary server or not
	Records []Record
	Clientlist []string // List of clients.
}

// Color coded logs
var system = color.New(color.FgCyan).Add(color.BgBlack)

/*
The default method called by all RPCs. This method receives different
types of requests, and calls the appropriate functions.
*/
func (central *Central) HandleIncomingMessage(msg *message.Message, reply *message.Message) error {
	
	switch msg.Type {
	case JOIN:
		central.Clientlist = append(central.Clientlist, msg.From)
		reply.Clientlist = central.Clientlist
		system.Println("Processed join request from", msg.From, central.Clientlist)
	case WRITE_FORWARD:

	case READ:

	case WRITE:

	default:
		log.Fatal("This should never happen???")
	}
	return nil
}


func (central *Central) CreateNetwork(){
	system.Println("I am creating a new network")
	central.Records = make([]Record, 0)
	system.Println("Initialized records in memory...", central.Records)
}