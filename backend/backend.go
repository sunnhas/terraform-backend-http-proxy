package backend

import (
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"terraform-backend-http-proxy/storage"
	"terraform-backend-http-proxy/storage/storagetypes"
)

// Errors
var (
	// StateIsLocked indicates that the state is already locked
	// and can't be locked again.
	StateIsLocked = errors.New("state is locked")

	// NotLockedByMe indicates that the state is locked by
	// another person and that any procedures should not be done.
	NotLockedByMe = errors.New("state is not locked by me")
)

// ParseRequestData is parsing request data to the requests
// client type.
func ParseRequestData(params *gin.Context) (*storagetypes.ClientData, error) {
	requestData := storagetypes.ClientData{
		ID:   params.Query("ID"),
		Type: params.Query("type"),
	}

	storageClient, err := storage.GetStorageClient(requestData)
	if err != nil {
		return nil, err
	}

	requestData.Metadata = storageClient.CreateParams(params)

	return &requestData, nil
}

// LockState is trying to lock the current Terraform state.
// Returning StateIsLocked will also return the given
// lock info for the already acquired lock.
func LockState(requestData *storagetypes.ClientData, rawLockData []byte) (*storagetypes.LockInfo, error) {
	storageClient, err := storage.GetStorageClient(*requestData)
	if err != nil {
		return nil, err
	}

	lockData, err := storageClient.GetLockData(requestData.Metadata)

	// Having no lock is perfect 🙃
	if err != nil && !errors.Is(err, storagetypes.ErrLockMissing) {
		return nil, err
	}

	// State is already locked, we can't proceed
	if lockData != nil {
		return lockData, StateIsLocked
	}

	if err := storageClient.LockState(requestData.Metadata, rawLockData); err != nil {
		return nil, err
	}

	return nil, nil
}

// UnlockState is unlocking the Terraform state.
// It's returning an error if it fails.
// It's a requirement that the lock is acquired by the one trying to unlock.
func UnlockState(requestData *storagetypes.ClientData, rawLockData []byte) error {
	// Force unlock the Terraform state has the lock ID set in the params
	// whereas the regular unlock hasn't. We need to get that from the
	// body instead.
	force := requestData.ID != ""
	if !force {
		var lock storagetypes.LockInfo
		if err := json.Unmarshal(rawLockData, &lock); err != nil {
			return err
		}

		requestData.ID = lock.ID
	}

	storageClient, err := storage.GetStorageClient(*requestData)
	if err != nil {
		return err
	}

	// We can only unlock the state we have acquired our self
	if err := lockedByMe(requestData, storageClient); err != nil {
		return err
	}

	if err := storageClient.UnlockState(requestData.Metadata); err != nil {
		return err
	}

	return nil
}

// GetState will get the raw json state from the storage client.
func GetState(requestData *storagetypes.ClientData) ([]byte, error) {
	storageClient, err := storage.GetStorageClient(*requestData)
	if err != nil {
		return nil, err
	}

	state, err := storageClient.GetState(requestData.Metadata)
	if err != nil {
		return nil, err
	}

	return state, nil
}

// UpdateState updates the raw json state in the storage client.
func UpdateState(requestData *storagetypes.ClientData, body []byte) error {
	storageClient, err := storage.GetStorageClient(*requestData)
	if err != nil {
		return err
	}

	// We can only update state if we obtained the lock
	if err := lockedByMe(requestData, storageClient); err != nil {
		return err
	}

	if err := storageClient.UpdateState(requestData.Metadata, body); err != nil {
		return err
	}

	return nil
}

func DeleteState() {
	panic(errors.New("not implemented"))
}

func lockedByMe(data *storagetypes.ClientData, client storage.Client) error {
	lockInfo, err := client.GetLockData(data.Metadata)
	if err != nil {
		return err
	}

	if data.ID != lockInfo.ID {
		return NotLockedByMe
	}

	return nil
}
