/*
This code is intended to handle the death of the central manager/ backup manager. We will wait for a kill signal from either the backup manager,
or the central managers terminal, then we will send a message with all the details about the central manager to the backup manager, or vice versa.
The backup manager will then use the information that it receives from a call here, then it will contact all the clients in its clientlist, to tell
them that I am the new server. The clients will set their new serverIP as the other server.
*/
package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fatih/color"
	"github.com/fauzxan/ivy/central"
)

var codered = color.New(color.FgHiRed).Add(color.BgBlack)
 

func SetupCloseHandler(cm *central.Central) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		codered.Println("\n\n\nSERVER IS GOING DOWN!!!! FLUSH DATA TO BACKUP!!!!")
		cm.FlushData()
		time.Sleep(1 * time.Second)
		os.Exit(0)
	}()
}