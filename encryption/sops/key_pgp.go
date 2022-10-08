package sops

import (
	"go.mozilla.org/sops/v3"
	"go.mozilla.org/sops/v3/pgp"
	"os"
)

func init() {
	keyConfigs["pgp"] = &pgpKeys{}
}

type pgpKeys struct{}

func (c *pgpKeys) isActivated() bool {
	_, ok := os.LookupEnv("TF_BACKEND_HTTP_SOPS_PGP_FP")
	return ok
}

func (c *pgpKeys) keyGroup() (sops.KeyGroup, error) {
	fp := os.Getenv("TF_BACKEND_HTTP_SOPS_PGP_FP")

	var keyGroup sops.KeyGroup

	for _, k := range pgp.MasterKeysFromFingerprintString(fp) {
		keyGroup = append(keyGroup, k)
	}

	return keyGroup, nil
}
