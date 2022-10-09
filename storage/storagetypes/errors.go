package storagetypes

import "errors"

var (
	// ErrLockMissing indicate that the lock didn't exist when it was expected/required to
	ErrLockMissing = errors.New("was not locked")
)
