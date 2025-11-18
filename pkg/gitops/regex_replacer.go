package gitops

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

type RegexReplacer struct {
	// Destination repository for rendered files.
	DestinationRepo localRepository
	// Destination folder inside the repository for rendered files.
	DestinationFolder string
	// Files to search through for matches.
	Files []string
	// MatchRegex is the regex pattern to match.
	MatchRegex *regexp.Regexp
	// ReplaceTo is the string to replace the matched regex.
	ReplaceTo string
}

func (rr RegexReplacer) renderAllFiles() error {
	for _, file := range rr.Files {
		if err := rr.renderFile(file); err != nil {
			return fmt.Errorf("render file %q: %w", file, err)
		}
	}
	return nil
}

func (rr RegexReplacer) renderFile(fileName string) error {
	filePath := filepath.Join(
		rr.DestinationRepo.localPath(),
		rr.DestinationFolder,
		fileName,
	)
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}
	replaced := rr.MatchRegex.ReplaceAllString(string(fileContent), rr.ReplaceTo)
	if err := os.WriteFile(filePath, []byte(replaced), 0644); err != nil {
		return fmt.Errorf("write file: %w", err)
	}
	return nil
}
