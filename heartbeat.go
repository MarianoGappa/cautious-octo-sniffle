package main

import (
	"time"

	"io"

	log "github.com/Sirupsen/logrus"
	"golang.org/x/net/websocket"
)

type heartbeat struct {
	UUID string `json:"uuid"`
}

func processHeartbeats(wr wsRecv, out chan struct{}, uuid string, timeoutDuration time.Duration) {
	hbCh := make(chan struct{})
	timeout := time.NewTimer(timeoutDuration)

	go readHeartbeats(wr, hbCh, uuid)

	for {
		select {
		case <-timeout.C:
			close(out)
			return
		case <-hbCh:
			timeout.Reset(timeoutDuration)
		}
	}
}

func readHeartbeats(wr wsRecv, out chan struct{}, uuid string) {
	for {
		hb, err := wr.recv()
		if err == io.EOF {
			return
		}
		if err != nil {
			log.WithFields(log.Fields{"err": err}).Error("Error while reading heartbeat.")
			return
		}

		if hb.UUID == uuid {
			out <- struct{}{}
		}
	}
}

type wsRecv interface {
	recv() (heartbeat, error)
}

type wsReceiver struct {
	ws *websocket.Conn
}

func (wr wsReceiver) recv() (heartbeat, error) {
	var hb heartbeat
	err := websocket.JSON.Receive(wr.ws, &hb)
	return hb, err
}
