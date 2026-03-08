package ports

import "errors"

var (
	// ErrNotLoggedIn is returned when authentication is required but not present.
	ErrNotLoggedIn = errors.New("not logged in")
	// ErrNoActiveSession is returned when no chat session is active.
	ErrNoActiveSession = errors.New("no active chat session")
)
