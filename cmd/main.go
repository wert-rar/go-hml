package main

import (
	"flag"
	"log"

	"github.com/wert-rar/go-hml/internal/go-hml"
)

func main() {
	path := flag.String("p", ".", "path to scan")
	flag.Parse()
	results, err := hml.WalkDirectory(*path)
	if err != nil {
		log.Fatalf("scan failed: %v", err)
	}
	hml.PrintResults(results)
}
