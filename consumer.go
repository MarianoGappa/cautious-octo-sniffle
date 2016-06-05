package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/Shopify/sarama"
	"golang.org/x/net/websocket"
)

func setupConsumers(config *Config) ([]<-chan *sarama.ConsumerMessage, error) {
	partitionConsumers := []<-chan *sarama.ConsumerMessage{}
	for _, consumerConfig := range config.Consumers {
		topic, brokerString, partition := consumerConfig.Topic, consumerConfig.Broker, consumerConfig.Partition
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
				return nil, fmt.Errorf("Invalid value for consumer offset")
			}
		}

		if topic == "" {
			return nil, fmt.Errorf("Please define topic name for your consumer")
		}

		if brokerString == "" {
			brokerString = "localhost:9092"
		}

		if offset == 0 {
			offset = sarama.OffsetNewest
		}

		brokers := strings.Split(brokerString, ",")
		consumer, err := sarama.NewConsumer(brokers, nil)
		if err != nil {
			return nil, fmt.Errorf("Error creating consumer. err=%v", err)
		}

		var partitions []int32
		if partition == -1 {
			partitions, err = consumer.Partitions(topic)
			if err != nil {
				return nil, fmt.Errorf("Error fetching partitions for topic. err=%v", err)
			}
		} else {
			partitions = append(partitions, int32(partition))
		}

		for _, partition := range partitions {
			partitionConsumer, err := consumer.ConsumePartition(topic, int32(partition), offset)
			if err != nil {
				return nil, fmt.Errorf("Failed to consume partition %v err=%v\n", partition, err)
			}

			partitionConsumers = append(partitionConsumers, partitionConsumer.Messages())
		}
	}
	return partitionConsumers, nil
}

func demuxMessages(pc []<-chan *sarama.ConsumerMessage, q chan struct{}) chan *sarama.ConsumerMessage {
	c := make(chan *sarama.ConsumerMessage)
	for _, p := range pc {
		go func(p <-chan *sarama.ConsumerMessage) {
			for {
				select {
				case msg := <-p:
					c <- msg
				case <-q:
					return
				}
			}
		}(p)
	}
	return c
}

func sendMessagesToWsBlocking(ws *websocket.Conn, c chan *sarama.ConsumerMessage, q chan struct{}) {
	for {
		select {
		case cMsg := <-c:
			msg :=
				"{\"topic\": \"" + cMsg.Topic +
					"\", \"partition\": \"" + strconv.FormatInt(int64(cMsg.Partition), 10) +
					"\", \"offset\": \"" + strconv.FormatInt(cMsg.Offset, 10) +
					"\", \"key\": \"" + strings.Replace(string(cMsg.Key), `"`, `\"`, -1) +
					"\", \"value\": \"" + strings.Replace(string(cMsg.Value), `"`, `\"`, -1) +
					"\", \"consumedUnixTimestamp\": \"" + strconv.FormatInt(time.Now().Unix(), 10) +
					"\"}\n"

			log.Println("Sending message to WebSocket: " + msg)
			err := websocket.Message.Send(ws, msg)
			if err != nil {
				log.Println("Error while trying to send to WebSocket: ", err)
				err := ws.Close()
				if err != nil {
					log.Println("Error while closing WebSocket!: ", err)
				} else {
					log.Println("Closed WebSocket connection given quit signal.")
				}
				return
			}
		case <-q:
			err := ws.Close()
			if err != nil {
				log.Println("Error while closing WebSocket!: ", err)
			} else {
				log.Println("Closed WebSocket connection given quit signal.")
			}
			return
		}
	}
}
