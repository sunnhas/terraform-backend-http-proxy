package storagetypes

import (
	"terraform-backend-http-proxy/storage/internal"
	"time"
)

// ClientData is the data required for all type of clients
// regardless of the actual implementation.
type ClientData struct {
	// Type is the type of storage to be used (implementation type).
	Type string

	// ID is the id of the specific storage to use.
	// For Git this would be the repository url
	ID string

	// Metadata is storage specific request metadata
	Metadata storage.ClientTypeMetadata
}

// LockInfo represents a TF Lock Metadata.
// https://github.com/hashicorp/terraform/blob/v1.3.2/internal/states/statemgr/locker.go#L115-L138.
type LockInfo struct {
	// Unique ID for the lock. NewLockInfo provides a random ID, but this may
	// be overridden by the lock implementation. The final value of ID will be
	// returned by the call to Lock.
	ID string

	// Terraform operation, provided by the caller.
	Operation string

	// Extra information to store with the lock, provided by the caller.
	Info string

	// user@hostname when available
	Who string

	// Terraform version
	Version string

	// Time that the lock was taken.
	Created time.Time

	// Path to the state file when applicable. Set by the Lock implementation.
	Path string
}
