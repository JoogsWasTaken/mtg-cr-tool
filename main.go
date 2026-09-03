package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	// e.g. "2. Parts of a Card"
	reChapter = regexp.MustCompile(`^[1-9]\.\s+[A-Za-z]`)
	// e.g. "203. Illustration"
	reSection = regexp.MustCompile(`^[1-9][0-9]{2}\.\s+[A-Za-z]`)
	// e.g. "212.1. Each vanguard ...", "213.1a Most card sets, ..."
	reRule         = regexp.MustCompile(`^([1-9][0-9]{2}\.[0-9]+[a-z]?\.?)\s*(.*)`)
	reSafeFilename = regexp.MustCompile(`[^a-zA-Z0-9_]`)
)

// injected via ldflags
var (
	version = "local development build"
	commit  = "no sha available"
)

func sanitizeFileName(name string) string {
	clean := strings.ReplaceAll(name, " ", "_")
	clean = reSafeFilename.ReplaceAllString(clean, "_")
	return strings.ToLower(clean) + ".md"
}

func main() {
	var (
		crPath      string
		outputDir   string
		showVersion bool
	)

	flag.StringVar(&crPath, "cr", "", "path to comprehensive rules file")
	flag.StringVar(&outputDir, "o", "", "path to output directory")
	flag.BoolVar(&showVersion, "v", false, "show version and exit")

	flag.Parse()

	if showVersion {
		fmt.Printf("%s (%s)\n", version, commit)
		os.Exit(0)
	}

	os.MkdirAll(outputDir, 0755)

	rulesFile, err := os.Open(crPath)
	if err != nil {
		log.Fatalf("failed to open comprehensive rules: %v\n", err)
	}
	defer func() {
		if err := rulesFile.Close(); err != nil {
			log.Fatalf("failed to close file handle: %v\n", err)
		}
	}()

	scanner := bufio.NewScanner(rulesFile)

	// large capacity for long lines
	const maxCapacity = 1024 * 1024

	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)

	inMainRules := false

	var currentFile *os.File
	var currentWriter *bufio.Writer
	closeCurrentFile := func() {
		if currentFile != nil {
			if currentWriter != nil {
				if err := currentWriter.Flush(); err != nil {
					log.Fatalf("failed to flush writer: %v\n", err)
				}
			}

			if currentFile != nil {
				if err := currentFile.Close(); err != nil {
					log.Fatalf("failed to close file handle: %v\n", err)
				}
			}
		}
	}
	defer closeCurrentFile()

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// "Glossary" is where the main part ends
		if inMainRules && line == "Glossary" {
			break
		}

		// wait for rules to start
		if !inMainRules {
			// "Credits" is the last line in the table of contents
			if line == "Credits" {
				inMainRules = true
			}

			continue
		}

		// skip empty lines
		if line == "" {
			continue
		}

		// if starting a new chapter, open a new file and write heading
		if reChapter.MatchString(line) {
			closeCurrentFile()

			fileName := sanitizeFileName(line)
			filePath := filepath.Join(outputDir, fileName)

			f, err := os.Create(filePath)
			if err != nil {
				log.Fatalf("failed to create file at %s: %v\n", filePath, err)
			}

			currentFile = f
			currentWriter = bufio.NewWriter(currentFile)

			log.Printf("writing section `%s` to %s\n", line, filePath)
			// write chapter as h1
			currentWriter.WriteString(fmt.Sprintf("# %s\n\n", line))

			continue
		}

		// skip if writer is not set
		if currentWriter == nil {
			continue
		}

		// write section as h2
		if reSection.MatchString(line) {
			currentWriter.WriteString(fmt.Sprintf("## %s\n\n", line))
			continue
		}

		if matches := reRule.FindStringSubmatch(line); len(matches) > 2 {
			// write rule as line with rule number in bold
			ruleNum := matches[1]
			ruleText := matches[2]

			currentWriter.WriteString(fmt.Sprintf("**%s** %s\n\n", ruleNum, ruleText))
			continue
		}

		if strings.HasPrefix(line, "Example:") {
			// write example as blockquote
			currentWriter.WriteString(fmt.Sprintf("> *%s*\n\n", line))
			continue
		}

		// write as regular text
		currentWriter.WriteString(fmt.Sprintf("%s\n\n", line))
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("error while scanning file: %v\n", err)
	}
}
