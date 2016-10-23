package main

import (
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"encoding/json"

	"github.com/Shopify/sarama"
	"golang.org/x/net/websocket"
)

type message struct {
	Key       string                 `json:"key"`
	Value     map[string]interface{} `json:"value"`
	Topic     string                 `json:"topic"`
	Partition int32                  `json:"partition"`
	Offset    int64                  `json:"offset"`
	Timestamp time.Time              `json:"timestamp"` // only set if kafka is version 0.10+
}

type cluster struct {
	brokers  []string
	consumer sarama.Consumer
	client   sarama.Client

	partitionConsumers []sarama.PartitionConsumer
	pcLock             sync.Mutex

	chs     []<-chan *sarama.ConsumerMessage
	chsLock sync.Mutex

	es errorlist
}

func (c *cluster) addPartitionConsumer(pc sarama.PartitionConsumer) {
	c.pcLock.Lock()
	c.partitionConsumers = append(c.partitionConsumers, pc)
	c.pcLock.Unlock()
}

func (c *cluster) addCh(ch <-chan *sarama.ConsumerMessage) {
	c.chsLock.Lock()
	c.chs = append(c.chs, ch)
	c.chsLock.Unlock()
}

func (c *cluster) close() {
	log.Printf("Trying to close cluster with brokers %v", c.brokers)

	log.Printf("Trying to close %v partition consumers for cluster with brokers %v", len(c.partitionConsumers), c.brokers)
	for _, pc := range c.partitionConsumers {
		if err := pc.Close(); err != nil {
			log.Printf("Error while trying to close partition consumer for cluster with brokers %v. err=%v", c.brokers, err)
		}
	}

	log.Printf("Trying to close consumer for cluster with brokers %v", c.brokers)
	if err := c.consumer.Close(); err != nil {
		log.Printf("Error while trying to close consumer for cluster with brokers %v. err=%v", c.brokers, err)
	} else {
		log.Printf("Successfully closed consumer for cluster with brokers %v", c.brokers)
	}

	log.Printf("Trying to close client for cluster with brokers %v", c.brokers)
	if err := c.client.Close(); err != nil {
		log.Printf("Error while trying to close client for cluster with brokers %v. err=%v", c.brokers, err)
	} else {
		log.Printf("Successfully closed client for cluster with brokers %v", c.brokers)
	}

	log.Printf("Finished trying to close cluster with brokers %v", c.brokers)
}

func (c *cluster) addConsumer(conf consumerConfig) {
	topic, brokers, partition := conf.topic, conf.brokers, conf.partition
	client, consumer := c.client, c.consumer

	partitions, err := resolvePartitions(topic, partition, consumer)
	if err != nil {
		c.es.add(err.Error())
		return
	}

	for _, partition := range partitions {
		offset, err := resolveOffset(conf.offset, topic, partition, client)
		if err != nil {
			c.es.add(fmt.Sprintf("Could not resolve offset for %v, %v, %v. err=%v", brokers, topic, partition, err))
			return
		}

		partitionConsumer, err := consumer.ConsumePartition(topic, int32(partition), offset)
		if err != nil {
			c.es.add(fmt.Sprintf("Failed to consume partition %v err=%v\n", partition, err))
			return
		}

		c.addPartitionConsumer(partitionConsumer)
		c.addCh(partitionConsumer.Messages())
		log.Printf("Consuming topic [%v], partition [%v] from offset [%v]", topic, partition, offset)
	}
}

func setupCluster(conf *config) cluster {
	c := cluster{brokers: conf.brokers}

	saramaConfig := sarama.NewConfig()
	saramaConfig.Version = sarama.V0_10_0_0
	client, err := sarama.NewClient(c.brokers, saramaConfig)
	if err != nil {
		c.es.add(fmt.Sprintf("Error creating client. err=%v", err))
		return c
	}

	consumer, err := sarama.NewConsumerFromClient(client)
	if err != nil {
		c.es.add(fmt.Sprintf("Error creating consumer. err=%v", err))
		return c
	}

	c.client = client
	c.consumer = consumer

	var wg sync.WaitGroup
	for _, consumerConf := range conf.consumers {
		wg.Add(1)
		go func(consumerConf consumerConfig, c *cluster, wg *sync.WaitGroup) {
			defer wg.Done()
			c.addConsumer(consumerConf)
		}(consumerConf, &c, &wg)
	}
	wg.Wait()

	return c
}

func resolvePartitions(topic string, partition int, consumer sarama.Consumer) ([]int32, error) {
	var partitions []int32
	if partition == -1 {
		var err error

		partitions, err = consumer.Partitions(topic)
		if err != nil {
			return partitions, fmt.Errorf("Error fetching partitions for topic %v. err=%v", topic, err)
		}
	} else {
		partitions = append(partitions, int32(partition))
	}
	return partitions, nil
}

func resolveOffset(configOffset string, topic string, partition int32, client sarama.Client) (int64, error) {
	if configOffset == "oldest" {
		return sarama.OffsetOldest, nil
	} else if configOffset == "newest" {
		return sarama.OffsetNewest, nil
	} else if numericOffset, err := strconv.ParseInt(configOffset, 10, 64); err == nil {
		if numericOffset >= -2 {
			return numericOffset, nil
		}

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

func joinMessages(pc []<-chan *sarama.ConsumerMessage) chan *sarama.ConsumerMessage {
	c := make(chan *sarama.ConsumerMessage)
	for _, p := range pc {
		go func(p <-chan *sarama.ConsumerMessage) {
			for {
				select {
				case msg := <-p:
					c <- msg
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

func sendMessagesToWsBlocking(ws *websocket.Conn, c chan *sarama.ConsumerMessage, sender iSender, rules []rule, globalFSMId string) {
	ticker := time.NewTicker(time.Millisecond * 100)

	buffer := []message{}
	fsmIdAliases := map[string]string{}
	sendSuccess("Starting to send messages!", ws)

	for {
		select {
		case cMsg := <-c:
			m, err := newMessage(*cMsg)
			if err != nil {
				sendError(fmt.Sprintf("Could not parse %v into message", err), ws)
			}
			if m.Timestamp.UnixNano() <= 0 {
				m.Timestamp = time.Now()
			}
			for i := range buffer {
				target := len(buffer) - 1 - i
				if m.Timestamp.UnixNano() >= buffer[target].Timestamp.UnixNano() {
					buffer = sliceInsert(buffer, target+1, m)
					break
				}
			}
			if len(buffer) == 0 {
				buffer = append(buffer, m)
			}
		case <-ticker.C:
			events := []event{}
			incompleteEvents := []event{}
			for i := 0; len(buffer) > 0 && i < 1000; i++ {
				err := processMessage(buffer[0], rules, fsmIdAliases, &events, &incompleteEvents, globalFSMId)
				if err != nil {
					sendError(fmt.Sprintf("Error while processing message: err=%v", err), ws)
					break
				}
				buffer = buffer[1:]
			}

			for _, ie := range incompleteEvents {
				events = aggregate(events, ie, ie.Aggregate, globalFSMId)
			}

			if len(events) == 0 {
				break
			}

			byt, err := json.Marshal(events)
			if err != nil {
				sendError(fmt.Sprintf("Error while marshalling events: err=%v\n", err), ws)
				continue
			}

			err = sender.Send(ws, string(byt))
			if err != nil {
				log.Printf("Error while trying to send to WebSocket: err=%v\n", err)
				return
			}
		}
	}
}

func newMessage(cm sarama.ConsumerMessage) (message, error) {
	var v interface{}
	if err := json.Unmarshal(cm.Value, &v); err != nil {
		return message{}, err
	}

	return message{
		Key:       string(cm.Key),
		Value:     v.(map[string]interface{}),
		Topic:     cm.Topic,
		Partition: cm.Partition,
		Offset:    cm.Offset,
		Timestamp: cm.Timestamp,
	}, nil
}

func sliceInsert(slice []message, index int, value message) []message {
	if index == 0 {
		return append([]message{value}, slice...)
	}
	if index >= len(slice) {
		return append(slice, value)
	}
	// Grow the slice by one element.
	slice = slice[0 : len(slice)+1]
	// Use copy to move the upper part of the slice out of the way and open a hole.
	copy(slice[index+1:], slice[index:])
	// Store the new value.
	slice[index] = value
	// Return the result.
	return slice
}
