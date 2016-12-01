package main

import (
	"flag"
	"fmt"

	"github.com/pkg/profile"
)

var cpuprofile = flag.Bool("cpuprofile", false, "write cpu profile to file")
var bookieUrl = flag.String("bookieUrl", "", "bookie url, including port (optional)")

func main() {
	flag.Parse()
	if *cpuprofile {
		defer profile.Start().Stop()
	}

	port := 41234
	listener := mustGetListener(port)
	baseTemplate := mustParseBasePageTemplate()
	bookie := bookie{url: *bookieUrl}

	fmt.Printf("Flowbro is your bro on localhost:%v!\n", port)
	serve(&flowbro{}, baseTemplate, listener, bookie)
}
