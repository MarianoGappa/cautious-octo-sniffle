package main

import (
	"testing"
	"time"
)

func TestProcessHeartbeatTimesOut(t *testing.T) {
	timeout := make(chan struct{})

	go processHeartbeats(blockingWR{}, timeout, "uuid", 10*time.Millisecond)

	select {
	case <-timeout:
	case <-time.After(20 * time.Millisecond):
		t.Error("Didn't timeout")
	}
}

func TestProcessHeartbeatTimesOutGivenWrongUUID(t *testing.T) {
	timeout := make(chan struct{})

	go processHeartbeats(invalidWR{}, timeout, "uuid", 10*time.Millisecond)

	select {
	case <-timeout:
	case <-time.After(20 * time.Millisecond):
		t.Error("Didn't timeout")
	}
}

func TestProcessHeartbeatDoesntTimeout(t *testing.T) {
	timeout := make(chan struct{})

	go processHeartbeats(validWR{}, timeout, "uuid", 10*time.Millisecond)

	select {
	case <-timeout:
		t.Error("Timed out")
	case <-time.After(20 * time.Millisecond):
	}
}

type blockingWR struct{}
type invalidWR struct{}
type validWR struct{}

func (wr blockingWR) recv() (heartbeat, error) { select {}; return heartbeat{}, nil }
func (wr invalidWR) recv() (heartbeat, error)  { return heartbeat{UUID: "invalid"}, nil }
func (wr validWR) recv() (heartbeat, error)    { return heartbeat{UUID: "uuid"}, nil }
