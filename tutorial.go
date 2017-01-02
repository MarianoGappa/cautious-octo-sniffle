package main

import (
	"time"

	"github.com/Shopify/sarama"
)

func tutorial() chan *sarama.ConsumerMessage {
	c := make(chan *sarama.ConsumerMessage)

	go pushTutorialMessages(c)

	return c
}

func pushTutorialMessages(c chan *sarama.ConsumerMessage) {
	for _, e := range tutorialEvents() {
		switch e.kind {
		case "sleep":
			time.Sleep(e.duration)
		case "message":
			c <- e.message
		}
	}
}

func tutorialEvents() []tutorialEvent {
	return []tutorialEvent{
		{kind: "message", message: &sarama.ConsumerMessage{Topic: "one", Value: []byte("{}")}},
		{kind: "sleep", duration: 5 * time.Second},
		{kind: "message", message: &sarama.ConsumerMessage{Topic: "one", Value: []byte("{}")}},
		{kind: "sleep", duration: 5 * time.Second},
		{kind: "message", message: &sarama.ConsumerMessage{Topic: "one", Value: []byte("{}")}},
		{kind: "sleep", duration: 5 * time.Second},
	}
}

type tutorialEvent struct {
	kind     string
	duration time.Duration
	message  *sarama.ConsumerMessage
}
