package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

type consumerConfigJson struct {
	Brokers   string `json:"brokers,omitempty"`
	Partition *int   `json:"partition,omitempty"`
	Topic     string `json:"topic"`
	Offset    string `json:"offset,omitempty"`
}

type kafka struct {
	Brokers   string               `json:"brokers,omitempty"`
	Consumers []consumerConfigJson `json:"consumers"`
	Grep      string               `json:"grep"`
	Offset    string               `json:"offset"`
}

type event struct {
	EventType string                 `json:"eventType"`
	SourceId  string                 `json:"sourceId"`
	TargetId  string                 `json:"targetId"`
	Text      string                 `json:"text"`
	Key       string                 `json:"key"`
	KeyAlias  string                 `json:"keyAlias"`
	JSON      map[string]interface{} `json:"json"`
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
	Rules []rule
	Kafka kafka
	Key   string
}

type consumerConfig struct {
	brokers   []string
	partition int
	topic     string
	offset    string
}

type config struct {
	consumers []consumerConfig `json:"consumers"`
}

func processConfig(configJSON *configJSON) (*config, error) {
	config := &config{}

	globalBrokers := configJSON.Kafka.Brokers
	globalOffset := configJSON.Kafka.Offset
	for _, consumerJSON := range configJSON.Kafka.Consumers {
		consumer := consumerConfig{}

		if len(consumerJSON.Topic) == 0 {
			return config, fmt.Errorf("Please define topic name for your consumer %v", consumerJSON)
		}
		consumer.topic = consumerJSON.Topic

		if len(consumerJSON.Brokers) == 0 {
			if len(globalBrokers) == 0 {
				return config, fmt.Errorf("No broker information available in %v", consumerJSON)
			}
			consumer.brokers = strings.Split(globalBrokers, ",")
		} else {
			consumer.brokers = strings.Split(consumerJSON.Brokers, ",")
		}

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

func mustResolvePort(num int) int {
	port, err := resolvePort(num)
	if err != nil {
		log.Fatalf("Could not resolve port %v. err=%v", num, err)
	}
	return port
}

func resolvePort(num int) (int, error) {
	if len(os.Args) >= 2 {
		return strconv.Atoi(os.Args[1])
	}
	return num, nil
}
