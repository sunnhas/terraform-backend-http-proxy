package git

import (
	"errors"
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/storage"
	"github.com/go-git/go-git/v5/storage/memory"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// gitSession represents a particular Git repository
type gitSession struct {
	// auth credentials for remote operations
	auth transport.AuthMethod

	// storer used for local working tree config
	storer storage.Storer

	// fs can be used to access local working tree
	fs billy.Filesystem

	// repository represents a git repository
	repository *git.Repository

	// mutex since we can't be doing parallel complex operations on a single working tree, involving checkout branches etc.,
	// we need to use the lock and make sure only one tread is "connected" (interacts with the repository using local working tree).
	mutex sync.Mutex
}

// newStorageSession makes a fresh clone to in-memory FS and saves everything to the StorageSession
func newStorageSession(params *requestMetadataParams) (*gitSession, error) {
	storageSession := &gitSession{
		storer: memory.NewStorage(),
		fs:     memfs.New(),
		mutex:  sync.Mutex{},
	}

	if err := storageSession.clone(params); err != nil {
		return nil, err
	}

	return storageSession, nil
}

// clone remote repository
func (gitSession *gitSession) clone(params *requestMetadataParams) error {
	auth, err := auth(params)
	if err != nil {
		return err
	}

	gitSession.auth = auth

	refer := ref(params.Ref, false)
	cloneOptions := &git.CloneOptions{
		URL:           params.Repository,
		Auth:          auth,
		ReferenceName: refer,
	}

	repository, err := git.Clone(gitSession.storer, gitSession.fs, cloneOptions)
	if err != nil {
		return err
	}

	gitSession.repository = repository

	return nil
}

// checkoutMode configures checkout behaviour
type checkoutMode uint8

const (
	// checkoutModeDefault is default checkout mode - no special behaviour
	checkoutModeDefault checkoutMode = 1 << iota
	// checkoutModeCreate will indicate that the new local branch needs to be created at checkout
	checkoutModeCreate
	// checkoutModeRemote will indicate that the remote branch needs to be checked out
	checkoutModeRemote
)

// checkout this repository working copy to specified branch.
// If create flag was true, it will make an attempt to create a new branch, and it will return an error if it already existed.
func (gitSession *gitSession) checkout(branch string, mode checkoutMode) error {
	if mode&checkoutModeCreate != 0 && mode&checkoutModeRemote != 0 {
		return errors.New("checkoutModeCreate and checkoutModeRemote cannot be used simultaniously")
	}

	tree, err := gitSession.repository.Worktree()
	if err != nil {
		return err
	}

	checkoutOptions := &git.CheckoutOptions{
		Branch: ref(branch, mode&checkoutModeRemote != 0),
		Force:  true,
		Create: mode&checkoutModeCreate != 0,
	}

	if err := tree.Checkout(checkoutOptions); err != nil {
		return err
	}

	return nil
}

// Attempt to pull from remote to the current branch.
// This branch must already exist locally and upstream must be set for it to know where to pull from.
// It will ignore git.NoErrAlreadyUpToDate.
func (gitSession *gitSession) pull(branch string) error {
	tree, err := gitSession.repository.Worktree()
	if err != nil {
		return err
	}

	pullOptions := git.PullOptions{
		ReferenceName: ref(branch, false),
		Force:         true,
		Auth:          gitSession.auth,
	}

	if err := tree.Pull(&pullOptions); err != nil && err != git.NoErrAlreadyUpToDate {
		return err
	}

	return nil
}

// Attempt to fetch from remote for specified ref specs.
// It will ignore git.NoErrAlreadyUpToDate.
func (gitSession *gitSession) fetch(refs []config.RefSpec) error {
	fetchOptions := git.FetchOptions{
		RefSpecs: refs,
		Auth:     gitSession.auth,
	}

	remote, err := gitSession.getRemote()
	if err != nil {
		return err
	}

	if err := remote.Fetch(&fetchOptions); err != nil && err != git.NoErrAlreadyUpToDate {
		return err
	}

	return nil
}

// Will delete the branch locally.
// Additionally delete branch remotely if deleteRemote was set true.
// Operation is idempotent, i.e. no error will be returned if the branch did not exist.
func (gitSession *gitSession) deleteBranch(branch string, deleteRemote bool) error {
	ref := ref(branch, false)

	if err := gitSession.repository.Storer.RemoveReference(ref); err != nil {
		return err
	}

	if !deleteRemote {
		return nil
	}

	remote, err := gitSession.getRemote()
	if err != nil {
		return err
	}

	pushOptions := &git.PushOptions{
		RefSpecs: []config.RefSpec{
			config.RefSpec(":" + ref),
		},
		Auth: gitSession.auth,
	}

	if err := remote.Push(pushOptions); err != nil && err != git.NoErrAlreadyUpToDate {
		return err
	}

	return nil
}

// add path to the local working tree
func (gitSession *gitSession) add(path string) error {
	tree, err := gitSession.repository.Worktree()
	if err != nil {
		return err
	}

	if _, err := tree.Add(path); err != nil {
		return err
	}

	return nil
}

// commit currently staged changes to the local working tree
func (gitSession *gitSession) commit(msg string) error {
	user, err := gitSession.getUserDetails()
	if err != nil {
		return err
	}

	tree, err := gitSession.repository.Worktree()
	if err != nil {
		return err
	}

	commitOptions := git.CommitOptions{
		Author: &object.Signature{
			Name:  user.name,
			Email: user.email,
			When:  time.Now(),
		},
	}

	if _, err := tree.Commit(msg, &commitOptions); err != nil {
		return err
	}

	return nil
}

// push current working tree state to the remote repository
// It assumes the upstream has been set for the current branch - it will not do anything to define the ref.
func (gitSession *gitSession) push() error {
	remote, err := gitSession.getRemote()
	if err != nil {
		return err
	}

	pushOptions := git.PushOptions{
		Auth: gitSession.auth,
	}

	if err := remote.Push(&pushOptions); err != nil {
		return err
	}

	return nil
}

type userDetails struct {
	name, email string
}

func (gitSession *gitSession) getUserDetails() (*userDetails, error) {
	name, err := gitExecute("config", "user.name")
	if err != nil {
		return nil, err
	}

	email, err := gitExecute("config", "user.email")
	if err != nil {
		return nil, err
	}

	return &userDetails{
		name:  name,
		email: email,
	}, nil
}

// ref convert short branch name string to a full ReferenceName
func ref(branch string, remote bool) plumbing.ReferenceName {
	var ref string
	if remote {
		ref = "refs/remotes/origin/"
	} else {
		ref = "refs/heads/"
	}

	return plumbing.ReferenceName(ref + branch)
}

func gitExecute(arg ...string) (string, error) {
	cmd := exec.Command("git", arg...)

	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}

// getRemote returns "origin" remote.
// Since we never specified a name for our remote, it should always be origin.
func (gitSession *gitSession) getRemote() (*git.Remote, error) {
	remote, err := gitSession.repository.Remote("origin")
	if err != nil {
		return nil, err
	}

	return remote, nil
}
