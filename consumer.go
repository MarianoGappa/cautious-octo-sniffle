package main

import (
	"log"
	"strconv"
	"sync"

	"github.com/Shopify/sarama"
)

func startConsumers(config *Config, c chan *sarama.ConsumerMessage, mainQuit chan struct{}, reqQuit chan struct{}) {
	for _, consumerConfig := range config.Consumers {
		topic, broker, partition := consumerConfig.Topic, consumerConfig.Broker, consumerConfig.Partition
		var offset int64 = -1

		if numericOffset, err := strconv.ParseInt(consumerConfig.Offset, 10, 64); err == nil {
			offset = numericOffset
		} else {
			switch consumerConfig.Offset {
			case "oldest":
				offset = -2
			case "newest":
				offset = -1
			default:
				log.Println("Invalid value for consumer offset")
				close(reqQuit)
				return
			}
		}

		go consume(c, mainQuit, reqQuit, topic, broker, partition, offset)
	}
}

func consume(c chan *sarama.ConsumerMessage, mainQuit chan struct{}, reqQuit chan struct{}, topic string, broker string, partition int, offset int64) {
	if topic == "" {
		log.Println("Please define topic name for your consumer")
		close(reqQuit)
		return
	}

	if c == nil {
		log.Println("Channel is not initialised")
		close(reqQuit)
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
		close(reqQuit)
		return
	}

	var partitions []int32
	if partition == -1 {
		partitions, err = consumer.Partitions(topic)
		if err != nil {
			log.Println(err)
			close(reqQuit)
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
				case <-mainQuit:
					close(reqQuit)
				case <-reqQuit:
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
