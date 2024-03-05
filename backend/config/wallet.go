package config

import (
	"crypto/ecdsa"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/crypto"
)

type Wallet struct {
	Keystore         string `yaml:"keystore"`
	KeystorePassword string `yaml:"keystore_password"`
	PrivateKey       string `yaml:"private_key"`
}

var (
	ErrMissingKeystoreOrPrivateKey = errors.New("missing wallet keystore or private key")
	ErrMissingKeystorePassword     = errors.New("missing wallet keystore password")
	ErrFailedToReadKeystoreFile    = errors.New("failed to read keystore file")
)

func (w Wallet) ECDSA() (*ecdsa.PrivateKey, error) {
	if w.PrivateKey != "" {
		hex := strings.TrimPrefix(w.PrivateKey, "0x")
		return crypto.HexToECDSA(hex)
	}

	if w.Keystore == "" {
		return nil, ErrMissingKeystoreOrPrivateKey
	}
	if w.KeystorePassword == "" {
		return nil, ErrMissingKeystorePassword
	}

	keystoreJSON := []byte(w.Keystore)
	if _, err := os.Stat(w.Keystore); err == nil {
		b, err := os.ReadFile(w.Keystore)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrFailedToReadKeystoreFile, err)
		}
		keystoreJSON = b
	}

	key, err := keystore.DecryptKey(keystoreJSON, w.KeystorePassword)
	if err != nil {
		return nil, err
	}
	return key.PrivateKey, nil
}
