package auth

import "errors"

// NonceCache is a cache established to store and check nonce sent with OIDC authorization code flow.
type NonceCache interface {
	// GenerateNonce generates a nonce.
	GenerateNonce() (nonce string, err error)

	// VerifyNonce verifies the nonce.
	// After verifying, the same nonce won't be able to be verified again (cannot be used again).
	// ErrNonceInvalid will be returned for invalid nonce.
	VerifyNonce(nonce string) (err error)
}

var ErrNonceInvalid = errors.New("nonce is invalid and cannot be used")
