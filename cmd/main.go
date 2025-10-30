package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	hml "github.com/wert-rar/go-hml/internal/go-hml"
)

type Config struct {
	IgnoreList   string `json:"ignore_list"`
	Extensions   string `json:"extensions"`
	CommentsList string `json:"comments_list"`
	Quiet        *bool  `json:"quiet"`
}

func isFlagPassed(name string) bool {
	for _, arg := range os.Args[1:] {
		if strings.HasPrefix(arg, "-"+name) || strings.HasPrefix(arg, "--"+name) {
			return true
		}
	}
	return false
}

func main() {
	var ignoreList string
	var extensions string
	var commentsList string
	var quiet bool

	flag.StringVar(&ignoreList, "i", "", "comma-separated list of directories to ignore")
	flag.StringVar(&extensions, "e", "", "comma-separated list of file extensions to include (without dot)")
	flag.StringVar(&commentsList, "c", "//", "comma-separated list of comment syntax to consider (e.g., //,#,/* */)")
	flag.BoolVar(&quiet, "q", false, "otput only the final numbers of lines of code and comments")

	configPath := flag.String("config", "", "path to JSON config file (e.g. --config=config.json)")
	help := flag.Bool("help", false, "show help")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [flags] <path>\n\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()

	if *help {
		flag.Usage()
		os.Exit(0)
	}

	args := flag.Args()
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "error: path argument is required")
		flag.Usage()
		os.Exit(2)
	}
	path := args[0]

	if *configPath != "" {
		data, err := os.ReadFile(*configPath)
		if err != nil {
			log.Fatalf("cannot read config file %q: %v", *configPath, err)
		}
		var cfg Config
		if err := json.Unmarshal(data, &cfg); err != nil {
			log.Fatalf("invalid JSON in config file %q: %v", *configPath, err)
		}

		if cfg.IgnoreList != "" && !isFlagPassed("i") && !isFlagPassed("ignore") {
			ignoreList = cfg.IgnoreList
		}
		if cfg.Extensions != "" && !isFlagPassed("e") && !isFlagPassed("extensions") {
			extensions = cfg.Extensions
		}
		if cfg.CommentsList != "" && !isFlagPassed("c") && !isFlagPassed("comments") {
			commentsList = cfg.CommentsList
		}
		if cfg.Quiet != nil && !isFlagPassed("q") && !isFlagPassed("quiet") {
			quiet = *cfg.Quiet
		}
	}

	hml.SetIgnoreList(ignoreList)
	hml.SetExtensions(extensions)
	hml.SetCommentsList(commentsList)

	results, err := hml.WalkDirectory(path)
	if err != nil {
		log.Fatalf("scan failed: %v", err)
	}

	if quiet {
		hml.PrintQuiet(results)
		return
	}
	hml.PrintResults(results)
}
