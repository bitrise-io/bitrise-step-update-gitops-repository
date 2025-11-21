package gitops

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

//go:generate moq -out templates_moq_test.go . AllFilesRenderer
type AllFilesRenderer interface {
	renderAllFiles() error
}

// templates implements the allFilesRenderer interface.
var _ AllFilesRenderer = (*Templates)(nil)

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

// MultiTemplates renders a folder of templates to multiple destinations in a local repository.
type MultiTemplates struct {
	// Source folder of templates.
	SourceFolder string
	// Deployments contains multiple deployment configurations.
	Deployments []Deployment
	// Destination repository for rendered files.
	DestinationRepo localRepository
}

// Ensure MultiTemplates implements AllFilesRenderer
var _ AllFilesRenderer = (*MultiTemplates)(nil)

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

func (mtr MultiTemplates) renderAllFiles() error {
	// Render templates for each deployment
	for i, deployment := range mtr.Deployments {
		if err := mtr.renderDeployment(deployment); err != nil {
			return fmt.Errorf("render deployment[%d] to %q: %w", i, deployment.Path, err)
		}
	}
	return nil
}

func (mtr MultiTemplates) renderDeployment(deployment Deployment) error {
	// Get all template file names from the source folder.
	files, err := os.ReadDir(mtr.SourceFolder)
	if err != nil {
		return fmt.Errorf("read files in %q: %w", mtr.SourceFolder, err)
	}

	// Render templates one-by-one to the destination folder
	// (substituting values given for this deployment).
	for _, file := range files {
		if err := mtr.renderFile(file.Name(), deployment); err != nil {
			return fmt.Errorf("render file %q: %w", file.Name(), err)
		}
	}
	return nil
}

func (mtr MultiTemplates) renderFile(fileName string, deployment Deployment) error {
	// Parse template.
	sourceFilePath := filepath.Join(mtr.SourceFolder, fileName)
	t, err := template.ParseFiles(sourceFilePath)
	if err != nil {
		return fmt.Errorf("parse template %q: %w", sourceFilePath, err)
	}

	// Create a file for the rendered template.
	destinationFilePath := filepath.Join(
		mtr.DestinationRepo.localPath(), deployment.Path, fileName)

	// Ensure the destination directory exists
	destinationDir := filepath.Dir(destinationFilePath)
	if err := os.MkdirAll(destinationDir, 0755); err != nil {
		return fmt.Errorf("create destination directory: %w", err)
	}

	f, err := os.Create(destinationFilePath)
	if err != nil {
		return fmt.Errorf("create destination file: %w", err)
	}
	defer f.Close()

	// Render the template to the previously created file.
	if err := t.Option("missingkey=error").Execute(f, deployment.Values); err != nil {
		return fmt.Errorf("execute template %q: %w", sourceFilePath, err)
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
