package hml

import (
	"bufio"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
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

type resItem struct {
	ext string
	fr  FileResult
}

const (
	ansiCyan  = "\x1b[36m"
	ansiReset = "\x1b[0m"
)

var ignoreList []string
var extensions []string
var commentsList []string

func contains(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}

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

func WalkDirectory(root string) (map[string]ExtResult, error) {

	var files []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			log.Println("Error accessing path:", path, "error:", err)
			return err
		}

		// ignore list check
		if contains(ignoreList, d.Name()) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
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
		if len(extensions) > 0 && !contains(extensions, ext) {
			return nil
		}

		files = append(files, path)
		return nil
	})

	if err != nil {
		return nil, err
	}

	workerCount := max(runtime.NumCPU(), 1)

	jobs := make(chan string, len(files))
	resultsCh := make(chan resItem, len(files))
	errCh := make(chan error, 1)

	var wg sync.WaitGroup

	// Workers get files from jobs channel
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for p := range jobs {
				fr, perr := ParseFile(p)
				if perr != nil {
					select {
					case errCh <- perr:
					default:
					}
					return
				}
				ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(p), "."))
				resultsCh <- resItem{ext: ext, fr: fr}
			}
		}()
	}

	// Add files to jobs channel
	for _, f := range files {
		jobs <- f
	}
	close(jobs)

	go func() {
		wg.Wait()
		close(resultsCh)
	}()

	// Results aggregation
	results := make(map[string]ExtResult)
	for r := range resultsCh {
		er := results[r.ext]
		er.Files++
		er.CodeLines += r.fr.CodeLines
		er.CommLines += r.fr.CommLines
		results[r.ext] = er
	}

	select {
	case e := <-errCh:
		return results, e
	default:
		return results, nil
	}
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
	if len(results) == 0 {
		fmt.Println("No results")
		return
	}

	// Sorts the extensions
	exts := make([]string, 0, len(results))
	for k := range results {
		exts = append(exts, k)
	}
	sort.Strings(exts)

	// Headers
	hExt := "Extension"
	hFiles := "Files"
	hCode := "Code"
	hComm := "Comments"

	// Calculate column widths
	extW := len(hExt)
	filesW := len(hFiles)
	codeW := len(hCode)
	commW := len(hComm)

	for _, e := range exts {

		r := results[e]
		extW = max(len(e), extW)
		filesW = max(len(strconv.Itoa(r.Files)), filesW)
		codeW = max(len(strconv.Itoa(r.CodeLines)), codeW)
		commW = max(len(strconv.Itoa(r.CommLines)), commW)
	}

	// Make widths even (if odd, add 1)
	extW += 2 + extW%2
	filesW += 2 + filesW%2
	codeW += 2 + codeW%2
	commW += 2 + commW%2

	repeat := func(s string, count int) string {
		return strings.Repeat(s, count)
	}
	pad := func(s string, width int) string {
		if len(s) >= width {
			return s
		}
		spaces := width - len(s)
		leftPad := spaces / 2
		rightPad := spaces - leftPad
		return repeat(" ", leftPad) + s + repeat(" ", rightPad)
	}

	useColor := os.Getenv("NO_COLOR") == ""
	colorize := func(s string) string {
		if !useColor {
			return s
		}
		return ansiCyan + s + ansiReset
	}

	hLine := "+" + repeat("-", extW) + "+" + repeat("-", filesW) + "+" + repeat("-", codeW) + "+" + repeat("-", commW) + "+"
	fmt.Println(hLine)
	fmt.Printf("|%s|%s|%s|%s|\n", pad(hExt, extW), pad(hFiles, filesW), pad(hCode, codeW), pad(hComm, commW))

	fmt.Println(hLine)

	totalFiles := 0
	totalCode := 0
	totalComm := 0

	for _, e := range exts {
		r := results[e]

		totalFiles += r.Files
		totalCode += r.CodeLines
		totalComm += r.CommLines

		extCell := colorize(e) + repeat(" ", extW-len(e))
		filesCell := pad(strconv.Itoa(r.Files), filesW)
		codeCell := pad(strconv.Itoa(r.CodeLines), codeW)
		commCell := pad(strconv.Itoa(r.CommLines), commW)

		fmt.Printf("|%s|%s|%s|%s|\n", extCell, filesCell, codeCell, commCell)
	}

	fmt.Println(hLine)
	fmt.Printf("Total: Files=%d  Code=%d  Comments=%d\n", totalFiles, totalCode, totalComm)
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
