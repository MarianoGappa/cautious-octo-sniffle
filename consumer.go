package main

import (
	"fmt"
	"log"
	"time"

	"encoding/json"

	"github.com/Shopify/sarama"
	"golang.org/x/net/websocket"
)

type message struct {
	Key       string                 `json:"key"`
	Value     map[string]interface{} `json:"value"`
	Topic     string                 `json:"topic"`
	Partition int32                  `json:"partition"`
	Offset    int64                  `json:"offset"`
	Timestamp time.Time              `json:"timestamp"` // only set if kafka is version 0.10+
	Count     int64                  // only for bookie counts
	FSMId     string                 // only for bookie counts
}

type iSender interface {
	Send(*websocket.Conn, string) error
}

type sender struct{}

func (s sender) Send(ws *websocket.Conn, msg string) error {
	return websocket.Message.Send(ws, msg)
}

func process(ws *websocket.Conn, c chan *sarama.ConsumerMessage, sender iSender, rules []rule, globalFSMId string, uuid string, bookieCounts map[string]int64) {
	ticker := time.NewTicker(time.Millisecond * 100)

	buffer := []message{}
	for t, c := range bookieCounts {
		buffer = append(buffer, message{Count: c, Topic: t, FSMId: globalFSMId})
	}

	fsmIdAliases := map[string]string{}
	sendSuccess("Starting to send messages!", ws)

	hbCh := make(chan struct{})
	go processHeartbeats(wsReceiver{ws: ws}, hbCh, uuid, 10*time.Second)

	for {
		select {
		case cMsg := <-c:
			m, err := newMessage(*cMsg)
			if err != nil {
				sendError(fmt.Sprintf("Could not parse %v into message", err), ws)
			}
			if m.Timestamp.UnixNano() <= 0 {
				m.Timestamp = time.Now()
			}
			buffer = append(buffer, m)
		case <-ticker.C:
			events := []event{}
			incompleteEvents := []event{}
			for i := 0; len(buffer) > 0 && i < 1000; i++ {
				err := processMessage(buffer[0], rules, fsmIdAliases, &events, &incompleteEvents, globalFSMId)
				if err != nil {
					sendError(fmt.Sprintf("Error while processing message: err=%v", err), ws)
					break
				}
				buffer = buffer[1:]
			}

			for _, ie := range incompleteEvents {
				events = aggregate(events, ie, ie.Aggregate, globalFSMId)
			}

			if len(events) == 0 {
				break
			}

			byt, err := json.Marshal(events)
			if err != nil {
				sendError(fmt.Sprintf("Error while marshalling events: err=%v\n", err), ws)
				continue
			}

			err = sender.Send(ws, string(byt))
			if err != nil {
				log.Printf("Error while trying to send to WebSocket: err=%v\n", err)
				return
			}
		case <-hbCh:
			sendError("Timing out due to heartbeat not received.", ws)
			return
		}
	}
}

func newMessage(cm sarama.ConsumerMessage) (message, error) {
	var v interface{}
	if err := json.Unmarshal(cm.Value, &v); err != nil {
		return message{}, err
	}

	return message{
		Key:       string(cm.Key),
		Value:     v.(map[string]interface{}),
		Topic:     cm.Topic,
		Partition: cm.Partition,
		Offset:    cm.Offset,
		Timestamp: cm.Timestamp,
	}, nil
}

func sliceInsert(slice []message, index int, value message) []message {
	if index == 0 {
		return append([]message{value}, slice...)
	}
	if index >= len(slice) {
		return append(slice, value)
	}
	// Grow the slice by one element.
	slice = slice[0 : len(slice)+1]
	// Use copy to move the upper part of the slice out of the way and open a hole.
	copy(slice[index+1:], slice[index:])
	// Store the new value.
	slice[index] = value
	// Return the result.
	return slice
}
