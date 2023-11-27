package client

import (
	"ivy/message"
	"net/rpc"
)


func (client *Client) CallCentralRPC(msg message.Message, IP string) message.Message {
	system.Println("Client ", client.IP, " is sending message of type ", msg.Type, " to ", IP)
	clnt, err := rpc.Dial("tcp", IP)
	reply := message.Message{}
	if err != nil {
		system.Println("Error dialing RPC", msg.Type)
		reply.Type = EMPTY
		return reply
	}
	err = clnt.Call("Central.HandleIncomingMessage", msg, &reply)
	if err != nil {
		system.Println("Error callling RPC", msg.Type)
		reply.Type = EMPTY
		return reply
	}
	system.Println("Client ", client.IP ," received reply from ", msg.From)
	return reply
}

func (client *Client) PrintClientList(){
	system.Println(client.Clientlist)
}

func (client *Client) PrintCentralIP(){
	system.Println(client.ServerIP)
}