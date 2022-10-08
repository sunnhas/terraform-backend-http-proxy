package encryption

import (
	"fmt"
	"os"
	"terraform-backend-http-proxy/encryption/sops"
)

// EncryptionProvider is the provider for any encryption
// that can happen on the state content.
type EncryptionProvider interface {
	Encrypt([]byte) ([]byte, error)
	Decrypt([]byte) ([]byte, error)
}

var encryptionProviders = make(map[string]EncryptionProvider)

func init() {
	encryptionProviders["sops"] = &sops.EncryptionProvider{}
}

// GetEncryptionProvider get the encryption provider based
// on the environment variables set before launching the tool.
func GetEncryptionProvider() (EncryptionProvider, error) {
	provider, enabled := os.LookupEnv("TF_BACKEND_HTTP_ENCRYPTION_PROVIDER")
	if enabled {
		if p, ok := encryptionProviders[provider]; ok {
			return p, nil
		}

		return nil, fmt.Errorf("unknown encryption provider %q", provider)
	}

	return nil, nil
}
