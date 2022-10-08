package git

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"sync"
	"terraform-backend-http-proxy/storage/internal"
	"terraform-backend-http-proxy/storage/storagetypes"
)

// StorageClient implementation for Git storage type
type StorageClient struct {
	// sessions key is repository URL, value is everything we need to interact with it
	sessions map[string]*gitSession

	// sessionsMutex used for locking sessions map for adding new repositories
	sessionsMutex sync.Mutex
}

// NewStorageClient creates new StorageClient
func NewStorageClient() *StorageClient {
	return &StorageClient{
		sessions:      make(map[string]*gitSession),
		sessionsMutex: sync.Mutex{},
	}
}

func (client *StorageClient) CreateParams(params *gin.Context) storage.ClientTypeMetadata {
	return &requestMetadataParams{
		Repository: params.Query("repository"),
		Ref:        params.Query("ref"),
		State:      params.Query("state"),
	}
}

func (client *StorageClient) GetLockData(data storage.ClientTypeMetadata) (*storagetypes.LockInfo, error) {
	params := data.(*requestMetadataParams)

	session, err := client.getSession(params)
	if err != nil {
		return nil, err
	}

	session.mutex.Lock()
	defer session.mutex.Unlock()

	if err := session.fetch(locksRefSpecs); err != nil {
		return nil, err
	}

	lockBranchName := getLockBranchName(params)

	// Delete any local leftovers from the past
	if err := session.deleteBranch(lockBranchName, false); err != nil {
		return nil, err
	}

	if err := session.checkout(lockBranchName, checkoutModeRemote); err != nil {
		if err == plumbing.ErrReferenceNotFound {
			return nil, storagetypes.ErrLockMissing
		}
		return nil, err
	}

	if err := session.pull(lockBranchName); err != nil {
		return nil, err
	}

	lock, err := session.readFile(getLockPath(params))
	if err != nil {
		return nil, err
	}

	var lockInfo storagetypes.LockInfo
	if err := json.Unmarshal(lock, &lockInfo); err != nil {
		return nil, err
	}

	return &lockInfo, nil
}

func (client *StorageClient) LockState(data storage.ClientTypeMetadata, rawLockData []byte) error {
	params := data.(*requestMetadataParams)

	session, err := client.getSession(params)
	if err != nil {
		return err
	}

	session.mutex.Lock()
	defer session.mutex.Unlock()

	if err := session.checkout(params.Ref, checkoutModeDefault); err != nil {
		return err
	}

	lockBranchName := getLockBranchName(params)

	// Delete any local leftovers from the past
	if err := session.deleteBranch(lockBranchName, false); err != nil {
		return err
	}

	// Create local branch to start preparing a new lock metadata for push
	if err := session.checkout(lockBranchName, checkoutModeCreate); err != nil {
		return err
	}

	lockPath := getLockPath(params)

	if err := session.writeFile(lockPath, rawLockData); err != nil {
		return err
	}

	if err := session.add(lockPath); err != nil {
		return err
	}

	if err := session.commit("Lock " + params.State); err != nil {
		return err
	}

	if err := session.push(); err != nil {
		return err
	}

	return nil
}

func (client *StorageClient) UnlockState(data storage.ClientTypeMetadata) error {
	params := data.(*requestMetadataParams)

	session, err := client.getSession(params)
	if err != nil {
		return err
	}

	if err := session.deleteBranch(getLockBranchName(params), true); err != nil {
		return err
	}

	return nil
}

func (client *StorageClient) GetState(data storage.ClientTypeMetadata) ([]byte, error) {
	params := data.(*requestMetadataParams)

	session, err := client.getSession(params)
	if err != nil {
		return nil, err
	}

	session.mutex.Lock()
	defer session.mutex.Unlock()

	if err := session.checkout(params.Ref, checkoutModeDefault); err != nil {
		return nil, err
	}

	if err := session.pull(params.Ref); err != nil {
		return nil, err
	}

	s, err := session.readFile(params.State)
	if err != nil {
		return s, err
	}

	return s, nil
}

func (client *StorageClient) UpdateState(data storage.ClientTypeMetadata, state []byte) error {
	params := data.(*requestMetadataParams)

	session, err := client.getSession(params)
	if err != nil {
		return err
	}

	session.mutex.Lock()
	defer session.mutex.Unlock()

	if err := session.checkout(params.Ref, checkoutModeDefault); err != nil {
		return err
	}

	if err := session.pull(params.Ref); err != nil {
		return err
	}

	if err := session.writeFile(params.State, state); err != nil {
		return err
	}

	if err := session.add(params.State); err != nil {
		return err
	}

	if err := session.commit("Update " + params.State); err != nil {
		return err
	}

	if err := session.push(); err != nil {
		return err
	}

	return nil
}

func (client *StorageClient) getSession(data *requestMetadataParams) (*gitSession, error) {
	client.sessionsMutex.Lock()
	defer client.sessionsMutex.Unlock()

	session, ok := client.sessions[data.Repository]
	if !ok {
		s, err := newStorageSession(data)
		if err != nil {
			return nil, err
		}

		client.sessions[data.Repository] = s
		session = s
	}

	return session, nil
}

func getLockPath(params *requestMetadataParams) string {
	return params.State + ".lock"
}

func getLockBranchName(params *requestMetadataParams) string {
	return "lock/" + params.State
}

var (
	locksRefSpecs = []config.RefSpec{
		"refs/heads/locks/*:refs/remotes/origin/locks/*",
	}
)
