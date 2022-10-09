package storage

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"terraform-backend-http-proxy/storage/git"
	"terraform-backend-http-proxy/storage/internal"
	"terraform-backend-http-proxy/storage/storagetypes"
)

// knownStorageTypes map storage types to storage clients before starting the server so backend knows what's supported
var knownStorageTypes = make(map[string]Client)

func init() {
	
}

// GetStorageClient gets the storage client based on the client
func GetStorageClient(data storagetypes.ClientData) (Client, error) {
	if client, ok := knownStorageTypes[data.Type]; ok {
		return client, nil
	}

	return nil, fmt.Errorf("unknown storage type %s", data.Type)
}

type Client interface {
	CreateParams(params *gin.Context) storage.ClientTypeMetadata
	GetLockData(data storage.ClientTypeMetadata) (*storagetypes.LockInfo, error)
	LockState(storage.ClientTypeMetadata, []byte) error
	UnlockState(storage.ClientTypeMetadata) error
	GetState(storage.ClientTypeMetadata) ([]byte, error)
	UpdateState(storage.ClientTypeMetadata, []byte) error
}
