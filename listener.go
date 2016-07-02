package main

import (
	"log"
	"net"
)

func mustGetListener(port int, q chan struct{}) *net.TCPListener {
	listener, err := newListener(port, q)
	if err != nil {
		log.Fatalf("Could not open listener on port %v. err=%v", port, err)
	}
	return listener
}

func newListener(port int, q chan struct{}) (*net.TCPListener, error) {
	listener, err := net.ListenTCP("tcp", &net.TCPAddr{
		IP:   net.ParseIP("localhost").To4(),
		Port: port,
	})
	if err != nil {
		return listener, err
	}

	go func(listener *net.TCPListener, q chan struct{}) {
		<-q
		log.Println("Received quit signal; closing tcp listener.")
		listener.Close()
		log.Println("closed tcp listener")
	}(listener, q)

	return listener, nil
}
