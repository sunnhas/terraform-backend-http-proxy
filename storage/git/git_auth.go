package git

import (
	"errors"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"os"
	"strings"
)

// auth determines authentication method and discovers Git credentials in the environment
func auth(params *requestMetadataParams) (transport.AuthMethod, error) {
	if strings.HasPrefix(params.Repository, "http") {
		auth, err := authBasicHTTP()
		if err != nil {
			return nil, err
		}

		return auth, nil
	}

	return nil, errors.New("only http is supported right now")
}

func authBasicHTTP() (*http.BasicAuth, error) {
	username, okUsername := os.LookupEnv("GIT_USERNAME")
	if !okUsername {
		return nil, errors.New("git protocol was http but username was not set")
	}

	password, okPassword := os.LookupEnv("GIT_PASSWORD")
	if !okPassword {
		ghToken, okGhToken := os.LookupEnv("GITHUB_TOKEN")
		if !okGhToken {
			return nil, errors.New("git protocol was http but neither password nor token was set")
		}
		password = ghToken
	}

	return &http.BasicAuth{
		Username: username,
		Password: password,
	}, nil
}
