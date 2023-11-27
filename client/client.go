package client

import (
	"ivy/message"
	"log"

	"github.com/fatih/color"
)

// Color coded logs
var system = color.New(color.FgCyan).Add(color.BgBlack)

// Page access mode types
const (
	WRITE = "write"
	READ = "read"
)

// Message types
const (
	INVALIDATE = "invalidate"
	READ_REQUEST = "read_request"
	WRITE_REQUEST = "write_request"
	EMPTY = "empty"
	JOIN = "join"
)

type Page struct{
	ID int // ID of the page
	Content int // Content in the page
	AccessMode string // READ | WRITE
}


type Client struct{
	IP string
	Cache []Page // List of pages in its cache
	ServerIP string
	Clientlist []string
}


/*
The default method called by all RPCs. This method receives different
types of requests, and calls the appropriate functions.
*/
func (client *Client) HandleIncomingMessage(msg *message.Message, reply *message.Message) error {
	
	switch msg.Type {
	case JOIN:
		client.Clientlist = append(client.Clientlist, msg.From)
		system.Println("Processed join request from", msg.From, client.Clientlist)
	case READ:

	case WRITE:

	default:
		log.Fatal("This should never happen???")
	}
	return nil
}


func (client *Client) JoinNetwork(helper string){
	client.ServerIP = helper
	// Contact the server and retrieve the clientlist first.
	reply := client.CallCentralRPC(message.Message{Type: JOIN, From: client.IP}, client.ServerIP)
	client.Clientlist = reply.Clientlist

	for _, ip := range client.Clientlist{
		if ip == client.IP {continue}
		message := message.Message{Type: JOIN, From: client.IP}
		go func(){
			client.CallCentralRPC(message, ip)
			system.Println("Successfully notified", ip)
		}()
	}
}

