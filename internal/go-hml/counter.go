package hml

import (
	"bufio"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type FileResult struct {
	Ext       string
	CodeLines int
	CommLines int
}

type ExtResult struct {
	Files     int
	CodeLines int
	CommLines int
}

// walks in dir
func WalkDirectory(path string) (map[string]ExtResult, error) {
	results := make(map[string]ExtResult)

	err := filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			log.Fatal(err)
		}
		if d.IsDir() {
			return nil
		} else {
			name := d.Name()
			var ext string
			if strings.Contains(name, ".") {
				ext = strings.Split(name, ".")[1]
			} else {
				return nil
			}

			result, err := ParseFile(path)
			if err != nil {
				log.Fatal(err)
			}

			extResult, ok := results[ext]
			if !ok {
				extResult = ExtResult{}
			}
			extResult.Files += 1
			extResult.CodeLines += result.CodeLines
			extResult.CommLines += result.CommLines

			results[ext] = extResult

		}

		return nil
	})
	return results, err
}

func ParseFile(name string) (FileResult, error) {
	f, err := os.Open(name)
	if err != nil {
		return FileResult{}, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	result := FileResult{}

	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		} else if strings.HasPrefix(line, "//") {
			result.CommLines++
		} else {
			result.CodeLines++
		}
	}

	if err := scanner.Err(); err != nil {
		return FileResult{}, err
	}

	return result, nil

}

func PrintResults(results map[string]ExtResult) {
	fmt.Println("RESULTS:")

	for ext, result := range results {
		fmt.Printf("\t%s :  Files: %d  Code %d lines  Comments %d lines",
			ext, result.Files, result.CodeLines, result.CommLines)
		fmt.Println()

	}

}
