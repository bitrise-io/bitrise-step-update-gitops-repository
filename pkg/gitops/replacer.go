package gitops

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path"
	"regexp"
	"strings"
)

type Replacer struct {
	// Delimiter to look for when rendering key-value pairs
	Delimiter string
	// Destination repository for rendered files.
	DestinationRepo localRepository
	// Values to substitute into the templates.
	Values map[string]string
	// Files to search through for matches
	Files []string
	// Destination folder inside the repository for rendered files.
	DestinationFolder string
}

func (rp Replacer) renderAllFiles() error {
	// Replace values in files sitting in the destinaton folder one-by-one
	// (substituting values given).
	for _, file := range rp.Files {
		originalFile := path.Join(rp.DestinationRepo.localPath(), rp.DestinationFolder, file)
		renderedFile, err := rp.renderFile(originalFile)

		if err != nil {
			return fmt.Errorf("render file %q: %w", originalFile, err)
		}
		defer os.Remove(renderedFile)

		// the rendered / replaced file becomes the source
		source, err := os.Open(renderedFile)
		if err != nil {
			return fmt.Errorf("open rendered file: %w", err)
		}
		defer source.Close()

		// the original file location becomes the destionation
		destination, err := os.Create(originalFile)
		if err != nil {
			return fmt.Errorf("open destination file: %w", err)
		}

		// write the updated file contents to the original location
		if err := copy(source, destination); err != nil {
			return fmt.Errorf("saving rendered file: %w", err)
		}
	}
	return nil
}

func (rp Replacer) renderFile(fileName string) (string, error) {
	// Open file in local repository.
	f, err := os.Open(fileName)
	if err != nil {
		return "", fmt.Errorf("open file %s: %w", fileName, err)
	}
	defer f.Close()

	// Create temporary file.
	tmpFile, err := os.CreateTemp("", "")
	if err != nil {
		return "", fmt.Errorf("creating tmp file %s: %w", tmpFile.Name(), err)
	}
	defer tmpFile.Close()

	w := bufio.NewWriter(tmpFile)

	// Read file and replace occurances line by line.
	reader := bufio.NewReader(f)
	var EOF bool
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				return "", fmt.Errorf("reading lines : %w", err)
			}
			EOF = true
		}
		replacedLine, err := replaceAll(line, rp.Delimiter, rp.Values)
		if err != nil {
			return "", fmt.Errorf("could replace values in file (%s) line (%s): %w", fileName, line, err)
		}
		w.WriteString(replacedLine) // nolint:errcheck

		if strings.HasSuffix(fileName, "values.ci.yaml") {
			fmt.Println(replacedLine)
		}
		if EOF {
			break
		}
	}
	w.Flush()

	return tmpFile.Name(), nil
}

// buffered copy to handle large files.
func copy(src *os.File, dst *os.File) error {
	buf := make([]byte, 1024)
	for {
		n, err := src.Read(buf)
		if err != nil && err != io.EOF {
			return fmt.Errorf("error reading file: %w", err)
		}
		if n == 0 {
			break
		}

		if _, err := dst.Write(buf[:n]); err != nil {
			return fmt.Errorf("error writing file: %w", err)
		}
	}
	return nil
}

func replaceAll(line string, delimiter string, values map[string]string) (string, error) {
	replaced := line
	for k, v := range values {
		replaced = replaceOccurance(replaced, k, v, delimiter)
	}

	return replaced, nil
}

// matches key-value pair with provided delimiter, by extracting end replacing the value part.
// the value matching stops at the first comma, exclamation mark,
// single quote, double quote or whitespace occurance.
// Example:
//
//	provided content: foo=bar
//	provided delimiter: =
//	key: foo
//	newValue: qux
//
// The function will match bar, stop at the end of the string, then replace
// bar with qux.
func replaceOccurance(content string, key string, value string, delimiter string) string {
	pattern := fmt.Sprintf("%s%s[^,!'\"\\s]+", key, delimiter)
	m := regexp.MustCompile(pattern)
	tmpl := fmt.Sprintf("%s%s%s", key, delimiter, value)
	res := m.ReplaceAllString(content, tmpl)
	return res
}
