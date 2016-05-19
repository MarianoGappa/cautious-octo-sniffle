package main

import (
	"fmt"
	"github.com/Shopify/sarama"
	"golang.org/x/net/websocket"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
	"path/filepath"
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

func startConsumers(config *Config, c chan *sarama.ConsumerMessage, quit chan bool) {
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
				quit <- true
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

	quit := make(chan bool)
	c := make(chan *sarama.ConsumerMessage)

	startConsumers(&config, c, quit)

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
				quit <- true
			}
		case <-quit:
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
		chttp.ServeHTTP(w, r)
	}
}

func serveBaseHTML(w http.ResponseWriter, r *http.Request) {
	baseHTML := `
		<!DOCTYPE html>
		<html>
		<head>
		    <title>Flowbro</title>
		</head>
		<body>
		    <ul>
		        {{range $i, $e := .}}<li>
		            <a href="{{.Url}}">{{.Title}}</a>
		        </li>{{end}}
		    </ul>
		    <style>

		    </style>
		</body>
		</html>
	`

	files, err := ioutil.ReadDir("webroot/configs")
	if err != nil {
		log.Fatal(err)
	}

	links := []Link{}
	for _, file := range files {
		config := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))
		links = append(links, Link{
			Url:   "?config=" + config,
			Title: config,
		})
	}

	templ, err := template.New("base").Parse(baseHTML)
	if err != nil {
		log.Fatal(err)
	}

	err = templ.Execute(w, links)
	if err != nil {
		log.Fatal(err)
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

var chttp = http.NewServeMux()

func main() {
	listenToSignals()

	http.Handle("/ws", websocket.Handler(onConnected))
	chttp.Handle("/", http.FileServer(http.Dir("webroot")))
	http.HandleFunc("/", baseHandler)

	if len(os.Args) == 1 {
		fmt.Println("usage: flowbro {portToServeOn}\n")
		os.Exit(2)
	}

	port := os.Args[1]
	err := http.ListenAndServe(":"+string(port), nil)
	if err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
