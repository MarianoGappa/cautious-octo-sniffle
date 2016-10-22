package main

import (
	"log"
	"net"
)

func mustGetListener(port int) *net.TCPListener {
	listener, err := newListener(port)
	if err != nil {
		log.Fatalf("Could not open listener on port %v. err=%v", port, err)
	}
	return listener
}

func newListener(port int) (*net.TCPListener, error) {
	listener, err := net.ListenTCP("tcp", &net.TCPAddr{
		IP:   net.ParseIP("localhost").To4(),
		Port: port,
	})
	if err != nil {
		return listener, err
	}
	return listener, nil
}
