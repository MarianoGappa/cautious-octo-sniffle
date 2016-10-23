package main

import (
	"fmt"
	"sync"
)

type errorlist struct {
	errors []error
	l      sync.Mutex
}

func (el *errorlist) add(s string) errorlist {
	el.l.Lock()
	el.errors = append(el.errors, fmt.Errorf(s))
	el.l.Unlock()
	return *el
}
