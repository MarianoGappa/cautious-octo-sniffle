package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/Shopify/sarama"
	"golang.org/x/net/websocket"
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

type Link struct {
	Url   string
	Title string
}

func startConsumers(config *Config, c chan *sarama.ConsumerMessage, quit chan struct{}) {
	for _, consumerConfig := range config.Consumers {
		topic, broker, partition := consumerConfig.Topic, consumerConfig.Broker, consumerConfig.Partition
		var offset int64 = -1

		if numericOffset, err := strconv.ParseInt(consumerConfig.Offset, 10, 64); err == nil {
			offset = numericOffset
		} else {
			switch consumerConfig.Offset {
			case "oldest":
				offset = -2
			case "newest":
				offset = -1
			default:
				log.Println("Invalid value for consumer offset")
				close(quit)
				return
			}
		}

		go consume(c, quit, topic, broker, partition, offset)
	}
}

func onConnected(ws *websocket.Conn) {
	log.Println("Opened WebSocket connection!")

	var config Config
	err := websocket.JSON.Receive(ws, &config)
	if err != nil {
		ws.Close()
		log.Println("Didn't receive config from WebSocket!", err)
		return
	}

	c := make(chan *sarama.ConsumerMessage)

	startConsumers(&config, c, q)

	for {
		select {
		case consumerMessage := <-c:
			msg :=
				"{\"topic\": \"" + consumerMessage.Topic +
					"\", \"partition\": \"" + strconv.FormatInt(int64(consumerMessage.Partition), 10) +
					"\", \"offset\": \"" + strconv.FormatInt(consumerMessage.Offset, 10) +
					"\", \"key\": \"" + strings.Replace(string(consumerMessage.Key), `"`, `\"`, -1) +
					"\", \"value\": \"" + strings.Replace(string(consumerMessage.Value), `"`, `\"`, -1) +
					"\", \"consumedUnixTimestamp\": \"" + strconv.FormatInt(time.Now().Unix(), 10) +
					"\"}\n"

			log.Println("Sending message to WebSocket: " + msg)
			err := websocket.Message.Send(ws, msg)
			if err != nil {
				log.Println("Error while trying to send to WebSocket: ", err)
				close(q)
			}
		case <-q:
			err := ws.Close()
			log.Println("Closed WebSocket connection given quit signal.")
			if err != nil {
				log.Println("Error while closing WebSocket!: ", err)
			}
			return
		}
	}
}

func baseHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" && r.URL.RawQuery == "" {
		serveBaseHTML(w, r)
	} else {
		log.Println(r.URL.Path)
		http.FileServer(http.Dir("webroot")).ServeHTTP(w, r)
	}
}

func listenToSignals(q chan struct{}) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-signals
		close(q)
	}()
}

var mux = http.NewServeMux()

var q = make(chan struct{})

func main() {
	listenToSignals(q)

	mux.Handle("/ws", websocket.Handler(onConnected))
	mux.HandleFunc("/", baseHandler)

	port := 41234
	if len(os.Args) >= 2 {
		var err error
		port, err = strconv.Atoi(os.Args[1])
		if err != nil {
			log.Fatalf("Port receoved from flag could not be converted to int: %v", err)
		}
	}

	listener, err := net.ListenTCP("tcp", &net.TCPAddr{
		IP:   net.ParseIP("localhost").To4(),
		Port: port,
	})
	if err != nil {
		log.Fatalf("Listen: %v", err)
	}

	fmt.Printf("Flowbro is your bro on localhost:%v!\n", port)

	go func(listener *net.TCPListener, q chan struct{}) {
		<-q
		log.Println("Received quit signal; closing Snitch.")
		defer log.Println("closed Snitch")
		listener.Close()
	}(listener, q)

	http.Serve(listener, mux)
}
