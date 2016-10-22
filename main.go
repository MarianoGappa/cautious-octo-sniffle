package main

import "fmt"

func main() {
	port := mustResolvePort(41234)
	listener := mustGetListener(port)
	baseTemplate := mustParseBasePageTemplate()

	fmt.Printf("Flowbro is your bro on localhost:%v!\n", port)
	serve(&flowbro{}, baseTemplate, listener)
}
