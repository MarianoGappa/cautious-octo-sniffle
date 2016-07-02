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

type ConfigJSON struct {
	Brokers   string               `json:"brokers,omitempty"`
	Consumers []consumerConfigJson `json:"consumers"`
}

type consumerConfig struct {
	brokers   []string
	partition int
	topic     string
	offset    string
}

type Config struct {
	consumers []consumerConfig `json:"consumers"`
}

func processConfig(configJSON *ConfigJSON) (*Config, error) {
	config := &Config{}

	globalBrokers := configJSON.Brokers
	for _, consumerJSON := range configJSON.Consumers {
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
			consumer.offset = "newest"
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
