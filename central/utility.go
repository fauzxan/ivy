package central

import (
        "github.com/fauzxan/ivy/message"
        "net/rpc"
)

func (central *Central) CallRPC(msg message.Message, IP string) message.Message {
        system.Println("Client ", central.IP, " is sending message of type ", msg.Type, " to ", IP)
        clnt, err := rpc.Dial("tcp", IP)
        reply := message.Message{}
        if err != nil {
                system.Println("Error dialing RPC", msg.Type)
                reply.Type = EMPTY
                return reply
        }
        err = clnt.Call("Client.HandleIncomingMessage", msg, &reply)
        if err != nil {
                system.Println("Error callling RPC", msg.Type)
                reply.Type = EMPTY
                return reply
        }
        system.Println("Client ", central.IP, " received reply from ", msg.From)
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