package main

import (
	"fmt"
	"strings"
)

type consumerConfigJson struct {
	Brokers         string `json:"brokers,omitempty"`
	Partition       *int   `json:"partition,omitempty"`
	Topic           string `json:"topic"`
	Offset          string `json:"offset,omitempty"`
	BookieCountOnly bool   `json:"bookieCountOnly,omitempty"`
}

type kafka struct {
	Brokers   string               `json:"brokers,omitempty"`
	Consumers []consumerConfigJson `json:"consumers"`
	Grep      string               `json:"grep"`
	Offset    string               `json:"offset"`
}

type event struct {
	EventType  string                   `json:"eventType"`
	SourceId   string                   `json:"sourceId"`
	TargetId   string                   `json:"targetId"`
	Text       string                   `json:"text"`
	FSMId      string                   `json:"fsmId"`
	FSMIdAlias string                   `json:"fsmIdAlias"`
	JSON       []map[string]interface{} `json:"json"`
	Aggregate  bool                     `json:"aggregate"`
	Color      string                   `json:"color"`
	Count      int64                    `json:"count"`
	NoJSON     bool                     `json:"noJSON,omitempty"`
}

type pattern struct {
	Field   string
	Pattern string
}

type rule struct {
	Patterns []pattern
	Events   []event
}

type configJSON struct {
	Rules         []rule `json:"rules"`
	Kafka         kafka  `json:"kafka"`
	FSMId         string `json:"fsmId"`
	HeartbeatUUID string `json:"heartbeatUUID"`
}

type consumerConfig struct {
	brokers   []string
	partition int
	topic     string
	offset    string
}

type config struct {
	consumers       []consumerConfig
	brokers         []string
	fsmId           string
	bookieCountOnly []string
}

func processConfig(configJSON *configJSON) (*config, error) {
	config := &config{
		brokers:         strings.Split(configJSON.Kafka.Brokers, ","),
		fsmId:           configJSON.FSMId,
		bookieCountOnly: []string{},
	}

	globalOffset := configJSON.Kafka.Offset
	for _, consumerJSON := range configJSON.Kafka.Consumers {
		if consumerJSON.BookieCountOnly {
			config.bookieCountOnly = append(config.bookieCountOnly, consumerJSON.Topic)
			continue
		}

		consumer := consumerConfig{}

		if len(consumerJSON.Topic) == 0 {
			return config, fmt.Errorf("Please define topic name for your consumer %v", consumerJSON)
		}
		consumer.topic = consumerJSON.Topic
		consumer.brokers = config.brokers

		if len(consumerJSON.Offset) == 0 {
			if len(globalOffset) > 0 {
				consumer.offset = globalOffset
			} else {
				consumer.offset = "newest"
			}
		} else {
			consumer.offset = consumerJSON.Offset
		}

		if consumerJSON.Partition != nil {
			consumer.partition = *consumerJSON.Partition
		} else {
			consumer.partition = -1
		}
		config.consumers = append(config.consumers, consumer)
	}

	return config, nil
}
