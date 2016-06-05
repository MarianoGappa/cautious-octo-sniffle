package main

import (
	"fmt"
	"html/template"
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

type consumerConfig struct {
	Broker    string `json:"broker"`
	Partition int    `json:"partition"`
	Topic     string `json:"topic"`
	Offset    string `json:"offset"`
}

type Config struct {
	Consumers []consumerConfig `json:"consumers"`
}

type link struct {
	URL   string
	Title string
}

func onConnected(mainQuit chan struct{}) func(ws *websocket.Conn) {
	return func(ws *websocket.Conn) {
		log.Println("Opened WebSocket connection!")

		var config Config
		err := websocket.JSON.Receive(ws, &config)
		if err != nil {
			ws.Close()
			log.Println("Didn't receive config from WebSocket!", err)
			return
		}

		c := make(chan *sarama.ConsumerMessage)
		reqQuit := make(chan struct{})

		startConsumers(&config, c, mainQuit, reqQuit)

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
					err := ws.Close()
					if err != nil {
						log.Println("Error while closing WebSocket!: ", err)
					} else {
						log.Println("Closed WebSocket connection given quit signal.")
					}
					return
				}
			case <-reqQuit:
				err := ws.Close()
				if err != nil {
					log.Println("Error while closing WebSocket!: ", err)
				} else {
					log.Println("Closed WebSocket connection given quit signal.")
				}
				return
			case <-mainQuit:
				close(reqQuit)
			}
		}
	}
}

func baseHandler(template *template.Template) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" && r.URL.RawQuery == "" {
			if err := serveBaseHTML(template, w, r); err != nil {
				log.Printf("Loading base page failed; ignoring. err=%v\n", err)
			}
		} else {
			http.FileServer(http.Dir("webroot")).ServeHTTP(w, r)
		}
	}
}

func mustListenToSignals(q chan struct{}) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-signals
		close(q)
	}()
}

func resolvePort(def int) (int, error) {
	if len(os.Args) >= 2 {
		return strconv.Atoi(os.Args[1])
	}
	return def, nil
}

func newListener(port int, q chan struct{}) (*net.TCPListener, error) {
	listener, err := net.ListenTCP("tcp", &net.TCPAddr{
		IP:   net.ParseIP("localhost").To4(),
		Port: port,
	})
	if err != nil {
		return listener, err
	}

	go func(listener *net.TCPListener, q chan struct{}) {
		<-q
		log.Println("Received quit signal; closing tcp listener.")
		listener.Close()
		log.Println("closed tcp listener")
	}(listener, q)

	return listener, nil
}

func main() {
	port, err := resolvePort(41234)
	if err != nil {
		log.Fatalf("Port received from flag could not be converted to int: %v", err)
	}

	q := make(chan struct{})

	listener, err := newListener(port, q)
	if err != nil {
		log.Fatalf("Could not open tcp listener: %v", err)
	}

	baseTemplate, err := parseBasePageTemplate()
	if err != nil {
		log.Fatalf("Could: %v", err)
	}

	mustListenToSignals(q)

	fmt.Printf("Flowbro is your bro on localhost:%v!\n", port)

	mux := http.NewServeMux()
	mux.Handle("/ws", websocket.Handler(onConnected(q)))
	mux.HandleFunc("/", baseHandler(baseTemplate))

	http.Serve(listener, mux)
}
