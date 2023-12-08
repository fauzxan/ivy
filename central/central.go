package central

import (
	"container/list"
	"log"
	"time"

	"github.com/fauzxan/ivy/message"

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
	SERVER_JOIN        = "server_join"
	FAILURE            = "failure"
	NEW_SERVER         = "new_server"
)

/*
To maintain a list of pages, and the corresponding Copies that the clients own.
*/
type Record struct {
	PageId  int      // id of the page being stored
	Copies  []string // id of the nodes that have the copy
	OwnerIP string   // id of the nodes that owns the copy
}

type Central struct {
	IP          string
	Records     []*Record
	Clientlist  []string           // List of clients.
	writeQueue  map[int]*list.List // List of pageids and their corresponding write requests; I will only respond to the head of the queue.
	otherServer string
	DoIPing   bool
	DoesTheOtherPing bool
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
	case SERVER_JOIN:
		central.otherServer = msg.From
		system.Println("Another server has just booted up at IP:", central.otherServer)
		reply.Type = ACK
	case READ:
		for _, record := range central.Records {
			if record.PageId == msg.PageId {
				// Send a read forward to the owner. The owner will directly respond to msg.From
				// Delay is introduced so that there is no race condition when there are two requests for a new page with the same id.
				time.Sleep(750 * time.Millisecond)
				go central.CallRPC(message.Message{Type: READ_FORWARD, From: msg.From, PageId: msg.PageId}, record.OwnerIP)
				reply.Type = ACK
				return nil
			}
		}
		// If you couldn't find the record, then it means this is a new page
		reply.Type = FIRST
		central.Records = append(central.Records, &Record{PageId: msg.PageId, Copies: make([]string, 0), OwnerIP: msg.From}) // Enter the owner IP as the owner
		system.Println("Added new record to Records", central.Records)
	case I_HAVE_COPY:
		system.Println("Client ", msg.From, " has a copy of ", msg.PageId)
		for _, record := range central.Records {
			if record.PageId == msg.PageId {
				record.Copies = append(record.Copies, msg.From)
				system.Println("Records: ", central.Records)
				break
			}
		}
	case WRITE:
		system.Println("Received a write request from ", msg.From, " for page ", msg.PageId)
		if central.writeQueue == nil {
			central.writeQueue = make(map[int]*list.List)
		}

		// check to see if msg.PageId exists in the map. Else assign it to an empty queue
		_, ok := central.writeQueue[msg.PageId]
		if !ok {
			central.writeQueue[msg.PageId] = list.New().Init()
		}

		// enter the new request for that page into writeQueue
		central.writeQueue[msg.PageId].PushBack(msg.From)
		frontElement := central.writeQueue[msg.PageId].Front()
		if frontElement != nil && frontElement.Value == msg.From {
			// msg.From is at the head of the queue. Call writeHandler()
			system.Println("Executing WRITE request for ", msg.From)
			first := central.writeHandler(msg.From, msg.PageId)
			if first {
				central.Records = append(central.Records, &Record{PageId: msg.PageId, Copies: make([]string, 0), OwnerIP: msg.From})
				reply.Type = FIRST
			} else {
				reply.Type = ACK
			}
		}
	case WRITE_CONFIRMATION:
		system.Println("Received a write confirmation from ", msg.From, " for page ", msg.PageId)
		frontElement := central.writeQueue[msg.PageId].Front() // Remove element from the head of the queue
		if central.writeQueue[msg.PageId].Len() != 0{
			central.writeQueue[msg.PageId].Remove(frontElement)
		}
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

func (central *Central) HandleCopyMessages(msg *SelfCopyMessage, reply *message.Message) error {
	system.Println("I am on main duty now! Setting new details!")
	// Update details locally
	central.otherServer = msg.Central.IP
	central.writeQueue = msg.Central.writeQueue
	central.Clientlist = msg.Central.Clientlist
	central.Records = msg.Central.Records
	central.DoIPing = msg.Central.DoesTheOtherPing
	central.DoesTheOtherPing = msg.Central.DoIPing
	reply.Type = ACK
	go func() {
		for _, client := range central.Clientlist {
			central.CallRPC(message.Message{Type: NEW_SERVER, From: central.IP}, client)
			system.Println("Notified ", client, " about new server")
		}
		if central.DoIPing{
			go central.PingMain()
		}
	}()
	return nil
}

func (central *Central) CreateNetwork(backup string) {
	system.Println("I am creating a new network")
	system.Println("Initialized records in memory...", central.Records)
	central.writeQueue = make(map[int]*list.List)
	if backup != "" {
		central.otherServer = backup
		go central.CallCentralRPC(message.Message{Type: SERVER_JOIN, From: central.IP}, central.otherServer)
	}
}

func (central *Central) invalidateSender(pageid int) {
	// extract the index of the page
	record_id := -1
	for id, record := range central.Records {
		if record.PageId == pageid {
			record_id = id
		}
	}

	if record_id != -1 { // Checking if the central manager has the record or nah
		for _, copyIP := range central.Records[record_id].Copies {
			if copyIP != central.Records[record_id].OwnerIP { //  Don't call INVALIDATE on the current owner. The current owner will delete the record anyway after receving write forward.
				go central.CallRPC(message.Message{Type: INVALIDATE, PageId: pageid}, copyIP)
			}
			// removing the element to which you just sent invalidate to. remove the owner as well.
			if len(central.Records[record_id].Copies) > 0{
				central.Records[record_id].Copies = append(central.Records[record_id].Copies[:0], central.Records[record_id].Copies[1:]...)
			}
		}
	}

}

/*
1. Execute invalidate
2. Send write forward
3. Set owner as head of queue
return.
*/
func (central *Central) writeHandler(IP string, PageId int) bool {
	central.invalidateSender(PageId)
	for _, record := range central.Records {
		if record.PageId == PageId {
			go central.CallRPC(message.Message{Type: WRITE_FORWARD, PageId: PageId, From: IP}, record.OwnerIP)
			record.OwnerIP = IP
			return false
		}
	}
	return true
}

/*
Intended to send message of type: FAILURE to central.otherServer
*/
func (central *Central) FlushData() {
	message := SelfCopyMessage{Central: Central{
		IP:          central.IP,
		Records:     central.Records,
		Clientlist:  central.Clientlist,
		writeQueue:  central.writeQueue,
		otherServer: central.otherServer,
		DoIPing: central.DoIPing,
		DoesTheOtherPing: central.DoesTheOtherPing,
	},
	}
	central.SelfCopySender(message, central.otherServer)
}

/*
Once ping receives an ACK, we know the main server is back up again. Meaning, we can just send a copy to it, and it will take care of the rest. 
*/
func (central *Central) PingMain() {
	for {
		time.Sleep(5 * time.Second)
		system.Println("Trying to ping main server to see if it is back up again... Do not be alarmed if RPC failed.")
		message := SelfCopyMessage{Central: Central{
			IP:          central.IP,
			Records:     central.Records,
			Clientlist:  central.Clientlist,
			writeQueue:  central.writeQueue,
			otherServer: central.otherServer,
			DoIPing: central.DoIPing,
			DoesTheOtherPing: central.DoesTheOtherPing,
		},
		}
		reply := central.SelfCopySender(message, central.otherServer)
		if reply.Type == ACK {
			system.Println("MAIN CM IS BACK AGAIN")
			return
		}
	}
}
