package client

import (
        "github.com/fauzxan/ivy/message"
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
        INVALIDATE         = "invalidate"
        READ               = "read_request"
        READ_FORWARD       = "read_forward"
        WRITE_FORWARD      = "write_forward"
        WRITE              = "write_request"
        WRITE_CONFIRMATION = "write_confirmation"
        EMPTY              = "empty"
        JOIN               = "join"
        COPY               = "copy"
        WRITE_COPY         = "write_copy"
        FIRST              = "first" // When you're the first node requesting read access to the page, so you can directly set it's
        ACK                = "ack"
        I_HAVE_COPY        = "i_have_copy"
        NEW_SERVER		   = "new_server"
)

type Page struct {
        Content    int    // Content in the page. For the purpose of this implementation, we will make it a simple integer.
        AccessMode string // READ | WRITE
}

type Client struct {
        IP         string
        Cache      map[int]Page // List of pages in its cache. pageId -> Page{Content, Access Mode}
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
                reply.Type = ACK
        case READ_FORWARD:
                /*
                   I would receive read forward if someone somewhere has requested. I will receive read forward from the central manager.
                   Upon receipt, I just need to send a copy of the page to the requestor, which has been tagged in msg.From
                */
                system.Println("I received a message of type READ FORWARD for client", msg.From)
                go client.CallRPC(message.Message{Type: COPY, Content: client.Cache[msg.PageId].Content, From: client.IP, PageId: msg.PageId}, msg.From)
        case WRITE_FORWARD:
                system.Println("Received a write forward instruction. Will forward copy of ", msg.PageId, " to ", msg.From)
                client.CallRPC(message.Message{Type: WRITE_COPY, From: client.IP, Content: client.Cache[msg.PageId].Content, PageId: msg.PageId}, msg.From)
                delete(client.Cache, msg.PageId)
        case COPY:
                system.Println("Copy of ", msg.PageId, " received from ", msg.From, ": ", msg.Content)
                client.Cache[msg.PageId] = Page{Content: msg.Content, AccessMode: READ_MODE}
                go client.CallCentralRPC(message.Message{Type: I_HAVE_COPY, PageId: msg.PageId, From: client.IP}, client.ServerIP)
        case WRITE_COPY:
                system.Println("Finally received a write copy of ", msg.PageId, " from ", msg.From)
                content := rand.Intn(100)
                client.Cache[msg.PageId] = Page{Content: content, AccessMode: READ_MODE}
                go client.CallCentralRPC(message.Message{Type: WRITE_CONFIRMATION, PageId: msg.PageId, From: msg.From}, client.ServerIP)
        case INVALIDATE:
                system.Println(msg.Type, " received for page ", msg.PageId)
                go client.invalidate(msg.PageId)
                reply.Type = ACK
        case NEW_SERVER:
                client.ServerIP = msg.From
                system.Println("There is a new server: ", client.ServerIP)
        default:
                log.Fatal("This should never happen???")
        }
        return nil
}

/*
Function called to join the network, by contacting the central manager. This is so the central manager knows about you in its metadata.
The client also obtains the clientlist from the central manager, and notifies all the other clients of its existence.
*/
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
                        client.CallRPC(message, copy)
                        system.Println("Successfully notified", copy)
                }()
        }
        client.Cache = make(map[int]Page)
}

/*
A requesting node will first check it's own cache and then check with the central manager, in case it doesn't have it. 
It may also update the record and become it's owner, if there is no other node in the network that owns this page.
*/
func (client *Client) ReadRequest() {
        for {
                timeNow := time.Now()
                if timeNow.Second()%15 == 0 {
                        break
                }
        } // Even if I press one at slightly different times, they will all send simultaneously at the same time.

        pageid := rand.Intn(3)
        system.Println("Making request to read ", pageid)
        value, ok := client.Cache[pageid]
        if !ok { // if page is not in cache
                reply := client.CallCentralRPC(message.Message{Type: READ, PageId: pageid, From: client.IP}, client.ServerIP)
                if reply.Type == FIRST {
                        system.Println("I am actually the first one requesting to read this page.")
                        content := rand.Intn(100)
                        client.Cache[pageid] = Page{Content: content, AccessMode: READ_MODE}
                        system.Println("Set new entry for page ", pageid, " with content ", content)
                        go client.CallCentralRPC(message.Message{Type: I_HAVE_COPY, PageId: pageid, From: client.IP}, client.ServerIP)
                } else if reply.Type == ACK {
                        return
                }
        } else {
                system.Println("COPY IN LOCAL CACHE:", value)
        }
}

func (client *Client) WriteRequest() {
        for {
                timeNow := time.Now()
                if timeNow.Second()%15 == 0 {
                        break
                }
        } // Even if I press one at slightly different times, they will all send concurrently at the same time.

        pageid := rand.Intn(3)
        system.Println("Making request to write ", pageid)
        reply := client.CallCentralRPC(message.Message{Type: WRITE, From: client.IP, PageId: pageid}, client.ServerIP)
        if reply.Type == FIRST {
                system.Println("I am actually the first one requesting to write to this page.")
                content := rand.Intn(100)
                client.Cache[pageid] = Page{Content: content, AccessMode: READ_MODE}
                system.Println("Set new entry for page ", pageid, " with content ", content)
                go client.CallCentralRPC(message.Message{Type: WRITE_CONFIRMATION, PageId: pageid, From: client.IP}, client.ServerIP)
        }

}

// Loop through Cache, and invalidate the page whose id == pageId parameter.
func (client *Client) invalidate(pageId int) {
        delete(client.Cache, pageId)
}