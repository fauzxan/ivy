package central

import (
        "container/list"
        "github.com/fauzxan/ivy/message"
        "log"
        "time"

        "github.com/fatih/color"
)

const (
        WRITE_FORWARD      = "write_forward"
        READ_FORWARD       = "read_forward"
        READ               = "read_request"
        WRITE              = "write_request"
        WRITE_CONFIRMATION = "write_confirmation"
        JOIN               = "join"
        EMPTY              = "empty"
        ACK                = "ack"
        FIRST              = "first" // first person to request for that page.
        I_HAVE_COPY        = "i_have_copy"
        INVALIDATE         = "invalidate"
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
        Records    []*Record
        Clientlist []string           // List of clients.
        writeQueue map[int]*list.List // List of pageids and their corresponding write requests; I will only respond to the head of the queue.
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
        case READ:
                for _, record := range central.Records {
                        if record.pageId == msg.PageId {
                                // Send a read forward to the owner. The owner will directly respond to msg.From
                                // Delay is introduced so that there is no race condition when there are two requests for a new page with the same id.
                                time.Sleep(750 * time.Millisecond)
                                go central.CallRPC(message.Message{Type: READ_FORWARD, From: msg.From, PageId: msg.PageId}, record.ownerIP)
                                reply.Type = ACK
                                return nil
                        }
                }
                // If you couldn't find the record, then it means this is a new page
                reply.Type = FIRST
                central.Records = append(central.Records, &Record{pageId: msg.PageId, copies: make([]string, 0), ownerIP: msg.From}) // Enter the owner IP as the owner
                system.Println("Added new record to Records", central.Records)
        case I_HAVE_COPY:
                system.Println("Client ", msg.From, " has a copy of ", msg.PageId)
                for _, record := range central.Records {
                        if record.pageId == msg.PageId {
                                record.copies = append(record.copies, msg.From)
                                system.Println("Records: ", central.Records)
                                break
                        }
                }
        case WRITE:
                system.Println("Received a write request from ", msg.From, " for page ", msg.PageId)
                _, ok := central.writeQueue[msg.PageId]
                if !ok {
                        central.writeQueue[msg.PageId] = list.New().Init()
                }

                central.writeQueue[msg.PageId].PushBack(msg.From)
                frontElement := central.writeQueue[msg.PageId].Front()
                if frontElement != nil && frontElement.Value == msg.From {
                        // msg.From is at the head of the queue. Call writeHandler()
                        system.Println("Executing WRITE request for ", msg.From)
                        first := central.writeHandler(msg.From, msg.PageId)
                        if first {
                                central.Records = append(central.Records, &Record{pageId: msg.PageId, copies: make([]string, 0), ownerIP: msg.From})
                                reply.Type = FIRST
                        } else {
                                reply.Type = ACK
                        }
                }
        case WRITE_CONFIRMATION:
                system.Println("Received a write confirmation from ", msg.From, " for page ", msg.PageId)
                frontElement := central.writeQueue[msg.PageId].Front() // Remove element from the head of the queue
                central.writeQueue[msg.PageId].Remove(frontElement)
                if central.writeQueue[msg.PageId].Len() != 0 { // If there are any more requests, then execute write handler on that request.
                        frontElement = central.writeQueue[msg.PageId].Front()
                        ip, _ := frontElement.Value.(string)
                        go central.writeHandler(ip, msg.PageId)
                }
        default:
                log.Fatal("This should never happen???")
        }
        return nil
}

func (central *Central) CreateNetwork() {
        system.Println("I am creating a new network")
        system.Println("Initialized records in memory...", central.Records)
        central.writeQueue = make(map[int]*list.List)
}

func (central *Central) invalidateSender(pageid int) {
        // extract the index of the page
        record_id := -1
        for id, record := range central.Records {
                if record.pageId == pageid {
                        record_id = id
                }
        }
        system.Println(pageid, " exists, sending invalidate now")
        if record_id != -1 {
                for _, copyIP := range central.Records[record_id].copies {
                        go central.CallRPC(message.Message{Type: INVALIDATE, PageId: pageid}, copyIP)
                }
        }

}

/*
1. Execute invalidate
2. Send write forward
3. Set owner as head of queue
return.
*/
func (central *Central) writeHandler(IP string, pageId int) bool {
        central.invalidateSender(pageId)
        for _, record := range central.Records {
                if record.pageId == pageId {
                        go central.CallRPC(message.Message{Type: WRITE_FORWARD, PageId: pageId, From: IP}, record.ownerIP)
                        record.ownerIP = IP
                        return false
                }
        }
        return true
}