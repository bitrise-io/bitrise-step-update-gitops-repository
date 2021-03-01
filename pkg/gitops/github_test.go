package gitops

import (
	"testing"

	"github.com/bitrise-io/go-steputils/stepconf"
	"github.com/stretchr/testify/require"
)

var newGithubRepoCases = map[string]struct {
	url, user string
	token     stepconf.Secret
	want      *githubRepo
	wantErr   bool
}{
	"simple https url for github": {
		url:   "https://github.com/szabolcsgelencser/sample-deploy-config.git",
		user:  "my-bot-user",
		token: "my-bot-users-pat",
		want: &githubRepo{
			url:   "https://my-bot-user:my-bot-users-pat@github.com/szabolcsgelencser/sample-deploy-config.git",
			token: "my-bot-users-pat",
			owner: "szabolcsgelencser",
			name:  "sample-deploy-config",
		},
	},
	"another simple https url for github": {
		url:   "https://github.com/bitrise-io/den.git",
		user:  "my-other-user",
		token: "my-other-users-pat",
		want: &githubRepo{
			url:   "https://my-other-user:my-other-users-pat@github.com/bitrise-io/den.git",
			token: "my-other-users-pat",
			owner: "bitrise-io",
			name:  "den",
		},
	},
	"unsupported ssh url for github": {
		url:     "git@github.com:bitrise-io/den.git",
		wantErr: true,
	},
	"malformed url (missing prefix)": {
		url:     "bitrise-io/den.git",
		wantErr: true,
	},
	"malformed https url (missing postfix)": {
		url:     "https://github.com/bitrise-io/den",
		wantErr: true,
	},
	"malformed https url (not having owner/repo)": {
		url:     "https://github.com/den.git",
		wantErr: true,
	},
}

func TestNewGithubRepo(t *testing.T) {
	for name, tc := range newGithubRepoCases {
		t.Run(name, func(t *testing.T) {
			got, gotErr := NewGithubRepo(tc.url, tc.user, tc.token)
			if tc.wantErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
			require.Equal(t, tc.want, got)
		})
	}
}
