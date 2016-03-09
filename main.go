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
	"strconv"
	"strings"
	"syscall"
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
				messages +=
					"{\"topic\": \"" + consumerMessage.Topic +
						"\", \"partition\": \"" + strconv.FormatInt(int64(consumerMessage.Partition), 10) +
						"\", \"offset\": \"" + strconv.FormatInt(consumerMessage.Offset, 10) +
						"\", \"key\": \"" + strings.Replace(string(consumerMessage.Key), `"`, `\"`, -1) +
						"\", \"value\": \"" + strings.Replace(string(consumerMessage.Value), `"`, `\"`, -1) + "\"}\n"

				if len(messages) >= 10 {
					w.Header().Set("Access-Control-Allow-Origin", "*")
					io.WriteString(w, messages)
					return
				}
			case <-timer.C:
				w.Header().Set("Access-Control-Allow-Origin", "*")
				io.WriteString(w, messages)
				return
			}
		}
	}
}

func listenToSignals() {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-signals
		os.Exit(1)
	}()
}

func main() {
	config := readConfig()
	c := make(chan *sarama.ConsumerMessage)

	startConsumers(config, c)
	listenToSignals()

	var addr = flag.String("addr", ":41234", "host:port I'm serving on (default :41234)")
	flag.Parse()

	http.Handle("/", http.HandlerFunc(ServeWithChannel(c)))
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
