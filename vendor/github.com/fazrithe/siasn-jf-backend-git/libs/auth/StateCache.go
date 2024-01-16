package auth

import (
	"errors"
	"net/http"
)

type StateCache interface {
	// CreateStateFromRequest reads the HTTP request to login and create a state data from it that can be stored with
	// SaveState. The returned item is typically in the form of StateData.
	CreateStateFromRequest(request *http.Request) (data *StateData, err error)

	// SaveState saves state and additional data in cache.
	// State ID will be generated.
	SaveState(data *StateData) (stateKey string, err error)

	// VerifyState retrieves state from cache.
	// If it does not exist it will throw ErrStateInvalid error. Just like nonce, it cannot be verified twice.
	// Data will be decoded into item, and so item should be a pointer and cannot be nil.
	VerifyState(stateKey string, data *StateData) (err error)
}

type StateData struct {
	ReturnUrl string
	Extra     map[string]interface{}
}

var ErrStateInvalid = errors.New("state is invalid and cannot be used")
