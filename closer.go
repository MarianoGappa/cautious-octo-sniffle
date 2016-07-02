package main

import (
	"io"
	"log"
)

func closeAll(cls []io.Closer) []error {
	errors := []error{}
	log.Printf("Closing %v closeables used on this websocket\n", len(cls))
	for _, cl := range cls {
		if err := cl.Close(); err != nil {
			log.Printf("Failed to close: %v", err)
			errors = append(errors, err)
		} else {
			log.Printf("Closeable %v closed successfully!", cl)
		}
	}
	log.Printf("Finished trying to close %v closeables\n", len(cls))
	return errors
}
