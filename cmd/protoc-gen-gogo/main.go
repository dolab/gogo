package main

import (
	"flag"
	"fmt"
	"os"

	gen "github.com/dolab/gogo/internal/protoc-gen"
)

var (
	version bool
)

func main() {
	flag.BoolVar(&version, "version", false, "print version of protoc-gen-gogo plugin and exit")
	flag.Parse()
	if version {
		fmt.Println(gen.Version)
		os.Exit(0)
	}

	gen.Run(newGenerator())
}
