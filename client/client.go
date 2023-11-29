package client

import (
        "ivy/message"
        "log"
        "math/rand"
        "time"

        "github.com/fatih/color"
)

// Color coded logs
var system = color.New(color.FgCyan).Add(color.BgBlack)

// Page access mode types
const (
        WRITE_MODE = "write"
        READ_MODE  = "read"
)

// Message types
const (
        INVALIDATE   = "invalidate"
        READ         = "read_request"
        READ_FORWARD = "read_forward"
        WRITE        = "write_request"
        EMPTY        = "empty"
        JOIN         = "join"
        COPY         = "copy"
        FIRST        = "first"
        ACK          = "ack"
		I_HAVE_COPY = "i_have_copy"
)

type Page struct {
        Content    int    // Content in the page. For the purpose of this implementation, we will make it a simple integer.
        AccessMode string // READ | WRITE
}

type Client struct {
        IP         string
        Cache      map[int]Page // List of pages in its cache
        ServerIP   string
        Clientlist []string
        Timestamp  int // scalar clock to indicate when the request was made. If there are concurrent requests, then we need to break ties
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
        case READ_FORWARD:
                /*
                        I would receive read forward if someone somewhere has requested. I will receive read forward from the central manager.
                        Upon receipt, I just need to send a copy of the page to the requestor, which has been tagged in msg.From
                */
                system.Println("I received a message of type READ FORWARD for client", msg.From)
                go client.CallRPC(message.Message{Type: COPY, Content: client.Cache[msg.PageId].Content, From: client.IP}, msg.From)

        case COPY:
                system.Println("Copy of ", msg.PageId, " received from ", msg.From, ": ", msg.Content)
				client.Cache[msg.PageId] = Page{Content: msg.Content, AccessMode: READ_MODE}
				go client.CallCentralRPC(message.Message{Type: I_HAVE_COPY, PageId: msg.PageId, From: msg.From}, client.ServerIP)
        default:
                log.Fatal("This should never happen???")
        }
        return nil
}

func (client *Client) JoinNetwork(helper string) {
        client.ServerIP = helper
        // Contact the server and retrieve the clientlist first.
        reply := client.CallCentralRPC(message.Message{Type: JOIN, From: client.IP}, client.ServerIP)
        client.Clientlist = reply.Clientlist

        for _, ip := range client.Clientlist {
                if ip == client.IP {
                        continue
                }
                message := message.Message{Type: JOIN, From: client.IP}
                copy := ip
                go func() {
                        client.CallCentralRPC(message, copy)
                        system.Println("Successfully notified", copy)
                }()
        }
        client.Cache = make(map[int]Page)
}

/*
A requesting node will first
*/
func (client *Client) ReadRequest() {
        for {
                timeNow := time.Now()
                if timeNow.Second()%10 == 0 {
                        break
                }
        } // Even if I press one at slightly different times, they will all send concurrently at the same time.

        pageid := rand.Intn(3)
		system.Println("Making request to read ", pageid)
        value, ok := client.Cache[pageid]
        if !ok { // if page is not in cache
                reply := client.CallCentralRPC(message.Message{Type: READ, PageId: pageid, From: client.IP}, client.ServerIP)
                if reply.Type == FIRST {
                        system.Println("I am actually the first one requesting for this page.")
                        content := rand.Intn(100)
                        client.Cache[pageid] = Page{Content: content, AccessMode: READ_MODE}
                        system.Println("Set new entry for page ", pageid, " with content ", content)
                } else if reply.Type == ACK {
                        return
                }
        } else {
                system.Println("COPY IN LOCAL CACHE:", value)
        }
}

func (client *Client) WriteRequest(pageId int) {
        client.CallCentralRPC(message.Message{Type: WRITE}, client.ServerIP)

}