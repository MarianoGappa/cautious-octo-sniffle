package main

import (
	// "flag"
	// "fmt"
	"log"
	"os"
	"os/signal"
	// "strconv"
	// "strings"
	// "sync"
	// "time"

	"github.com/Shopify/sarama"
)

func consume(c chan *sarama.ConsumerMessage, topic string, broker string, partition int, offset int64) {
	if topic == "" {
		panic("Please define topic name for your consumer")
	}

	if c == nil {
		panic("Channel is not initialised")
	}

	if broker == "" {
		broker = "localhost:9092"
	}

	if offset == 0 {
		offset = sarama.OffsetNewest
	}

	consumer, err := sarama.NewConsumer([]string{broker}, nil)
	if err != nil {
		panic(err)
	}

	defer func() {
		if err := consumer.Close(); err != nil {
			log.Fatalln(err)
		}
	}()

	partitionConsumer, err := consumer.ConsumePartition(topic, 0, offset)
	if err != nil {
		panic(err)
	}

	defer func() {
		if err := partitionConsumer.Close(); err != nil {
			log.Fatalln(err)
		}
	}()

	// Trap SIGINT to trigger a shutdown.
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

ConsumerLoop:
	for {
		select {
		case msg := <-partitionConsumer.Messages():
			c <- msg
		case <-signals:
			break ConsumerLoop
		}
	}

}
