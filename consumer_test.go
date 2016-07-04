package main

import (
	"testing"

	"golang.org/x/net/websocket"

	"github.com/Shopify/sarama"
	"github.com/Shopify/sarama/mocks"
)

// func (m mockClient) GetOffset(topic string, partition int32, time int64) (int64, error) {
// 	if time == -1 {
// 		return m.newest, nil
// 	} else if time == -2 {
// 		return m.oldest, nil
// 	}

// 	return 0, fmt.Errorf("Test error :(")
// }

// func TestResolveOffset(t *testing.T) {
// 	offset, _ := resolveOffset("newest", []string{"localhost:9092"}, "test", 0, mockClientCreator{})
// 	if offset != -1 {
// 		t.Error("Offset should be -1 if newest is specified")
// 	}

// 	offset, _ = resolveOffset("oldest", []string{"localhost:9092"}, "test", 0, mockClientCreator{})
// 	if offset != -2 {
// 		t.Error("Offset should be -2 if oldest is specified")
// 	}

// 	offset, _ = resolveOffset("-2", []string{"localhost:9092"}, "test", 0, mockClientCreator{})
// 	if offset != -2 {
// 		t.Error("Offset should be -2 if -2 is specified")
// 	}

// 	offset, _ = resolveOffset("-10", []string{"localhost:9092"}, "test", 0, mockClientCreator{oldest: 100, newest: 200})
// 	if offset != 190 {
// 		t.Error("Offset should be 190 if -10 is specified and offset can be between [100, 200]")
// 	}

// 	offset, _ = resolveOffset("-10", []string{"localhost:9092"}, "test", 0, mockClientCreator{oldest: 100, newest: 105})
// 	if offset != 100 {
// 		t.Error("Offset should be 100 if -10 is specified and offset can be between [100, 105]")
// 	}
// }

func makeReadOnly(m chan *sarama.ConsumerMessage) <-chan *sarama.ConsumerMessage {
	return m
}

func TestDemuxMessages(t *testing.T) {
	m1 := make(chan *sarama.ConsumerMessage)
	m2 := make(chan *sarama.ConsumerMessage)
	m3 := make(chan *sarama.ConsumerMessage)

	mro1 := makeReadOnly(m1)
	mro2 := makeReadOnly(m2)
	mro3 := makeReadOnly(m3)
	m := []<-chan *sarama.ConsumerMessage{mro1, mro2, mro3}
	q := make(chan struct{})

	o := demuxMessages(m, q)

	m1 <- &sarama.ConsumerMessage{}
	<-o
	m2 <- &sarama.ConsumerMessage{}
	<-o
	m3 <- &sarama.ConsumerMessage{}
	<-o

	close(q)
}

type mockSender struct {
	err   error
	outCh chan string
}

func (m mockSender) Send(ws *websocket.Conn, msg string) error {
	m.outCh <- msg
	return m.err
}

type mockTimeNow struct {
	timestamp int64
}

func (m mockTimeNow) Unix() int64 {
	return m.timestamp
}

func TestSendMessagesToWsBlocking(t *testing.T) {
	c := make(chan *sarama.ConsumerMessage)
	q := make(chan struct{})
	o := make(chan string)

	go sendMessagesToWsBlocking(&websocket.Conn{}, c, q, mockSender{outCh: o}, mockTimeNow{timestamp: 123456})

	c <- &sarama.ConsumerMessage{Key: []byte("key"), Value: []byte("value"), Topic: "topic", Partition: 0, Offset: 123}
	s := <-o

	expected := `{"topic": "topic", "partition": "0", "offset": "123", "key": "key", "value": "value", "consumedUnixTimestamp": "123456"}` + "\n"

	if s != expected {
		t.Errorf("Result was %v rather than %v", s, expected)
	}

	close(q)
}

type mockUtils struct{}

func (k mockUtils) newClient(brokers []string) (sarama.Client, error) {
	return mockClient{}, nil
}

func (k mockUtils) newConsumerFromClient(client sarama.Client) (sarama.Consumer, error) {
	return &mocks.Consumer{}, nil
}

type mockClient struct {
	oldest     int64
	newest     int64
	closeError error
}

func (m mockClient) Config() *sarama.Config                                         { return nil }
func (m mockClient) Topics() ([]string, error)                                      { return nil, nil }
func (m mockClient) Partitions(topic string) ([]int32, error)                       { return nil, nil }
func (m mockClient) WritablePartitions(topic string) ([]int32, error)               { return nil, nil }
func (m mockClient) Leader(topic string, partitionID int32) (*sarama.Broker, error) { return nil, nil }
func (m mockClient) Replicas(topic string, partitionID int32) ([]int32, error)      { return nil, nil }
func (m mockClient) RefreshMetadata(topics ...string) error                         { return nil }
func (m mockClient) Coordinator(consumerGroup string) (*sarama.Broker, error)       { return nil, nil }
func (m mockClient) RefreshCoordinator(consumerGroup string) error                  { return nil }
func (m mockClient) Close() error                                                   { return nil }
func (m mockClient) Closed() bool                                                   { return false }
func (m mockClient) GetOffset(topic string, partitionID int32, time int64) (int64, error) {
	return 0, nil
}

func TestSetupClusters(t *testing.T) {
	clusters := make(map[string]*cluster)
	clusters["one"] = &cluster{brokers: []string{"one"}}
	clusters["two"] = &cluster{brokers: []string{"two"}}

	if clusters["one"].client != nil || clusters["one"].consumer != nil || clusters["two"].client != nil || clusters["two"].consumer != nil {
		t.Errorf("precondition for test failed")
	}

	errors := setupClusters(clusters, mockUtils{})

	if len(errors) > 0 {
		t.Errorf("setupClusters returned error(s)")
	}

	if clusters["one"].client == nil || clusters["one"].consumer == nil || clusters["two"].client == nil || clusters["two"].consumer == nil {
		t.Errorf("clients and consumers were not initialised properly")
	}
}
