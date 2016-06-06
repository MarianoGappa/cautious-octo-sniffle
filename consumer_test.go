package main

import (
	"fmt"
	"testing"

	"github.com/Shopify/sarama"
)

type mockClientCreator struct {
	oldest     int64
	newest     int64
	closeError error
}

type mockClient struct {
	oldest     int64
	newest     int64
	closeError error
}

func (m mockClientCreator) NewClient(brokers []string) (iClient, error) {
	return mockClient{oldest: m.oldest, newest: m.newest, closeError: m.closeError}, nil
}

func (m mockClient) GetOffset(topic string, partition int32, time int64) (int64, error) {
	if time == -1 {
		return m.newest, nil
	} else if time == -2 {
		return m.oldest, nil
	}

	return 0, fmt.Errorf("Test error :(")
}

func (m mockClient) Close() error {
	return m.closeError
}

func TestResolveOffset(t *testing.T) {
	offset, _ := resolveOffset("newest", []string{"localhost:9092"}, "test", 0, mockClientCreator{})
	if offset != -1 {
		t.Error("Offset should be -1 if newest is specified")
	}

	offset, _ = resolveOffset("oldest", []string{"localhost:9092"}, "test", 0, mockClientCreator{})
	if offset != -2 {
		t.Error("Offset should be -2 if oldest is specified")
	}

	offset, _ = resolveOffset("-2", []string{"localhost:9092"}, "test", 0, mockClientCreator{})
	if offset != -2 {
		t.Error("Offset should be -2 if -2 is specified")
	}

	offset, _ = resolveOffset("-10", []string{"localhost:9092"}, "test", 0, mockClientCreator{oldest: 100, newest: 200})
	if offset != 190 {
		t.Error("Offset should be 190 if -10 is specified and offset can be between [100, 200]")
	}

	offset, _ = resolveOffset("-10", []string{"localhost:9092"}, "test", 0, mockClientCreator{oldest: 100, newest: 105})
	if offset != 100 {
		t.Error("Offset should be 100 if -10 is specified and offset can be between [100, 105]")
	}
}

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
