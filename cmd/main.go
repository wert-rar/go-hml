package main

import (
	"flag"
	"log"

	"github.com/wert-rar/go-hml/internal/go-hml"
)

func main() {
	path := flag.String("p", ".", "path to scan")
	ignoreList := flag.String("i", "", "comma-separated list of directories to ignore")
	extensions := flag.String("e", "", "comma-separated list of file extensions to include (without dot)")
	commentsList := flag.String("c", "//", "comma-separated list of comment syntax to consider (e.g., //, #, /* */)")
	quiet := flag.Bool("q", false, "Output only the final numbers of lines of code and comments")

	flag.Parse()

	hml.SetIgnoreList(*ignoreList)
	hml.SetExtensions(*extensions)
	hml.SetCommentsList(*commentsList)

	results, err := hml.WalkDirectory(*path)
	if err != nil {
		log.Fatalf("scan failed: %v", err)
	}

	if *quiet {
		hml.PrintQuiet(results)
		return
	}
	hml.PrintResults(results)
}
