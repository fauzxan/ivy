package message

import ()

type Message struct {
        Type       string       // Multiple types
        From       string
        To         string
        Clientlist []string     // List of clients, used for when you join the network
        PageId     int
        Content    int          // content of the page
}