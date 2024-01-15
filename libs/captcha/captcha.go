// Package captcha provides an interface and implementation of captcha verification.
// All verifications will return verification result. The most important thing to see
// in verification result is score. The score should be above your desired limit to be
// valid. Score always range from 0 to 1.
//
// Some captcha implementation like hCaptcha (excluding hCaptcha enterprise) may not support scoring. Successful captcha will
// then always have score of 1, and failure will always result in score of 0.
package captcha

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/netip"
	"strings"
)

var (
	// ErrTokenNotFound is returned when the given token is empty.
	ErrTokenNotFound = errors.New("token is empty")
)

type VerifyRequest struct {
	// Token is mandatory. It is generated from the frontend.
	Token string
	// Expected site key.
	SiteKey string
	// Optional, the user remote IP address.
	RemoteAddress netip.Addr
	// Optional, the User-Agent in the request.
	UserAgent string
	// Optional, only supported in some Captcha providers.
	ExpectedAction string
}

// Result is a verification result.
type Result struct {
	// The score should be above your desired limit to be
	// valid. Score always range from 0 to 1.
	Score float32
}

// Verifier verify captcha request and return a score.
type Verifier interface {
	Verify(request *VerifyRequest) (score *Result, err error)
	VerifyCtx(ctx context.Context, request *VerifyRequest) (score *Result, err error)
}

// GetRemoteAddressFromRequest reads HTTP request and take client IP address from it.
// If X-Forwarded-For or Forwarded-For header exists, IP address will be taken from that instead.
func GetRemoteAddressFromRequest(request *http.Request) (netip.Addr, error) {
	xForwardedFor := request.Header.Get("X-Forwarded-For")
	forwardedFor := request.Header.Get("Forwarded-For")
	xForwardedFors := strings.Split(xForwardedFor, ", ")
	forwardedFors := strings.Split(forwardedFor, ", ")

	if forwardedFor != "" && len(forwardedFors) > 0 {
		parsed, err := netip.ParseAddr(forwardedFors[0])
		if err != nil {
			return netip.Addr{}, errors.New(fmt.Sprintf("cannot parse IP address from %s", forwardedFor))
		}

		return parsed, nil
	}

	if xForwardedFor != "" && len(xForwardedFors) > 0 {
		parsed, err := netip.ParseAddr(xForwardedFors[0])
		if err != nil {
			return netip.Addr{}, errors.New(fmt.Sprintf("cannot parse IP address from %s", xForwardedFor))
		}

		return parsed, nil
	}

	parsed, err := netip.ParseAddrPort(request.RemoteAddr)
	if err != nil {
		return netip.Addr{}, errors.New(fmt.Sprintf("cannot parse IP address and port from %s", request.RemoteAddr))
	}

	return parsed.Addr(), nil
}
