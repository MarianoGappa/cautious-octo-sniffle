package main

import (
	"encoding/json"
	"fmt"
	"github.com/Shopify/sarama"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
)

type ConsumerConfig struct {
	Broker    string `json:"broker"`
	Partition int    `json:"partition"`
	Topic     string `json:"topic"`
	Offset    string `json:"offset"`
}

type Config struct {
	Consumers []ConsumerConfig `json:"consumers"`
}

func readConfig() *Config {
	file, err := ioutil.ReadFile("./config.json")
	if err != nil {
		panic("Cannot open config.json")
	}
	config := new(Config)
	err = json.Unmarshal(file, &config)
	if err != nil {
		fmt.Println("error:", err)
	}

	return config
}

func main() {
	config := readConfig()

	topic, broker, partition := config.Consumers[0].Topic, config.Consumers[0].Broker, config.Consumers[0].Partition
	var offset int64 = -1
	switch config.Consumers[0].Offset {
	case "oldest":
		offset = -2
	case "newest":
		offset = -1
	default:
		panic("Invalid value for consumer offset")
	}

	c := make(chan *sarama.ConsumerMessage)

	go consume(c, topic, broker, partition, offset)

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	consumed := 0

MainLoop:
	for {
		select {
		case message := <-c:
			log.Printf("Consumed message offset %d\n", message.Offset)
			consumed++

		case <-signals:
			break MainLoop
		}
	}

	log.Printf("Consumed: %d\n", consumed)
}
