package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/docker/pkg/fileutils"
)

func ReadAll(reader io.Reader) ([]string, error) {
	if reader == nil {
		return nil, nil
	}

	scanner := bufio.NewScanner(reader)
	var excludes []string
	currentLine := 0

	utf8bom := []byte{0xEF, 0xBB, 0xBF}
	for scanner.Scan() {
		scannedBytes := scanner.Bytes()
		// We trim UTF8 BOM
		if currentLine == 0 {
			scannedBytes = bytes.TrimPrefix(scannedBytes, utf8bom)
		}
		pattern := string(scannedBytes)
		currentLine++
		// Lines starting with # (comments) are ignored before processing
		if strings.HasPrefix(pattern, "#") {
			continue
		}
		pattern = strings.TrimSpace(pattern)
		if pattern == "" {
			continue
		}
		pattern = filepath.Clean(pattern)
		pattern = filepath.ToSlash(pattern)
		excludes = append(excludes, pattern)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("Error reading ignore file: %v", err)
	}
	return excludes, nil
}

func Filter(ignoreFile string, filesToTest []string) ([]string, error) {
	f, err := os.Open(ignoreFile)
	// Note that a missing ignore file isn't treated as an error
	if err != nil {
		if os.IsNotExist(err) {
			return filesToTest, nil
		}
		return nil, err
	}
	excludes, _ := ReadAll(f)
	f.Close()

        unignored := []string{}
	for _, fileToTest := range filesToTest {
		shouldIgnore, _ := fileutils.Matches(fileToTest, excludes)
		if !shouldIgnore {
			unignored = append(unignored, fileToTest)
		}
	}
	return unignored, nil
}

func main() {
	ignoreFile := flag.String("f", "", "ignore file to parse and test against")
	flag.Parse()
	if *ignoreFile == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	files := []string{}
	if flag.NArg() == 0 {
		// read from stdin
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Split(bufio.ScanWords)
		for scanner.Scan() {
			files = append(files, scanner.Text())
		}
	} else {
		for _, val := range flag.Args() {
			files = append(files, val)
		}
	}

	filteredFiles, err := Filter(*ignoreFile, files)
	if err != nil {
		fmt.Errorf("Error filtering test files: %v", err)
		os.Exit(1)
	}
	for _, file := range filteredFiles {
		fmt.Println(file)
	}
}
