package central

import (
	"net/rpc"

	"github.com/fauzxan/ivy/message"
)

func (central *Central) CallRPC(msg message.Message, IP string) message.Message {
	system.Println("Central Manager ", central.IP, " is sending message of type ", msg.Type, " to ", IP)
	clnt, err := rpc.Dial("tcp", IP)
	reply := message.Message{}
	if err != nil {
		system.Println("Error dialing RPC", msg.Type)
		reply.Type = EMPTY
		return reply
	}
	err = clnt.Call("Client.HandleIncomingMessage", msg, &reply)
	if err != nil {
		system.Println("Error calling RPC", msg.Type)
		reply.Type = EMPTY
		return reply
	}
	system.Println("Client ", central.IP, " received reply from ", msg.From)
	return reply
}

func (central *Central) CallCentralRPC(msg message.Message, IP string) message.Message {
	system.Println("Central Manager ", central.IP, " is sending message of type ", msg.Type, " to ", IP)
	clnt, err := rpc.Dial("tcp", IP)
	reply := message.Message{}
	if err != nil {
		system.Println("Error dialing RPC", msg.Type)
		reply.Type = EMPTY
		return reply
	}
	err = clnt.Call("Central.HandleIncomingMessage", msg, &reply)
	if err != nil {
		system.Println("Error calling RPC", msg.Type)
		reply.Type = EMPTY
		return reply
	}
	system.Println("Client ", central.IP, " received reply from ", msg.From)
	return reply
}

func (central *Central) SelfCopySender(failureMessage SelfCopyMessage, IP string) message.Message {
	system.Println("CM Flushing Data to ", IP)
	clnt, err := rpc.Dial("tcp", IP)
	reply := message.Message{} // Placeholder, by the time it responds, the current server might be dead!
	if err != nil {
		system.Println("Error dialing RPC")
		return reply
	}
	err = clnt.Call("Central.HandleCopyMessages", failureMessage, &reply)
	if err != nil {
		system.Println("Error calling RPC\n", err)
		return reply
	}
	system.Println("Sucessfully flushed!")
	return reply
}

func (central *Central) PrintClientList() {
	system.Println(central.Clientlist)
}

func (central *Central) PrintRecords() {
	system.Println(central.Records)
	for _, record := range central.Records {
		system.Println(*record)
	}
}
