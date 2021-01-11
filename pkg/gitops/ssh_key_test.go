package gitops

import (
	"context"
	"crypto/rsa"
	"io/ioutil"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
)

func TestSSHKey(t *testing.T) {
	ctx := context.Background()
	wantGithubKeyID := time.Now().Second() + 1 // random from 1...60

	// Initialize mock Github client.
	var gotAuthorizedKey string
	var deletedKeyID int
	gh := &githuberMock{
		AddKeyFunc: func(_ context.Context, pub []byte) (int64, error) {
			gotAuthorizedKey = string(pub)
			return int64(wantGithubKeyID), nil
		},
		DeleteKeyFunc: func(_ context.Context, id int64) error {
			deletedKeyID = int(id)
			return nil
		},
	}

	sshKey, err := NewSSHKey(ctx, gh)
	t.Run("create new SSH key", func(t *testing.T) {
		require.NoError(t, err, "newSSHKey")
	})

	t.Run("local private and Github deploy key are a pair", func(t *testing.T) {
		privatePath := sshKey.privateKeyPath()
		gotPrivateKeyBytes, err := ioutil.ReadFile(privatePath)
		require.NoError(t, err, "read contents of %q", privatePath)
		gotPrivateKey, err := ssh.ParseRawPrivateKey(gotPrivateKeyBytes)
		require.NoError(t, err, "ssh.ParseRawPrivateKey")
		gotRSAPrivateKey, ok := gotPrivateKey.(*rsa.PrivateKey)
		require.True(t, ok, "gotPrivateKey.(*rsa.PrivateKey)")

		wantPublicKey := &gotRSAPrivateKey.PublicKey
		gotPublicKey, _, _, _, err := ssh.ParseAuthorizedKey([]byte(gotAuthorizedKey))
		require.NoError(t, err, "ssh.ParseAuthorizedKey")

		msg := "deployed key matches local private key"
		require.EqualValues(t, wantPublicKey, gotPublicKey, msg)
	})

	t.Run("close deletes deploy key from Github", func(t *testing.T) {
		assert.Equal(t, 0, deletedKeyID, "should not be deleted before close")
		sshKey.Close(ctx)
		assert.Equal(t, wantGithubKeyID, deletedKeyID, "deleted key ID")
	})
}
