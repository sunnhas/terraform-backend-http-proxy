package sops

import (
	"fmt"
	sp "go.mozilla.org/sops/v3"
	"go.mozilla.org/sops/v3/aes"
	"go.mozilla.org/sops/v3/cmd/sops/common"
	"go.mozilla.org/sops/v3/keyservice"
	"go.mozilla.org/sops/v3/stores/json"
	"go.mozilla.org/sops/v3/version"
	"os"
	"strconv"
)

type EncryptionProvider struct{}

// Encrypt will encrypt the data in buffer and return encrypted result.
func (p *EncryptionProvider) Encrypt(data []byte) ([]byte, error) {
	keyGroups, err := getActivatedKeyGroups()
	if err != nil {
		return nil, err
	}

	inputStore := &json.Store{}
	branches, err := inputStore.LoadPlainFile(data)
	if err != nil {
		return nil, err
	}

	tree := sp.Tree{
		Branches: branches,
		Metadata: sp.Metadata{
			KeyGroups: keyGroups,
			Version:   version.Version,
		},
	}

	if shamirThreshold, ok := os.LookupEnv("TF_BACKEND_HTTP_SOPS_SHAMIR_THRESHOLD"); ok {
		st, err := strconv.Atoi(shamirThreshold)
		if err != nil {
			return nil, err
		}
		tree.Metadata.ShamirThreshold = st
	}

	dataKey, errs := tree.GenerateDataKeyWithKeyServices([]keyservice.KeyServiceClient{keyservice.NewLocalClient()})
	if len(errs) > 0 {
		return nil, fmt.Errorf("Could not generate data key: %s", errs)
	}

	if err := common.EncryptTree(common.EncryptTreeOpts{
		DataKey: dataKey,
		Tree:    &tree,
		Cipher:  aes.NewCipher(),
	}); err != nil {
		return nil, err
	}

	outputStore := &json.Store{}
	return outputStore.EmitEncryptedFile(tree)
}

// Decrypt will decrypt the data in buffer.
func (p *EncryptionProvider) Decrypt(data []byte) ([]byte, error) {
	inputStore := &json.Store{}
	tree, _ := inputStore.LoadEncryptedFile(data)

	if tree.Metadata.Version == "" {
		return data, nil
	}

	if _, err := common.DecryptTree(common.DecryptTreeOpts{
		Cipher:      aes.NewCipher(),
		Tree:        &tree,
		KeyServices: []keyservice.KeyServiceClient{keyservice.NewLocalClient()},
	}); err != nil {
		return nil, err
	}

	outputStore := &json.Store{}
	return outputStore.EmitPlainFile(tree.Branches)
}

type keyConfig interface {
	isActivated() bool
	keyGroup() (sp.KeyGroup, error)
}

var keyConfigs = make(map[string]keyConfig)

func getActivatedKeyGroups() ([]sp.KeyGroup, error) {
	keyGroups := make([]sp.KeyGroup, 0)

	for _, config := range keyConfigs {
		if config.isActivated() {
			kg, err := config.keyGroup()
			if err != nil {
				return nil, err
			}
			keyGroups = append(keyGroups, kg)
		}
	}

	return keyGroups, nil
}
