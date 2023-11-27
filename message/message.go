package message


import (

)


type Message struct{
	Type string // REQFORWARD | READ | WRITE | 
	From string
	To string
	Clientlist []string // List of clients, used for when you join the network
}