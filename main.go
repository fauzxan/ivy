package main

import (
	"fmt"
	"net"
	"net/rpc"
	"os"
	"strconv"

	"github.com/fauzxan/ivy/central"
	"github.com/fauzxan/ivy/client"

	// "time"

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
        backup := ""
        reboot_address := ""
        doIPing := false
        doesTheOtherPing := false
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
                        doIPing = false // If I am the central, then I don't ping, the other does. 
                        doesTheOtherPing = true
                        if i+1 >= len(os.Args){
                                system.Println("There is no backup!")
                        } else{
                                backup = os.Args[i+1]
                                system.Println("There is a backup! IP of backup is: ", backup)
                                doIPing = true // If I am backup then I ping, the other doesn't (main)
                                doesTheOtherPing = false
                        }
                case "-r":
                        if i+1 >= len(os.Args){
                                system.Println("System is rebooting!! But not enough arguments!")
                        } else{
                                reboot_address = os.Args[i+1]
                                system.Println("Will get up and running again on: ", reboot_address)
                        }
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
                        case "4":
                                me.WriteRequest()
                        case "5":
                                me.PrintPages()
                        default:
                                system.Println("Enter valid input")
                        }
                }
        } else { //  case when you are a cm, or backup cm
                me := central.Central{}
                var addr string
                if reboot_address == ""{
                        port, _ = GetFreePort()
                        addr = GetOutboundIP().String() + ":" + strconv.Itoa(port)
                } else {
                        addr = reboot_address
                }

                me.DoIPing = doIPing
                me.DoesTheOtherPing = doesTheOtherPing
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
                go me.CreateNetwork(backup)
                go SetupCloseHandler(&me)
                showmenu()
                // Keep the parent thread alive
                for {
                        system.Println("Alive")
                        var input string
                        fmt.Scanln(&input)
                        switch input {
                        case "1":
                                system.Println("Clientlist requested")
                                me.PrintClientList()
                        case "2":
                                system.Println("Records requested")
                                me.PrintRecords()
                        default:
                                system.Println("Enter valid input")
                        }
                }
        }
}