package session

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/google/uuid"
)

var (
	sessionID    string
	sessionMutex sync.Mutex
)

// SetSession sets the session ID. If sessionID is empty, it generates a new one.
func SetSession(id string) error {
	sessionMutex.Lock()
	defer sessionMutex.Unlock()

	if id == "" {
		// Generate a new session ID if not provided
		newSessionID, err := uuid.NewUUID()
		if err != nil {
			return err
		}
		sessionID = newSessionID.String()
		sessionComponents := strings.Split(sessionID, "-")
		if len(sessionComponents) == 5 {
			sessionID = sessionComponents[4]  // This is the node part
		}
		fmt.Println("Generated new session ID:", sessionID)
	} else {
		sessionID = id
	}
	return nil
}

// GetSession returns the session ID.
func GetSession() (string, error) {
	sessionMutex.Lock()
	defer sessionMutex.Unlock()

	if sessionID == "" {
		return "", errors.New("session ID is not set")
	}
	return sessionID, nil
}
