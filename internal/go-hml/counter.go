package hml

import (
	"bufio"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"slices"
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

var ignoreList []string
var extensions []string
var commentsList []string

func parseCommaSeparatedList(list string) []string {
	if list == "" {
		return []string{}
	}
	parts := strings.Split(list, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}

func SetIgnoreList(list string) {
	ignoreList = parseCommaSeparatedList(list)
}

func SetExtensions(list string) {
	extensions = parseCommaSeparatedList(list)
}

func SetCommentsList(list string) {
	commentsList = parseCommaSeparatedList(list)
}

// walks in dir
func WalkDirectory(path string) (map[string]ExtResult, error) {
	results := make(map[string]ExtResult)

	err := filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			log.Println("Error accessing path:", path, "error:", err)
			return err
		}

		// ignore list check
		if slices.Contains(ignoreList, d.Name()) {
			return filepath.SkipDir
		}

		if d.IsDir() {
			return nil
		}

		ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(d.Name()), "."))

		// extension empty-filter check
		if ext == "" {
			return nil
		}
		// extension filter check
		if len(extensions) > 0 && !slices.Contains(extensions, ext) {
			return nil
		}

		result, err := ParseFile(path)
		if err != nil {
			return err
		}

		er := results[ext]
		er.Files++
		er.CodeLines += result.CodeLines
		er.CommLines += result.CommLines
		results[ext] = er

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
		trimmed := strings.TrimSpace(line)

		if trimmed == "" {
			continue
		}
		isComment := false

		// default comment syntax check
		if len(commentsList) == 0 {
			if strings.HasPrefix(trimmed, "//") {
				isComment = true
			}
		} else {
			for _, commentSyntax := range commentsList {
				if strings.HasPrefix(trimmed, commentSyntax) {
					isComment = true
					break
				}
			}
		}
		if isComment {
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

func PrintQuiet(results map[string]ExtResult) {
	totalCode := 0
	totalComm := 0
	for _, result := range results {
		totalCode += result.CodeLines
		totalComm += result.CommLines
	}
	fmt.Printf("%d (code: %d comments: %d)\n", totalCode+totalComm, totalCode, totalComm)

}
