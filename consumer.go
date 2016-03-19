package main

import (
	"github.com/Shopify/sarama"
	"log"
)

func consume(c chan *sarama.ConsumerMessage, quit chan bool, topic string, broker string, partition int, offset int64) {
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

	for {
		select {
		case msg := <-partitionConsumer.Messages():
			c <- msg
		case <-quit:
			return
		}
	}
}
