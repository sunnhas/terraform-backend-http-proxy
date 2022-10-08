package git

import "fmt"

type requestMetadataParams struct {
	Repository, Ref, State string
}

// String is a human-readable representation for this params set
func (params *requestMetadataParams) String() string {
	return fmt.Sprintf("%s?ref=%s//%s", params.Repository, params.Ref, params.State)
}
