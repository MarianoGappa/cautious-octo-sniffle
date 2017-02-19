package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net"
	"net/http"

	"github.com/Shopify/sarama"
	log "github.com/Sirupsen/logrus"
	"golang.org/x/net/websocket"
)

type flowbro struct{}

func (f *flowbro) onConnected() func(ws *websocket.Conn) {
	return func(ws *websocket.Conn) {
		log.Println("Opened WebSocket connection!")

		var configJSON configJSON
		err := websocket.JSON.Receive(ws, &configJSON)
		if err != nil {
			ws.Close()
			log.Println("Didn't receive config from WebSocket!", err)
			return
		}

		config, err := processConfig(&configJSON)
		if err != nil {
			sendError(fmt.Sprintf("Closing WebSocket connection due to: %v\n", err), ws)
			ws.Close()
			return
		}

		c, bookieCounts, cluster, ok := setupKafka(ws, config)
		if !ok {
			return
		}

		process(ws, c, sender{}, configJSON.Rules, configJSON.FSMId, configJSON.HeartbeatUUID, bookieCounts)

		if !config.tutorial {
			cluster.close()
		}
		ws.Close()
	}
}

func setupKafka(ws *websocket.Conn, config *config) (chan *sarama.ConsumerMessage, map[string]int64, *cluster, bool) {
	bookieCounts := map[string]int64{}
	bookie, f := bookie{}, fsm{}
	var err error
	if config.bookieUrl != "" {
		bookie = newBookie(config.bookieUrl)
		f, err = bookie.fsm(config.fsmId)
		if len(config.fsmId) > 0 && err != nil {
			log.WithFields(log.Fields{"err": err, "fsmId": config.fsmId, "url": bookie.url}).Warn("Failed to fetch FSMId from Bookie.")
		}
	}

	if config.tutorial {
		sendSuccess("Starting tutorial. Flowbro is not really connected to a Kafka broker; messages are being mocked.", ws)
		return tutorial(), bookieCounts, &cluster{}, true
	}

	cluster := setupCluster(config, f)
	if len(cluster.es.errors) > 0 {
		sendError(fmt.Sprintf("Closing WebSocket connection due to errors while setting up partition consumers: %v", cluster.es.errors), ws)
		cluster.close()
		ws.Close()
		return nil, bookieCounts, nil, false
	}

	c := joinMessages(cluster.chs)

	for _, t := range config.bookieCountOnly {
		if len(config.fsmId) == 0 {
			sendError(fmt.Sprintf("Note that, since fsmId is not set, you won't see any events coming from topic %v", t), ws)
			continue
		}

		if ti, ok := f.Topics[t]; ok {
			bookieCounts[t] = ti.Count
			continue
		}
		sendError(fmt.Sprintf("Didn't find message count for topic %v for fsmID %v on Bookie", t, f.Id), ws)
	}

	return c, bookieCounts, &cluster, true
}

func sendError(error string, ws *websocket.Conn) {
	log.Printf(error)

	byt, err := json.Marshal([]event{{EventType: "log", Text: error, Color: "error"}})
	if err != nil {
		log.Printf("Error while marshalling error: err=%v\n", err)
		return
	}

	websocket.Message.Send(ws, string(byt))
}

func sendSuccess(text string, ws *websocket.Conn) {
	log.Printf(text)

	byt, err := json.Marshal([]event{{EventType: "log", Text: text, Color: "happy"}})
	if err != nil {
		log.Printf("Error while marshalling error: err=%v\n", err)
		return
	}

	websocket.Message.Send(ws, string(byt))
}

func (f *flowbro) baseHandler(template *template.Template) func(http.ResponseWriter, *http.Request) {
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

func serve(f *flowbro, baseTemplate *template.Template, listener *net.TCPListener) {
	mux := http.NewServeMux()
	mux.Handle("/ws", websocket.Handler(f.onConnected()))
	mux.HandleFunc("/", f.baseHandler(baseTemplate))

	err := http.Serve(listener, mux)
	if err != nil {
		log.Println("Flowbro server went down: ", err)
	}
}
