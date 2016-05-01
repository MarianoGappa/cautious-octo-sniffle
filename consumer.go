package main

import (
	"github.com/Shopify/sarama"
	"log"
	"sync"
)

func consume(c chan *sarama.ConsumerMessage, quit chan bool, topic string, broker string, partition int, offset int64) {
	if topic == "" {
		log.Println("Please define topic name for your consumer")
		quit <- true
		return
	}

	if c == nil {
		log.Println("Channel is not initialised")
		quit <- true
		return
	}

	if broker == "" {
		broker = "localhost:9092"
	}

	if offset == 0 {
		offset = sarama.OffsetNewest
	}

	consumer, err := sarama.NewConsumer([]string{broker}, nil)
	if err != nil {
		log.Println(err)
		quit <- true
		return
	}

	var partitions []int32
	if partition == -1 {
		partitions, err = consumer.Partitions(topic)
		if err != nil {
			log.Println(err)
			quit <- true
			return
		}
	} else {
		partitions = append(partitions, int32(partition))
	}

	var wg sync.WaitGroup

consuming:
	for _, partition := range partitions {
		partitionConsumer, err := consumer.ConsumePartition(topic, int32(partition), offset)
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
