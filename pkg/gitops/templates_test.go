package gitops

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Possible templates to render.
const (
	templateValuesYAML = `---
api-service:
  image:
    name: "{{ .repository }}:{{ .tag }}"
`
	templateChartYAML = `---
apiVersion: v2
name: test
description: A test service
type: application
version: 0.1.0
appVersion: {{ .appVersion }}
dependencies:
- name: api-service
  version: "0.1.1"
  repository: "https://bitrise-io.github.io/k8s-recipes/"
`
	templateBrewFormulae = `
class BitriseDenAgent < Formula
	desc "CLI for Bitrise DEN agent"
	homepage "https://github.com/bitrise-io/bitrise-den-agent"
	url "https://github.com/bitrise-io/bitrise-den-agent.git",
	  tag:      "{{ .brew_forumlae_tag }}",
	  revision: "{{ .brew_forumlae_revision }}"
	license ""
end  
`
)

// Possible outputs of rendered templates.
const (
	myRenderedValuesYAML = `---
api-service:
  image:
    name: "myrepo:mytag"
`
	otherRenderedValuesYAML = `---
api-service:
  image:
    name: "foo:bar"
`
	renderedChartYAML = `---
apiVersion: v2
name: test
description: A test service
type: application
version: 0.1.0
appVersion: 2.4.5
dependencies:
- name: api-service
  version: "0.1.1"
  repository: "https://bitrise-io.github.io/k8s-recipes/"
`
	renderedBrewFormulae = `
class BitriseDenAgent < Formula
	desc "CLI for Bitrise DEN agent"
	homepage "https://github.com/bitrise-io/bitrise-den-agent"
	url "https://github.com/bitrise-io/bitrise-den-agent.git",
	  tag:      "v1.2.3",
	  revision: "aaabbbcccdddeeefff"
	license ""
end  
`
)

var renderAllFilesCases = map[string]struct {
	templates map[string]string
	values    map[string]string
	folder    string
	wantFiles map[string]string
	wantErr   bool
}{
	"only values.yaml is rendered": {
		templates: map[string]string{"values.yaml": templateValuesYAML},
		values:    map[string]string{"repository": "myrepo", "tag": "mytag"},
		folder:    "folder-with-values-yaml",
		wantFiles: map[string]string{"values.yaml": myRenderedValuesYAML},
	},
	"values.yaml is rendered with other input": {
		templates: map[string]string{"values.yaml": templateValuesYAML},
		values:    map[string]string{"repository": "foo", "tag": "bar"},
		folder:    "another-folder-with-values-yaml",
		wantFiles: map[string]string{"values.yaml": otherRenderedValuesYAML},
	},
	"only Chart.yaml is rendered": {
		templates: map[string]string{"Chart.yaml": templateChartYAML},
		values:    map[string]string{"appVersion": "2.4.5"},
		folder:    "folder-with-chart-yaml",
		wantFiles: map[string]string{"Chart.yaml": renderedChartYAML},
	},
	"both Chart.yaml and values.yaml are rendered": {
		templates: map[string]string{
			"values.yaml": templateValuesYAML,
			"Chart.yaml":  templateChartYAML,
		},
		values: map[string]string{
			"repository": "myrepo",
			"tag":        "mytag",
			"appVersion": "2.4.5",
		},
		folder: "folder-with-multiple-files",
		wantFiles: map[string]string{
			"values.yaml": myRenderedValuesYAML,
			"Chart.yaml":  renderedChartYAML,
		},
	},
	"an unused template variable is present (it's ignored)": {
		templates: map[string]string{"Chart.yaml": templateChartYAML},
		values:    map[string]string{"appVersion": "2.4.5", "un": "used"},
		folder:    "folder-with-unused-values",
		wantFiles: map[string]string{"Chart.yaml": renderedChartYAML},
	},
	"a template variable is missing (error)": {
		templates: map[string]string{"Chart.yaml": templateChartYAML},
		values:    map[string]string{"appVersionTypo": "2.4.5"},
		folder:    "wont-use-this-folder",
		wantErr:   true,
	},
	"html tags are correctly rendered in brew formulae": {
		templates: map[string]string{"formulae.rb": templateBrewFormulae},
		values:    map[string]string{"brew_forumlae_tag": "v1.2.3", "brew_forumlae_revision": "aaabbbcccdddeeefff"},
		folder:    "another-folder-with-values-yaml",
		wantFiles: map[string]string{"formulae.rb": renderedBrewFormulae},
	},
}

func TestRenderAllFiles(t *testing.T) {
	for name, tc := range renderAllFilesCases {
		t.Run(name, func(t *testing.T) {
			// Create temporary directory for templates.
			templatesDir, err := os.MkdirTemp("", "")
			require.NoError(t, err, "new temp templates dir")
			defer os.RemoveAll(templatesDir)

			// Create a mock temporary directory for local clone of repository.
			renderRepo, err := os.MkdirTemp("", "")
			require.NoError(t, err, "new temp render repo")
			defer os.RemoveAll(renderRepo)
			// Create directory inside that for rendered files.
			renderDir := path.Join(renderRepo, tc.folder)
			require.NoError(t, os.Mkdir(renderDir, 0700), "new temp render dir")

			// Copy desired templates to the previously created temp directory.
			for fileName, content := range tc.templates {
				filePath := path.Join(templatesDir, fileName)
				err := os.WriteFile(filePath, []byte(content), 0600)
				require.NoError(t, err, "write template %q", fileName)
			}

			// Run Templates.renderAllFiles.
			tr := Templates{
				SourceFolder: templatesDir,
				Values:       tc.values,
				DestinationRepo: &localRepositoryMock{
					localPathFunc: func() string {
						return renderRepo
					},
				},
				DestinationFolder: tc.folder,
			}

			// Assert for error.
			gotErr := tr.renderAllFiles()
			if tc.wantErr {
				require.Error(t, gotErr, "templatesRenderer.renderAllFiles")
				return
			}
			require.NoError(t, gotErr, "templatesRenderer.renderAllFiles")

			// Assert for name of rendered files.
			var wantFileNames []string
			for name := range tc.wantFiles {
				wantFileNames = append(wantFileNames, name)
			}

			var gotFileNames []string
			gotFileInfos, err := os.ReadDir(renderDir)
			require.NoError(t, err, "read files of render dir")
			for _, v := range gotFileInfos {
				gotFileNames = append(gotFileNames, v.Name())
			}

			require.ElementsMatch(t, wantFileNames, gotFileNames, "file names")

			// Assert for contents of rendered files.
			for fileName, want := range tc.wantFiles {
				filePath := path.Join(renderDir, fileName)
				got, err := os.ReadFile(filePath)
				require.NoError(t, err, "read contents of %q", filePath)
				assert.EqualValues(t, want, string(got), "contents of %q", fileName)
			}
		})
	}
}
