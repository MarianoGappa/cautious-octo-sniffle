package main

import (
	"github.com/Shopify/sarama"
	"log"
	"sync"
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

	var partitions []int32
	if (partition == -1) {
		partitions, err = consumer.Partitions(topic)
		if err != nil {
			panic(err)
		}
	} else {
		partitions = append(partitions, int32(partition))
	}

	var wg sync.WaitGroup

consuming:
	for partition := range partitions {
		partitionConsumer, err := consumer.ConsumePartition(topic, int32(partition), 0)
		if err != nil {
			log.Printf("Failed to consume partition %v err=%v\n", partition, err)
			continue consuming
		}
		wg.Add(1)

		go func(pc sarama.PartitionConsumer) {
			for {
				select {
				case msg := <-partitionConsumer.Messages():
					c <- msg
				case <-quit:
					partitionConsumer.Close()
					wg.Done()
					return
				}
			}
		}(partitionConsumer)
	}
	wg.Wait()
	consumer.Close()
}
