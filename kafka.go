package main

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/Shopify/sarama"
	log "github.com/Sirupsen/logrus"
)

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

func (c *cluster) addConsumer(conf consumerConfig, fsm fsm) {
	topic, brokers, partition := conf.topic, conf.brokers, conf.partition
	client, consumer := c.client, c.consumer

	partitions, err := resolvePartitions(topic, partition, consumer)
	if err != nil {
		c.es.add(err.Error())
		return
	}

	for _, partition := range partitions {
		offset, err := resolveOffset(fsm, conf.offset, topic, partition, client)
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

func setupCluster(conf *config, bookie bookie) cluster {
	f, err := bookie.fsm(conf.fsmId)
	if err != nil {
		log.WithFields(log.Fields{"err": err, "fsmId": conf.fsmId, "url": bookie.url}).Error("Failed to fetch FSMId from Bookie.")
	}

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
		go func(consumerConf consumerConfig, f fsm, c *cluster, wg *sync.WaitGroup) {
			defer wg.Done()
			c.addConsumer(consumerConf, f)
		}(consumerConf, f, &c, &wg)
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

func resolveOffset(fsm fsm, configOffset string, topic string, partition int32, client sarama.Client) (int64, error) {
	oldest, err := client.GetOffset(topic, partition, sarama.OffsetOldest)
	if err != nil {
		return 0, err
	}

	bookieOffset, ok := fsm.offset(topic, partition)
	if ok {
		if bookieOffset >= oldest {
			return bookieOffset, nil
		}

		log.WithFields(log.Fields{"oldest": oldest, "bookieOffset": bookieOffset}).Error("Bookie's offset is older than Kafka's oldest.")
		return oldest, nil
	}

	if configOffset == "oldest" {
		return sarama.OffsetOldest, nil
	}

	if configOffset == "newest" {
		return sarama.OffsetNewest, nil
	}

	numericOffset, err := strconv.ParseInt(configOffset, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("Invalid value for consumer offset")
	}

	if numericOffset >= -2 {
		return numericOffset, nil
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
