package gitops

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

//go:generate moq -out templates_moq_test.go . allFilesRenderer
type allFilesRenderer interface {
	renderAllFiles() error
}

// templates implements the allFilesRenderer interface.
var _ allFilesRenderer = (*Templates)(nil)

// Templates renders a folder of templates to a local repository.
type Templates struct {
	// Source folder of templates.
	SourceFolder string
	// Values to substitute into the templates.
	Values map[string]string
	// Destination repository for rendered files.
	DestinationRepo localRepository
	// Destination folder inside the repository for rendered files.
	DestinationFolder string
}

func (tr Templates) renderAllFiles() error {
	// Get all template file names from the source folder.
	files, err := os.ReadDir(tr.SourceFolder)
	if err != nil {
		return fmt.Errorf("read files in %q: %w", tr.SourceFolder, err)
	}

	// Render templates one-by-one to the destinaton folder
	// (substituting values given).
	for _, file := range files {
		if err := tr.renderFile(file.Name()); err != nil {
			return fmt.Errorf("render file %q: %w", file.Name(), err)
		}
	}
	return nil
}

func (tr Templates) renderFile(fileName string) error {
	// Parse template.
	sourceFilePath := filepath.Join(tr.SourceFolder, fileName)
	t, err := template.ParseFiles(sourceFilePath)
	if err != nil {
		return fmt.Errorf("parse template %q: %w", sourceFilePath, err)
	}

	// Create a file for the rendered template.
	destinationFilePath := filepath.Join(
		tr.DestinationRepo.localPath(), tr.DestinationFolder, fileName)
	f, err := os.Create(destinationFilePath)
	if err != nil {
		return fmt.Errorf("create destination file: %w", err)
	}

	// Render the template to the previously created file.
	if err := t.Option("missingkey=error").Execute(f, tr.Values); err != nil {
		return fmt.Errorf("execute template %q: %w", sourceFilePath, err)
	}
	return nil
}
