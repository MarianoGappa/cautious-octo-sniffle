package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime/pprof"
)

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")

func main() {
	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {

			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	port := mustResolvePort(41234)
	listener := mustGetListener(port)
	baseTemplate := mustParseBasePageTemplate()

	fmt.Printf("Flowbro is your bro on localhost:%v!\n", port)
	serve(&flowbro{}, baseTemplate, listener)
}
