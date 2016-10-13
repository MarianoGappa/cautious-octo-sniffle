package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"encoding/json"

	"github.com/Shopify/sarama"
	"golang.org/x/net/websocket"
)

func gowg(f func(), wg *sync.WaitGroup) {
	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		f()
	}(wg)
}

type message struct {
	Key       string                 `json:"key"`
	Value     map[string]interface{} `json:"value"`
	Topic     string                 `json:"topic"`
	Partition int32                  `json:"partition"`
	Offset    int64                  `json:"offset"`
	Timestamp time.Time              `json:"timestamp"` // only set if kafka is version 0.10+
}

type cluster struct {
	brokers            []string
	consumer           sarama.Consumer
	client             sarama.Client
	partitionConsumers []sarama.PartitionConsumer
	pcL                sync.Mutex
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

func closeAll(clusters map[string]*cluster) {
	log.Printf("Trying to close all clusters")
	for _, c := range clusters {
		c.close()
	}
	log.Printf("Finished trying to close all clusters")
}

type iKafkaUtils interface {
	newClient(brokers []string) (sarama.Client, error)
	newConsumerFromClient(client sarama.Client) (sarama.Consumer, error)
}

type kafkaUtils struct{}

func (k kafkaUtils) newClient(brokers []string) (sarama.Client, error) {
	return sarama.NewClient(brokers, nil)
}

func (k kafkaUtils) newConsumerFromClient(client sarama.Client) (sarama.Consumer, error) {
	return sarama.NewConsumerFromClient(client)
}

func setupClusters(clusters map[string]*cluster, utils iKafkaUtils) []error {
	errors := []error{}
	var errL sync.Mutex

	var wg sync.WaitGroup

	for b := range clusters {
		gowg(func(b string) func() {
			return func() {
				c := clusters[b]
				log.Printf("Adding client+consumer for cluster with brokers %v", c.brokers)
				client, err := utils.newClient(c.brokers)
				if err != nil {
					errL.Lock()
					errors = append(errors, fmt.Errorf("Error creating client. err=%v", err))
					errL.Unlock()
				}
				consumer, err := utils.newConsumerFromClient(client)
				if err != nil {
					errL.Lock()
					errors = append(errors, fmt.Errorf("Error creating consumer. err=%v", err))
					errL.Unlock()
				}
				c.client = client
				c.consumer = consumer
			}
		}(b), &wg)
	}
	wg.Wait()

	return errors
}

func setupPartitionConsumers(conf *config) ([]<-chan *sarama.ConsumerMessage, map[string]*cluster, bool) {
	clusters := make(map[string]*cluster)
	for _, c := range conf.consumers {
		b := strings.Join(c.brokers, ",")
		if _, exists := clusters[b]; !exists {
			clusters[b] = &cluster{brokers: c.brokers}
		}
	}

	errors := setupClusters(clusters, kafkaUtils{})
	var errL sync.Mutex
	if len(errors) > 0 {
		log.Printf("%v error(s) while setting up consumers:", len(errors))
		for i, e := range errors {
			log.Printf("Error #%v: %v", i, e)
		}
		closeAll(clusters)
		return nil, nil, false
	}

	partitionConsumerChans := []<-chan *sarama.ConsumerMessage{}
	var pccL sync.Mutex

	var wg sync.WaitGroup

	for _, consumerConf := range conf.consumers {
		gowg(func(consumerConf consumerConfig) func() {
			return func() {
				topic, brokers, partition := consumerConf.topic, consumerConf.brokers, consumerConf.partition
				brokStr := strings.Join(brokers, ",")
				cluster := clusters[brokStr]
				client, consumer, pcL := cluster.client, cluster.consumer, cluster.pcL

				var partitions []int32
				if partition == -1 {
					var err error
					partitions, err = consumer.Partitions(topic)
					if err != nil {
						errL.Lock()
						errors = append(errors, fmt.Errorf("Error fetching partitions for topic %v. err=%v", topic, err))
						errL.Unlock()
						return
					}
				} else {
					partitions = append(partitions, int32(partition))
				}

				for _, partition := range partitions {
					offset, err := resolveOffset(consumerConf.offset, brokers, topic, partition, client)
					if err != nil {
						errL.Lock()
						errors = append(errors, fmt.Errorf("Could not resolve offset for %v, %v, %v. err=%v", brokers, topic, partition, err))
						errL.Unlock()
						return
					}

					partitionConsumer, err := consumer.ConsumePartition(topic, int32(partition), offset)
					if err != nil {
						errL.Lock()
						errors = append(errors, fmt.Errorf("Failed to consume partition %v err=%v\n", partition, err))
						errL.Unlock()
						return
					}

					pcL.Lock()
					cluster.partitionConsumers = append(cluster.partitionConsumers, partitionConsumer)
					pcL.Unlock()
					pccL.Lock()
					partitionConsumerChans = append(partitionConsumerChans, partitionConsumer.Messages())
					pccL.Unlock()
				}
				log.Printf("Added %v partition consumer(s) for topic [%v]", len(partitions), topic)
			}
		}(consumerConf), &wg)
	}

	wg.Wait()

	if len(errors) > 0 {
		log.Printf("%v error(s) while setting up partition consumers:", len(errors))
		for i, e := range errors {
			log.Printf("Error #%v: %v", i, e)
		}
		closeAll(clusters)
		return nil, nil, false
	}

	log.Println("Successfully finished setting up partition consumers. Ready to consume, bro!")
	return partitionConsumerChans, clusters, true
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
	saramaConfig := sarama.NewConfig()
	saramaConfig.Version = sarama.V0_10_0_0
	return sarama.NewClient(brokers, saramaConfig)
}

func resolveOffset(configOffset string, brokers []string, topic string, partition int32, client sarama.Client) (int64, error) {
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

func sendMessagesToWsBlocking(ws *websocket.Conn, c chan *sarama.ConsumerMessage, q chan struct{}, sender iSender, timeNow iTimeNow, rules []rule, globalKey string) {
	var tick <-chan time.Time
	var ticker *time.Ticker

	currentTimestamp := int64(0)
	buffer := []message{}
	keyAliases := map[string]string{}
	initialTimerCh := time.After(5 * time.Second)

	for {
		select {
		case cMsg := <-c:
			m, err := newMessage(*cMsg)
			if err != nil {
				log.Printf("Could not parse %v into message", err)
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
		case <-initialTimerCh:
			initialTimerCh = nil
			if len(buffer) == 0 {
				currentTimestamp = time.Now().UnixNano() / 1000000
			} else {
				currentTimestamp = buffer[0].Timestamp.UnixNano() / 1000000
			}
			log.Printf("Starting at timestamp %v", currentTimestamp)
			ticker = time.NewTicker(time.Millisecond * 100)
			tick = ticker.C
		case <-tick:
			events := []event{}
			for i := 0; len(buffer) > 0 && i < 1000; i++ {
				if buffer[0].Timestamp.UnixNano()/1000000 <= currentTimestamp {
					evs, err := processMessage(buffer[0], rules, keyAliases, globalKey)
					if err != nil {
						log.Printf("Error while processing message: err=%v", err)
						break
					}
					events = append(events, evs...)
					buffer = buffer[1:]
				} else {
					break
				}
			}

			if len(events) == 0 {
				break
			}

			byt, err := json.Marshal(events)
			if err != nil {
				log.Printf("Error while marshalling events: err=%v\n", err)
				continue
			}

			err = sender.Send(ws, string(byt))
			if err != nil {
				log.Printf("Error while trying to send to WebSocket: err=%v\n", err)
				return
			}

			currentTimestamp += 100
		case <-q:
			log.Println("Received quit signal")
			return
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
