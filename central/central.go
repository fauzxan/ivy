package central

import (
        "ivy/message"
        "log"
        "time"

        "github.com/fatih/color"
)

const (
        WRITE_FORWARD = "write_forward"
        READ_FORWARD  = "read_forward"
        READ          = "read_request"
        WRITE         = "write_request"
        JOIN          = "join"
        EMPTY         = "empty"
        ACK           = "ack"
        FIRST         = "first" // first person to request for that page.
        I_HAVE_COPY   = "i_have_copy"
)

/*
To maintain a list of pages, and the corresponding copies that the clients own.
*/
type Record struct {
        pageId  int      // id of the page being stored
        copies  []string // id of the nodes that have the copy
        ownerIP string   // id of the nodes that owns the copy
}

type Central struct {
        IP         string
        IsPrimary  bool // Used to indicate if I am the primary server or not
        Records    []Record
        Clientlist []string // List of clients.
}

// Color coded logs
var system = color.New(color.FgCyan).Add(color.BgBlack)

/*
The default method called by all RPCs. This method receives different
types of requests, and calls the appropriate functions.
*/
func (central *Central) HandleIncomingMessage(msg *message.Message, reply *message.Message) error {
        system.Println("Message ", msg.Type)

        switch msg.Type {
        case JOIN:
                central.Clientlist = append(central.Clientlist, msg.From)
                reply.Clientlist = central.Clientlist
                system.Println("Processed join request from", msg.From, central.Clientlist)
        case READ_FORWARD:

        case READ:
                for _, record := range central.Records {
                        if record.pageId == msg.PageId {
                                // Send a read forward to the owner. The owner will directly respond to msg.From
                                // Delay is introduced so that there is no race condition when there are two requests for a new page with the same id.
                                time.Sleep(750 * time.Millisecond)
                                central.CallRPC(message.Message{Type: READ_FORWARD, From: msg.From, PageId: msg.PageId}, record.ownerIP)
                                reply.Type = ACK
                                return nil
                        }
                }
                // If you couldn't find the record, then it means this is a new page
                reply.Type = FIRST
                central.Records = append(central.Records, Record{pageId: msg.PageId, copies: make([]string, 0), ownerIP: msg.From}) // Enter the owner IP as the owner
                system.Println("Added new record to Records", central.Records)
        case I_HAVE_COPY:
                system.Println("Client ", msg.From, " has a copy of ", msg.PageId)
                go func() {
                        for _, record := range central.Records {
                                if record.pageId == msg.PageId {
                                        record.copies = append(record.copies, msg.From)
                                        system.Println("Records: ", central.Records)
                                        return
                                }
                        }
                }()

        default:
                log.Fatal("This should never happen???")
        }
        return nil
}

func (central *Central) CreateNetwork() {
        system.Println("I am creating a new network")
        system.Println("Initialized records in memory...", central.Records)
}