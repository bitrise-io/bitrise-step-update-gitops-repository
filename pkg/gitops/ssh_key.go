package gitops

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"golang.org/x/crypto/ssh"
)

//go:generate moq -out ssh_key_moq_test.go . sshKeyer
type sshKeyer interface {
	privateKeyPath() string
	Close(ctx context.Context)
}

// sshKey implements the sshKeyer interface.
var _ sshKeyer = (*sshKey)(nil)

type sshKey struct {
	PrivateKeyFile *os.File

	gh          githuber
	githubKeyID int64
}

// NewSSHKey generates and returns a new SSH key pair.
// It also uploads its public part as a deploy key to Github.
// It should be closed after usage (a repository should close it).
func NewSSHKey(ctx context.Context, gh githuber) (*sshKey, error) {
	// Generate and write private part of RSA key to a temporary file.
	privatekey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("generate private key: %w", err)
	}
	var privateKeyBytes []byte = x509.MarshalPKCS1PrivateKey(privatekey)
	privateKeyBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	}
	tmpPrivateFile, err := ioutil.TempFile("", "")
	if err != nil {
		return nil, fmt.Errorf("create temp private file: %w", err)
	}
	err = pem.Encode(tmpPrivateFile, privateKeyBlock)
	if err != nil {
		return nil, fmt.Errorf("encode private pem: %w", err)
	}

	// Upload public part to Github as deploy key of repository.
	publicKey, err := ssh.NewPublicKey(&privatekey.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("new ssh public key: %w", err)
	}
	keyID, err := gh.AddKey(ctx, ssh.MarshalAuthorizedKey(publicKey))
	if err != nil {
		return nil, fmt.Errorf("add github key: %w", err)
	}

	return &sshKey{
		PrivateKeyFile: tmpPrivateFile,
		gh:             gh,
		githubKeyID:    keyID,
	}, nil
}

func (kp sshKey) privateKeyPath() string {
	return kp.PrivateKeyFile.Name()
}

// Close closes all related resoruces of the SSH key.
// This is a best-effort operation, possible errors are logged as warning,
// not returned as an actual error.
func (kp sshKey) Close(ctx context.Context) {
	// Delete deploy key from Github repository.
	if err := kp.gh.DeleteKey(ctx, kp.githubKeyID); err != nil {
		log.Printf("warning: delete github key: %s\n", err)
	}
	// Delete temporary private key file from the local filesystem.
	path := kp.privateKeyPath()
	if err := os.Remove(path); err != nil {
		log.Printf("warning: remove private key (%q): %s\n", path, err)
	}
}
