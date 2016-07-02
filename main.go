package main

import "fmt"

func main() {
	q := make(chan struct{})

	port := mustResolvePort(41234)
	listener := mustGetListener(port, q)
	baseTemplate := mustParseBasePageTemplate()
	mustListenToSignals(q)

	fmt.Printf("Flowbro is your bro on localhost:%v!\n", port)
	serve(&flowbro{}, q, baseTemplate, listener)
}
