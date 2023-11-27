package main

import (
	"log"
	"net"
)

/*
	Function to automatically get the outbound IP without user input in .env file
*/
func GetOutboundIP() net.IP {
    conn, err := net.Dial("udp", "8.8.8.8:80")
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()

    localAddr := conn.LocalAddr().(*net.UDPAddr)

    return localAddr.IP
}

/*
	Function to get a port number that is currently not in use
*/
func GetFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
			return 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
			return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}