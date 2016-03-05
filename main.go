package main

import (
	"github.com/Shopify/sarama"
	"log"
	"os"
	"os/signal"
)

func main() {
	c := make(chan *sarama.ConsumerMessage)

	go consume(c, "new_topic", "", 0, -2)

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	consumed := 0

MainLoop:
	for {
		select {
		case message := <-c:
			log.Printf("Consumed message offset %d\n", message.Offset)
			consumed++

		case <-signals:
			break MainLoop
		}
	}

	log.Printf("Consumed: %d\n", consumed)
}
