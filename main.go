package main

import (
        "fmt"
        "ivy/central"
        "ivy/client"
        "net"
        "net/rpc"
        "os"
        "strconv"
        "time"

        "github.com/fatih/color"
)

// Color coded logs
var system = color.New(color.FgCyan).Add(color.BgBlack)

/*
Show a list of options to choose from.
*/
func showmenu() {
        system.Println("********************************")
        system.Println("\t\tMENU")
        system.Println("Press 1 to see the client list")
        system.Println("Press 2 for something")
        system.Println("Press m to see the menu")
        system.Println("********************************")
}

func main() {
        // get port from cli arguments (specified by user)
        helper := ""             // IP address of the port number we are using to join. Will be specified iff I am not the central manager.
        central_manager := false // to check if you are the primary or not
        var port int
        // var joinerPort string
        for i, arg := range os.Args {
                switch arg {
                case "-u":
                        // Specified if you are the client, and you want to get the clientlist metadata from the central manager
                        if i+1 >= len(os.Args) {
                                system.Println("Enter valid helper port!!")
                        }
                        helper = os.Args[i+1]
                case "-cm":
                        // Specified if you are the central manager
                        central_manager = true
                case "-bcm":
                        // Specified if you are the backup central manager
                        central_manager = false
                default:

                }
        }

        // Create new Node object for yourself
        if helper != "" { // case when you are a client
                me := client.Client{}
                me.ServerIP = helper
                system.Println("Joining using ", me.ServerIP)
                port, _ = GetFreePort()
                addr := GetOutboundIP().String() + ":" + strconv.Itoa(port)
                me.IP = addr
                system.Println("My IP is:", me.IP)
                // Bind yourself to a port and listen to it
                tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
                if err != nil {
                        system.Println("Error resolving TCP address", err)
                }
                inbound, err := net.ListenTCP("tcp", tcpAddr)
                if err != nil {
                        system.Println("Could not listen to TCP address", err)
                }

                // Register RPC methods and accept incoming requests
                rpc.Register(&me)
                system.Println("Client is runnning at IP address: ", tcpAddr)
                go rpc.Accept(inbound)
                me.JoinNetwork(helper)

                // Keep the parent thread alive
                for {
                        time.Sleep(5 * time.Second)
                        system.Println("Alive")
                        var input string
                        fmt.Scanln(&input)
                        switch input {
                        case "1":
                                system.Println("Clientlist requested")
                                me.PrintClientList()
                        case "2":
                                system.Println("Server IP Requested")
                                me.PrintCentralIP()
                        case "3":
                                me.ReadRequest()
                        default:
                                system.Println("Enter valid input")
                        }
                }
        } else { //  case when you are a cm, or backup cm
                me := central.Central{}

                if central_manager {
                        me.IsPrimary = true
                }
                port, _ = GetFreePort()
                addr := GetOutboundIP().String() + ":" + strconv.Itoa(port)
                me.IP = addr
                system.Println("My IP is:", me.IP)
                // Bind yourself to a port and listen to it
                tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
                if err != nil {
                        system.Println("Error resolving TCP address", err)
                }
                inbound, err := net.ListenTCP("tcp", tcpAddr)
                if err != nil {
                        system.Println("Could not listen to TCP address", err)
                }

                // Register RPC methods and accept incoming requests
                rpc.Register(&me)
                system.Println("Central manager is runnning at IP address:", tcpAddr)
                go rpc.Accept(inbound)
                me.CreateNetwork()
                showmenu()
                // Keep the parent thread alive
                for {
                        time.Sleep(5 * time.Second)
                        system.Println("Alive")
                        var input string
                        fmt.Scanln(&input)
                        switch input {
                        case "1":
                                system.Println("Clientlist requested")
                                me.PrintClientList()
                        default:
                                system.Println("Enter valid input")
                        }
                }
        }
}