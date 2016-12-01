package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

type bookie struct {
	url string
}

type partition struct {
	Start       int64 `json:"start"`
	End         int64 `json:"end"`
	LastScraped int64 `json:"lastScraped"`
	Count       int64 `json:"count"`
}

type topic struct {
	Partitions map[string]partition `json:"partitions"`
	Count      int64                `json:"count"`
}

type fsm struct {
	Topics  map[string]topic  `json:"topics"`
	Created string            `json:"created"`
	Id      string            `json:"id"`
	Tags    map[string]string `json:"tags"`
}

func (b bookie) fsm(fsmId string) (fsm, error) {
	f := fsm{}
	if len(b.url) == 0 {
		return f, fmt.Errorf("bookie is not enabled")
	}

	client := http.Client{Timeout: time.Duration(5 * time.Second)}
	r, err := client.Get(fmt.Sprintf("%v/fsm?id=%v", b.url, fsmId))
	if err != nil {
		return f, err
	}
	defer r.Body.Close()

	err = json.NewDecoder(r.Body).Decode(&f)
	return f, err
}

func (b bookie) latestFSMs(n int) ([]fsm, error) {
	if len(b.url) == 0 {
		return []fsm{}, fmt.Errorf("bookie is not enabled")
	}
	fsms := []fsm{}
	client := http.Client{Timeout: time.Duration(5 * time.Second)}
	r, err := client.Get(fmt.Sprintf("%v/latest?n=%v", b.url, n))
	if err != nil {
		return fsms, err
	}
	defer r.Body.Close()

	d := json.NewDecoder(r.Body)
	if err := d.Decode(&fsms); err != nil {
		fmt.Println(err)
		return fsms, err
	}

	return fsms, nil
}

func (f fsm) offset(topic string, partition int32) (int64, bool) {
	t, ok := f.Topics[topic]
	if !ok {
		return 0, false
	}

	p := strconv.Itoa(int(partition))

	if len(t.Partitions) == 0 {
		return 0, false
	}

	o, ok := t.Partitions[p]
	if !ok {
		return 0, false
	}

	return o.Start, true
}
