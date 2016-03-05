package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/Shopify/sarama"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
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

func startConsumers(config *Config, c chan *sarama.ConsumerMessage) {
	for _, consumerConfig := range config.Consumers {
		topic, broker, partition := consumerConfig.Topic, consumerConfig.Broker, consumerConfig.Partition
		var offset int64 = -1
		switch consumerConfig.Offset {
		case "oldest":
			offset = -2
		case "newest":
			offset = -1
		default:
			panic("Invalid value for consumer offset")
		}

		go consume(c, topic, broker, partition, offset)
	}
}

func ServeWithChannel(c chan *sarama.ConsumerMessage) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		messages := ""
		timer := time.NewTimer(time.Millisecond * 10)
		for {
			select {
			case consumerMessage := <-c:
				// fmt.Printf("Received message with value: [%v]\n", string(consumerMessage.Value))
				messages += string(consumerMessage.Value) + "\n"
			case <-timer.C:
				io.WriteString(w, messages)
				return
			}
		}
	}
}

func Die(w http.ResponseWriter, req *http.Request) {
	os.Exit(1)
}

func main() {
	config := readConfig()
	c := make(chan *sarama.ConsumerMessage)

	startConsumers(config, c)

	var addr = flag.String("addr", ":41234", "host:port I'm serving on (default :41234)")

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	flag.Parse()
	http.Handle("/", http.HandlerFunc(ServeWithChannel(c)))
	http.Handle("/die", http.HandlerFunc(Die))
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
