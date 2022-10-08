package sops

import (
	"go.mozilla.org/sops/v3"
	"go.mozilla.org/sops/v3/age"
	"os"
)

func init() {
	keyConfigs["age"] = &ageKeys{}
}

type ageKeys struct{}

func (c *ageKeys) isActivated() bool {
	_, ok := os.LookupEnv("TF_BACKEND_HTTP_SOPS_AGE_FP")
	return ok
}

func (c *ageKeys) keyGroup() (sops.KeyGroup, error) {
	recipients := os.Getenv("TF_BACKEND_HTTP_SOPS_AGE_FP")

	var keyGroup sops.KeyGroup

	masterKeys, err := age.MasterKeysFromRecipients(recipients)
	if err != nil {
		return nil, err
	}

	for _, k := range masterKeys {
		keyGroup = append(keyGroup, k)
	}

	return keyGroup, nil
}
