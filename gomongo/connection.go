package gomongo

import (
	"errors"
	"fmt"
	"time"
)

var (
	ErrInvalidSettings = errors.New("settings must be valid")
)

// ConnectionSettings is a struct that holds the connection settings.
type ConnectionSettings struct {
	URI          string // Uri is the connection string to the database.
	DatabaseName string // DatabaseName is the database name.

	ConnectionTimeout time.Duration // ConnectionTimeout is the timeout used for creating connections to the server. If it is negative, no timeout will be used. The default is 30 seconds.
}

func (cs *ConnectionSettings) validate() error {
	if cs.URI == "" {
		return fmt.Errorf("%w: URI can not be empty", ErrInvalidSettings)
	}

	if cs.DatabaseName == "" {
		return fmt.Errorf("%w: Database Name can not be empty", ErrInvalidSettings)
	}

	return nil
}
