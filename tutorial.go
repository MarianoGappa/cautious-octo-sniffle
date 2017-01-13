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
		{kind: "message", message: &sarama.ConsumerMessage{Topic: "tutorial.messages", Value: []byte(`{"log":"Hi! Thank you for trying Flowbro."}`)}},
		{kind: "sleep", duration: 8 * time.Second},
		{kind: "message", message: &sarama.ConsumerMessage{Topic: "tutorial.messages", Value: []byte(`{"log":"With Flowbro, you can better visualise what your distributed system is doing. Like this:"}`)}},
		{kind: "sleep", duration: 10 * time.Second},
		{kind: "message", message: &sarama.ConsumerMessage{Topic: "requests", Value: []byte(`{"target":"phone","message":"Lorem ipsum"}`)}},
		{kind: "message", message: &sarama.ConsumerMessage{Topic: "notifications", Value: []byte(`{"target":"phone","message":"Lorem ipsum"}`)}},
		{kind: "sleep", duration: 500 * time.Millisecond},
		{kind: "message", message: &sarama.ConsumerMessage{Topic: "requests", Value: []byte(`{"target":"desktop","message":"Lorem ipsum"}`)}},
		{kind: "message", message: &sarama.ConsumerMessage{Topic: "notifications", Value: []byte(`{"target":"desktop","message":"Lorem ipsum"}`)}},
		{kind: "sleep", duration: 500 * time.Millisecond},
		{kind: "message", message: &sarama.ConsumerMessage{Topic: "request", Value: []byte(`{"target":"tablet","message":"Lorem ipsum"}`)}},
		{kind: "message", message: &sarama.ConsumerMessage{Topic: "notifications", Value: []byte(`{"target":"tablet","message":"Lorem ipsum"}`)}},
		{kind: "sleep", duration: 500 * time.Millisecond, jumpRelative: -5, repeat: 8},

		{kind: "message", message: &sarama.ConsumerMessage{Topic: "tutorial.messages", Value: []byte(`{"log":"1) Monitoring (in real-time)"}`)}},
		{kind: "sleep", duration: 10 * time.Second},
		{kind: "message", message: &sarama.ConsumerMessage{Topic: "requests", Value: []byte(`{"target":"phone","message":"Lorem ipsum"}`)}},
		{kind: "message", message: &sarama.ConsumerMessage{Topic: "notifications", Value: []byte(`{"target":"phone","message":"Lorem ipsum"}`)}},
		{kind: "sleep", duration: 500 * time.Millisecond},
		{kind: "message", message: &sarama.ConsumerMessage{Topic: "requests", Value: []byte(`{"target":"desktop","message":"Lorem ipsum"}`)}},
		{kind: "message", message: &sarama.ConsumerMessage{Topic: "notifications", Value: []byte(`{"target":"desktop","message":"Lorem ipsum"}`)}},
		{kind: "sleep", duration: 500 * time.Millisecond},
		{kind: "message", message: &sarama.ConsumerMessage{Topic: "request", Value: []byte(`{"target":"tablet","message":"Lorem ipsum"}`)}},
		{kind: "message", message: &sarama.ConsumerMessage{Topic: "notifications", Value: []byte(`{"target":"tablet","message":"Lorem ipsum"}`)}},
		{kind: "sleep", duration: 500 * time.Millisecond, jumpRelative: -5, repeat: 8},

		{kind: "message", message: &sarama.ConsumerMessage{Topic: "tutorial.messages", Value: []byte(`{"log":"2) Debugging & support (in real-time or after the fact)"}`)}},
		{kind: "sleep", duration: 10 * time.Second},
		{kind: "message", message: &sarama.ConsumerMessage{Topic: "requests", Value: []byte(`{"target":"phone","message":"Lorem ipsum"}`)}},
		{kind: "message", message: &sarama.ConsumerMessage{Topic: "notifications", Value: []byte(`{"target":"phone","message":"Lorem ipsum"}`)}},
		{kind: "sleep", duration: 500 * time.Millisecond},
		{kind: "message", message: &sarama.ConsumerMessage{Topic: "requests", Value: []byte(`{"target":"desktop","message":"Lorem ipsum"}`)}},
		{kind: "message", message: &sarama.ConsumerMessage{Topic: "notifications", Value: []byte(`{"target":"desktop","message":"Lorem ipsum"}`)}},
		{kind: "sleep", duration: 500 * time.Millisecond},
		{kind: "message", message: &sarama.ConsumerMessage{Topic: "request", Value: []byte(`{"target":"tablet","message":"Lorem ipsum"}`)}},
		{kind: "message", message: &sarama.ConsumerMessage{Topic: "notifications", Value: []byte(`{"target":"tablet","message":"Lorem ipsum"}`)}},
		{kind: "sleep", duration: 500 * time.Millisecond, jumpRelative: -5, repeat: 8},

		{kind: "message", message: &sarama.ConsumerMessage{Topic: "tutorial.messages", Value: []byte(`{"log":"3) Making effective presentations"}`)}},
		{kind: "sleep", duration: 10 * time.Second},
		{kind: "message", message: &sarama.ConsumerMessage{Topic: "requests", Value: []byte(`{"target":"phone","message":"Lorem ipsum"}`)}},
		{kind: "message", message: &sarama.ConsumerMessage{Topic: "notifications", Value: []byte(`{"target":"phone","message":"Lorem ipsum"}`)}},
		{kind: "sleep", duration: 500 * time.Millisecond},
		{kind: "message", message: &sarama.ConsumerMessage{Topic: "requests", Value: []byte(`{"target":"desktop","message":"Lorem ipsum"}`)}},
		{kind: "message", message: &sarama.ConsumerMessage{Topic: "notifications", Value: []byte(`{"target":"desktop","message":"Lorem ipsum"}`)}},
		{kind: "sleep", duration: 500 * time.Millisecond},
		{kind: "message", message: &sarama.ConsumerMessage{Topic: "request", Value: []byte(`{"target":"tablet","message":"Lorem ipsum"}`)}},
		{kind: "message", message: &sarama.ConsumerMessage{Topic: "notifications", Value: []byte(`{"target":"tablet","message":"Lorem ipsum"}`)}},
		{kind: "sleep", duration: 500 * time.Millisecond, jumpRelative: -5, repeat: 8},

		{kind: "message", message: &sarama.ConsumerMessage{Topic: "tutorial.messages", Value: []byte(`{"log":"Flowbro translates Kafka messages to visual interactions and logs; the mapping is 1 to n, as defined by configured rules."}`)}},
		{kind: "sleep", duration: 10 * time.Second},
		{kind: "message", message: &sarama.ConsumerMessage{Topic: "requests", Value: []byte(`{"target":"phone","message":"Lorem ipsum"}`)}},
		{kind: "message", message: &sarama.ConsumerMessage{Topic: "notifications", Value: []byte(`{"target":"phone","message":"Lorem ipsum"}`)}},
		{kind: "sleep", duration: 500 * time.Millisecond},
		{kind: "message", message: &sarama.ConsumerMessage{Topic: "requests", Value: []byte(`{"target":"desktop","message":"Lorem ipsum"}`)}},
		{kind: "message", message: &sarama.ConsumerMessage{Topic: "notifications", Value: []byte(`{"target":"desktop","message":"Lorem ipsum"}`)}},
		{kind: "sleep", duration: 500 * time.Millisecond},
		{kind: "message", message: &sarama.ConsumerMessage{Topic: "request", Value: []byte(`{"target":"tablet","message":"Lorem ipsum"}`)}},
		{kind: "message", message: &sarama.ConsumerMessage{Topic: "notifications", Value: []byte(`{"target":"tablet","message":"Lorem ipsum"}`)}},
		{kind: "sleep", duration: 500 * time.Millisecond, jumpRelative: -5, repeat: 8},

		{kind: "message", message: &sarama.ConsumerMessage{Topic: "tutorial.messages", Value: []byte(`{"log":"Some tips:"}`)}},
		{kind: "sleep", duration: 10 * time.Second},
		{kind: "message", message: &sarama.ConsumerMessage{Topic: "requests", Value: []byte(`{"target":"phone","message":"Lorem ipsum"}`)}},
		{kind: "message", message: &sarama.ConsumerMessage{Topic: "notifications", Value: []byte(`{"target":"phone","message":"Lorem ipsum"}`)}},
		{kind: "sleep", duration: 500 * time.Millisecond},
		{kind: "message", message: &sarama.ConsumerMessage{Topic: "requests", Value: []byte(`{"target":"desktop","message":"Lorem ipsum"}`)}},
		{kind: "message", message: &sarama.ConsumerMessage{Topic: "notifications", Value: []byte(`{"target":"desktop","message":"Lorem ipsum"}`)}},
		{kind: "sleep", duration: 500 * time.Millisecond},
		{kind: "message", message: &sarama.ConsumerMessage{Topic: "request", Value: []byte(`{"target":"tablet","message":"Lorem ipsum"}`)}},
		{kind: "message", message: &sarama.ConsumerMessage{Topic: "notifications", Value: []byte(`{"target":"tablet","message":"Lorem ipsum"}`)}},
		{kind: "sleep", duration: 500 * time.Millisecond, jumpRelative: -5, repeat: 8},

		{kind: "message", message: &sarama.ConsumerMessage{Topic: "tutorial.messages", Value: []byte(`{"log":"1) use 'aggregate: true' to group messages together and save some CPU, like this:"}`)}},
		{kind: "sleep", duration: 10 * time.Second},
		{kind: "message", message: &sarama.ConsumerMessage{Topic: "requests", Value: []byte(`{"target":"phone","message":"Lorem ipsum"}`)}},
		{kind: "message", message: &sarama.ConsumerMessage{Topic: "requests", Value: []byte(`{"target":"phone","message":"Lorem ipsum"}`)}},
		{kind: "message", message: &sarama.ConsumerMessage{Topic: "requests", Value: []byte(`{"target":"phone","message":"Lorem ipsum"}`)}},
		{kind: "message", message: &sarama.ConsumerMessage{Topic: "notifications", Value: []byte(`{"target":"phone","message":"Lorem ipsum"}`)}},
		{kind: "message", message: &sarama.ConsumerMessage{Topic: "notifications", Value: []byte(`{"target":"phone","message":"Lorem ipsum"}`)}},
		{kind: "message", message: &sarama.ConsumerMessage{Topic: "notifications", Value: []byte(`{"target":"phone","message":"Lorem ipsum"}`)}},
		{kind: "sleep", duration: 500 * time.Millisecond},
		{kind: "message", message: &sarama.ConsumerMessage{Topic: "requests", Value: []byte(`{"target":"desktop","message":"Lorem ipsum"}`)}},
		{kind: "message", message: &sarama.ConsumerMessage{Topic: "requests", Value: []byte(`{"target":"desktop","message":"Lorem ipsum"}`)}},
		{kind: "message", message: &sarama.ConsumerMessage{Topic: "notifications", Value: []byte(`{"target":"desktop","message":"Lorem ipsum"}`)}},
		{kind: "message", message: &sarama.ConsumerMessage{Topic: "notifications", Value: []byte(`{"target":"desktop","message":"Lorem ipsum"}`)}},
		{kind: "sleep", duration: 500 * time.Millisecond},
		{kind: "message", message: &sarama.ConsumerMessage{Topic: "request", Value: []byte(`{"target":"tablet","message":"Lorem ipsum"}`)}},
		{kind: "message", message: &sarama.ConsumerMessage{Topic: "request", Value: []byte(`{"target":"tablet","message":"Lorem ipsum"}`)}},
		{kind: "message", message: &sarama.ConsumerMessage{Topic: "request", Value: []byte(`{"target":"tablet","message":"Lorem ipsum"}`)}},
		{kind: "message", message: &sarama.ConsumerMessage{Topic: "request", Value: []byte(`{"target":"tablet","message":"Lorem ipsum"}`)}},
		{kind: "message", message: &sarama.ConsumerMessage{Topic: "notifications", Value: []byte(`{"target":"tablet","message":"Lorem ipsum"}`)}},
		{kind: "message", message: &sarama.ConsumerMessage{Topic: "notifications", Value: []byte(`{"target":"tablet","message":"Lorem ipsum"}`)}},
		{kind: "message", message: &sarama.ConsumerMessage{Topic: "notifications", Value: []byte(`{"target":"tablet","message":"Lorem ipsum"}`)}},
		{kind: "message", message: &sarama.ConsumerMessage{Topic: "notifications", Value: []byte(`{"target":"tablet","message":"Lorem ipsum"}`)}},
		{kind: "sleep", duration: 500 * time.Millisecond, jumpRelative: -5, repeat: 20},

		{kind: "message", message: &sarama.ConsumerMessage{Topic: "tutorial.messages", Value: []byte(`{"log":"2) use 'noJSON: true' to not see it below this message"}`)}},
		{kind: "sleep", duration: 10 * time.Second},
		{kind: "message", message: &sarama.ConsumerMessage{Topic: "requests", Value: []byte(`{"target":"phone","message":"Lorem ipsum"}`)}},
		{kind: "message", message: &sarama.ConsumerMessage{Topic: "notifications", Value: []byte(`{"target":"phone","message":"Lorem ipsum"}`)}},
		{kind: "sleep", duration: 500 * time.Millisecond},
		{kind: "message", message: &sarama.ConsumerMessage{Topic: "requests", Value: []byte(`{"target":"desktop","message":"Lorem ipsum"}`)}},
		{kind: "message", message: &sarama.ConsumerMessage{Topic: "notifications", Value: []byte(`{"target":"desktop","message":"Lorem ipsum"}`)}},
		{kind: "sleep", duration: 500 * time.Millisecond},
		{kind: "message", message: &sarama.ConsumerMessage{Topic: "request", Value: []byte(`{"target":"tablet","message":"Lorem ipsum"}`)}},
		{kind: "message", message: &sarama.ConsumerMessage{Topic: "notifications", Value: []byte(`{"target":"tablet","message":"Lorem ipsum"}`)}},
		{kind: "sleep", duration: 500 * time.Millisecond, jumpRelative: -5, repeat: 8},

		{kind: "message", message: &sarama.ConsumerMessage{Topic: "tutorial.messages", Value: []byte(`{"log":"3) learn about fsmId and fsmIdAlias to color-code, filter and track the status of the underlying processes (or FSMs) that your system supports."}`)}},
	}
}
