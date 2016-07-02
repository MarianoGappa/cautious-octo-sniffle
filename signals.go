package main

import (
	"os"
	"os/signal"
	"syscall"
)

func mustListenToSignals(q chan struct{}) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-signals
		close(q)
	}()
}
