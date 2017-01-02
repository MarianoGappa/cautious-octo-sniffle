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
	es := tutorialEvents()
	i := 0
	for {
		if i == len(es) {
			break
		}
		switch es[i].kind {
		case "sleep":
			time.Sleep(es[i].duration)
		case "message":
			c <- es[i].message
		}
		if es[i].repeat > 0 {
			es[i].repeat -= 1
			i += es[i].jumpRelative
			continue
		}
		i++
	}
}

type tutorialEvent struct {
	kind         string
	duration     time.Duration
	jumpRelative int
	repeat       int
	message      *sarama.ConsumerMessage
}

func tutorialEvents() []tutorialEvent {
	return []tutorialEvent{
		{kind: "message", message: &sarama.ConsumerMessage{Topic: "one", Value: []byte("{}")}},
		{kind: "sleep", duration: 1 * time.Second},
		{kind: "message", message: &sarama.ConsumerMessage{Topic: "one", Value: []byte("{}")}},
		{kind: "sleep", duration: 1 * time.Second},
		{kind: "message", message: &sarama.ConsumerMessage{Topic: "one", Value: []byte("{}")}},
		{kind: "sleep", duration: 1 * time.Second, jumpRelative: -5, repeat: 2},
	}
}
