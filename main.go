package main

import (
	"flag"
	"fmt"

	"github.com/pkg/profile"
)

var cpuprofile = flag.Bool("cpuprofile", false, "write cpu profile to file")

func main() {
	flag.Parse()
	if *cpuprofile {
		defer profile.Start().Stop()
	}

	port := 41234
	listener := mustGetListener(port)
	baseTemplate := mustParseBasePageTemplate()

	fmt.Printf("Flowbro is your bro on localhost:%v!\n", port)
	serve(&flowbro{}, baseTemplate, listener)
}
