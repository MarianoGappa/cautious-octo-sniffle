package main

import (
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/Shopify/sarama"
	"golang.org/x/net/websocket"
)

func setupConsumers(conf *Config) ([]<-chan *sarama.ConsumerMessage, []io.Closer, error) {
	partitionConsumers := []<-chan *sarama.ConsumerMessage{}
	closeables := []io.Closer{}
	for _, consumerConfig := range conf.consumers {
		topic, brokers, partition := consumerConfig.topic, consumerConfig.brokers, consumerConfig.partition

		consumer, err := sarama.NewConsumer(brokers, nil)
		if err != nil {
			return nil, closeables, fmt.Errorf("Error creating consumer. err=%v", err)
		}

		var partitions []int32
		if partition == -1 {
			partitions, err = consumer.Partitions(topic)
			if err != nil {
				return nil, closeables, fmt.Errorf("Error fetching partitions for topic. err=%v", err)
			}
		} else {
			partitions = append(partitions, int32(partition))
		}

		for _, partition := range partitions {
			offset, err := resolveOffset(consumerConfig.offset, brokers, topic, partition, clientCreator{})
			if err != nil {
				return nil, closeables, fmt.Errorf("Could not resolve offset for %v, %v, %v. err=%v", brokers, topic, partition, err)
			}

			partitionConsumer, err := consumer.ConsumePartition(topic, int32(partition), offset)
			if err != nil {
				return nil, closeables, fmt.Errorf("Failed to consume partition %v err=%v\n", partition, err)
			}

			partitionConsumers = append(partitionConsumers, partitionConsumer.Messages())
			closeables = append(closeables, partitionConsumer)
		}
		closeables = append(closeables, consumer)
		log.Printf("Added consumer for topic [%v]", topic)
	}
	return partitionConsumers, closeables, nil
}

type iClient interface {
	GetOffset(string, int32, int64) (int64, error)
	Close() error
}

type client struct{}

func (c client) GetOffset(topic string, partition int32, time int64) (int64, error) {
	return c.GetOffset(topic, partition, time)
}

func (c client) Close() error {
	return c.Close()
}

type iClientCreator interface {
	NewClient([]string) (iClient, error)
}

type clientCreator struct{}

func (s clientCreator) NewClient(brokers []string) (iClient, error) {
	return sarama.NewClient(brokers, nil)
}

func resolveOffset(configOffset string, brokers []string, topic string, partition int32, clientCreator iClientCreator) (int64, error) {
	if configOffset == "oldest" {
		return sarama.OffsetOldest, nil
	} else if configOffset == "newest" {
		return sarama.OffsetNewest, nil
	} else if numericOffset, err := strconv.ParseInt(configOffset, 10, 64); err == nil {
		if numericOffset >= -2 {
			return numericOffset, nil
		}

		client, err := clientCreator.NewClient(brokers)
		if err != nil {
			return 0, fmt.Errorf("Failed to create client for %v, %v, %v", brokers, topic, partition)
		}
		defer client.Close()

		oldest, err := client.GetOffset(topic, partition, sarama.OffsetOldest)
		if err != nil {
			return 0, err
		}

		newest, err := client.GetOffset(topic, partition, sarama.OffsetNewest)
		if err != nil {
			return 0, err
		}

		if newest+numericOffset < oldest {
			return oldest, nil
		}

		return newest + numericOffset, nil
	}

	return 0, fmt.Errorf("Invalid value for consumer offset")
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

type iSender interface {
	Send(*websocket.Conn, string) error
}

type sender struct{}

func (s sender) Send(ws *websocket.Conn, msg string) error {
	return websocket.Message.Send(ws, msg)
}

type iTimeNow interface {
	Unix() int64
}

type timeNow struct{}

func (t timeNow) Unix() int64 {
	return time.Now().Unix()
}

func sendMessagesToWsBlocking(ws *websocket.Conn, c chan *sarama.ConsumerMessage, q chan struct{}, sender iSender, timeNow iTimeNow) {
	for {
		select {
		case cMsg := <-c:
			msg :=
				`{"topic": "` + cMsg.Topic +
					`", "partition": "` + strconv.FormatInt(int64(cMsg.Partition), 10) +
					`", "offset": "` + strconv.FormatInt(cMsg.Offset, 10) +
					`", "key": "` + strings.Replace(string(cMsg.Key), `"`, `\"`, -1) +
					`", "value": "` + strings.Replace(string(cMsg.Value), `"`, `\"`, -1) +
					`", "consumedUnixTimestamp": "` + strconv.FormatInt(timeNow.Unix(), 10) +
					`"}` + "\n"

			err := sender.Send(ws, msg)
			if err != nil {
				log.Printf("Error while trying to send to WebSocket: err=%v\n", err)
				return
			}
		case <-q:
			log.Println("Received quit signal")
			return
		}
	}
}
