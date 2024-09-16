package gitops

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReplaceAll(t *testing.T) {
	tcs := map[string]struct {
		line      string
		delimiter string
		values    map[string]string
		wantLine  string
	}{
		"no match, line stays the same": {
			line:      "DEN_MACOS_VERSION=10",
			delimiter: "=",
			values: map[string]string{
				"DEN_LINUX_VERSION": "15",
				"SOME_OTHER_KEY":    "SOME_OTHER_VALUE",
				"SECRET":            "VALUE",
			},
			wantLine: "DEN_MACOS_VERSION=10",
		},
		"one match, line is updated": {
			line:      "DEN_MACOS_VERSION=10",
			delimiter: "=",
			values: map[string]string{
				"DEN_MACOS_VERSION": "15",
			},
			wantLine: "DEN_MACOS_VERSION=15",
		},
		"multiple matches, line is updated": {
			line:      "some text before the variable, DEN_MACOS_VERSION=10 some text after the variable, DEN_LINUX_VERSION=9 some other text",
			delimiter: "=",
			values: map[string]string{
				"DEN_MACOS_VERSION": "v2.0.0",
				"DEN_LINUX_VERSION": "v2.1.0",
			},
			wantLine: "some text before the variable, DEN_MACOS_VERSION=v2.0.0 some text after the variable, DEN_LINUX_VERSION=v2.1.0 some other text",
		},
		"complex delimiter, multiple matches, line is updated": {
			line: "some text before the variable, DEN_MACOS_VERSION: 10 some text after the variable, DEN_LINUX_VERSION: 9 some other text",
			// notice the whitespace
			delimiter: ": ",
			values: map[string]string{
				"DEN_MACOS_VERSION": "v2.0.0",
				"DEN_LINUX_VERSION": "v2.1.0",
			},
			wantLine: "some text before the variable, DEN_MACOS_VERSION: v2.0.0 some text after the variable, DEN_LINUX_VERSION: v2.1.0 some other text",
		},
		"works with single quotes": {
			line:      "'us-region-docker.pkg.dev/ip-dev/build/service:hello'",
			delimiter: ":",
			values: map[string]string{
				"us-region-docker.pkg.dev/ip-dev/build/service": "world",
			},
			wantLine: "'us-region-docker.pkg.dev/ip-dev/build/service:world'",
		},
		"works with double quotes": {
			line:      "\"us-region-docker.pkg.dev/ip-dev/build/service:hello\"",
			delimiter: ":",
			values: map[string]string{
				"us-region-docker.pkg.dev/ip-dev/build/service": "world",
			},
			wantLine: "\"us-region-docker.pkg.dev/ip-dev/build/service:world\"",
		},
		"works with comma": {
			line:      "DEN_MACOS_VERSION=10,",
			delimiter: "=",
			values: map[string]string{
				"DEN_MACOS_VERSION": "v1.0.0",
			},
			wantLine: "DEN_MACOS_VERSION=v1.0.0,",
		},
		"works with exclamation mark": {
			line:      "DEN_MACOS_VERSION=10!",
			delimiter: "=",
			values: map[string]string{
				"DEN_MACOS_VERSION": "v1.0.0",
			},
			wantLine: "DEN_MACOS_VERSION=v1.0.0!",
		},
		"multiple matches with command and double quotes": {
			line:      "\"DEN_MACOS_VERSION=10, DEN_LINUX_VERSION=9\"",
			delimiter: "=",
			values: map[string]string{
				"DEN_MACOS_VERSION": "v2.0.0",
				"DEN_LINUX_VERSION": "v2.1.0",
			},
			wantLine: "\"DEN_MACOS_VERSION=v2.0.0, DEN_LINUX_VERSION=v2.1.0\"",
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			newLine, err := replaceAll(tc.line, tc.delimiter, tc.values)

			require.NoError(t, err, "replace all must return no errors")
			require.Equal(t, tc.wantLine, newLine)
		})
	}
}

func TestReplacerRenderFile(t *testing.T) {
	tcs := map[string]struct {
		content     string
		delimiter   string
		values      map[string]string
		wantContent string
	}{
		"file with identation rendered correctly": {
			content: `---
global:
  image:
    name: gce.repository.com:v1.0.0
`,
			delimiter: ":",
			values: map[string]string{
				"gce.repository.com": "v1.0.1",
			},
			wantContent: `---
global:
  image:
    name: gce.repository.com:v1.0.1
`,
		},
		"example rendered correctly": {
			content: `
DEN_MACOS_VERSION=v2.0.0
DEN_MACOS_x86_SHASUM=9891b3a858c802d5729da26c9d683c65691ed2d6f3f56d3025c41c14b586816c
DEN_LINUX_VERSION=v2.5.9
DEN_LINUX_x86_SHASUM=de99c96ba2eb9bdcc97569ec79395d5321335b2aeaff71a93b1c57df6e1dc779`,
			delimiter: "=",
			values: map[string]string{
				"DEN_MACOS_VERSION":    "v2.1.0",
				"DEN_MACOS_x86_SHASUM": "02c6846000240dd25a5a084133495e88401fd564b4edc2a258d1545c2a20212d",
				"DEN_LINUX_VERSION":    "v2.6.0",
				"DEN_LINUX_x86_SHASUM": "ec509e440a0fc64f9840c8e192f5e4573da972a93f3579c527b4dae94542daf5",
			},
			wantContent: `
DEN_MACOS_VERSION=v2.1.0
DEN_MACOS_x86_SHASUM=02c6846000240dd25a5a084133495e88401fd564b4edc2a258d1545c2a20212d
DEN_LINUX_VERSION=v2.6.0
DEN_LINUX_x86_SHASUM=ec509e440a0fc64f9840c8e192f5e4573da972a93f3579c527b4dae94542daf5`,
		},
		"empty content, expect the same": {
			content:   "",
			delimiter: "",
			values: map[string]string{
				"DEN_MACOS_VERSION":    "v2.1.0",
				"DEN_MACOS_x86_SHASUM": "02c6846000240dd25a5a084133495e88401fd564b4edc2a258d1545c2a20212d",
				"DEN_LINUX_VERSION":    "v2.6.0",
				"DEN_LINUX_x86_SHASUM": "ec509e440a0fc64f9840c8e192f5e4573da972a93f3579c527b4dae94542daf5",
			},
			wantContent: "",
		},
		"one liner content rendered correctly": {
			content:   "HELLO: WORLD",
			delimiter: ": ",
			values: map[string]string{
				"HELLO": "BITRISE",
			},
			wantContent: "HELLO: BITRISE",
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			source, err := os.CreateTemp("", "")
			require.NoError(t, err, "create temp source file")
			defer source.Close()
			defer os.Remove(source.Name())

			err = os.WriteFile(source.Name(), []byte(tc.content), 0600)
			require.NoError(t, err, "write temp source file")

			replacer := Replacer{
				Delimiter: tc.delimiter,
				Values:    tc.values,
				DestinationRepo: &localRepositoryMock{
					localPathFunc: func() string {
						return ""
					},
				},
				DestinationFolder: path.Dir(source.Name()),
				Files: []string{
					source.Name(),
				},
			}

			rendered, err := replacer.renderFile(source.Name())
			require.NoError(t, err, "renderFile")
			defer os.Remove(rendered)

			b, err := os.ReadFile(rendered)
			require.NoError(t, err, "reading rendered file")

			require.Equal(t, tc.wantContent, string(b))
		})
	}
}

func TestReplacerRenderAllFiles(t *testing.T) {
	tcs := map[string]struct {
		contents     map[string]string
		delimiter    string
		values       map[string]string
		wantContents map[string]string
	}{
		"example rendered correctly": {
			contents: map[string]string{
				"file1": `DEN_MACOS_VERSION=v2.0.0
DEN_MACOS_x86_SHASUM=9891b3a858c802d5729da26c9d683c65691ed2d6f3f56d3025c41c14b586816c
`,
				"file2": `DEN_LINUX_VERSION=v2.5.9
DEN_LINUX_x86_SHASUM=de99c96ba2eb9bdcc97569ec79395d5321335b2aeaff71a93b1c57df6e1dc779
`,
			},
			delimiter: "=",
			values: map[string]string{
				"DEN_MACOS_VERSION":    "v2.1.0",
				"DEN_MACOS_x86_SHASUM": "02c6846000240dd25a5a084133495e88401fd564b4edc2a258d1545c2a20212d",
				"DEN_LINUX_VERSION":    "v2.6.0",
				"DEN_LINUX_x86_SHASUM": "ec509e440a0fc64f9840c8e192f5e4573da972a93f3579c527b4dae94542daf5",
			},
			wantContents: map[string]string{
				"file1": `DEN_MACOS_VERSION=v2.1.0
DEN_MACOS_x86_SHASUM=02c6846000240dd25a5a084133495e88401fd564b4edc2a258d1545c2a20212d
`,
				"file2": `DEN_LINUX_VERSION=v2.6.0
DEN_LINUX_x86_SHASUM=ec509e440a0fc64f9840c8e192f5e4573da972a93f3579c527b4dae94542daf5
`,
			},
		},
		"real life example rendered correctly": {
			contents: map[string]string{
				"file1": `---
stateless-service:
  image:
    name: "us-region-docker.pkg.dev/ip-dev/build/service:hello"
`,
			},
			delimiter: ":",
			values: map[string]string{
				"us-region-docker.pkg.dev/ip-dev/build/service": "world",
			},
			wantContents: map[string]string{
				"file1": `---
stateless-service:
  image:
    name: "us-region-docker.pkg.dev/ip-dev/build/service:world"
`,
			},
		},
	}

	testFolderName := "./testfiles"
	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			require.NoError(t, os.MkdirAll(testFolderName, 0700), "creating local folder")
			defer os.RemoveAll(testFolderName)
			files := make([]string, 0)
			for fn, content := range tc.contents {
				err := os.WriteFile("./testfiles/"+fn, []byte(content), 0600)
				require.NoError(t, err, "write temp source file")
				files = append(files, fn)
			}

			replacer := Replacer{
				Delimiter: tc.delimiter,
				Values:    tc.values,
				DestinationRepo: &localRepositoryMock{
					localPathFunc: func() string {
						return ""
					},
				},
				DestinationFolder: testFolderName,
				Files:             files,
			}

			err := replacer.renderAllFiles()
			require.NoError(t, err, "renderAllFiles")

			renderedFiles, err := os.ReadDir(replacer.DestinationFolder)
			require.NoError(t, err, "ReadDir")
			for _, file := range renderedFiles {
				b, err := os.ReadFile(path.Join(testFolderName, file.Name()))
				require.NoError(t, err, "reading rendered file")

				require.Equal(t, tc.wantContents[file.Name()], string(b))
			}
		})
	}
}
