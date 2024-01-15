package captcha

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

type hCaptchaResponse struct {
	Success     bool     `json:"success"`
	ChallengeTs string   `json:"challenge_ts"`
	Hostname    string   `json:"hostname"`
	Credit      bool     `json:"credit"`
	ErrorCodes  []string `json:"error-codes"`
	Score       float32  `json:"score"`
	ScoreReason []string `json:"score_reason"`
}

// HCaptchaVerifier is a HCaptcha verifier.
// HCaptcha does not support expected action, and unless we are using hCaptcha enterprise, it does not support
// scoring too.
type HCaptchaVerifier struct {
	HttpClient *http.Client
	VerifyUrl  string
	Secret     string
}

func NewHCaptchaVerifier(secret string) *HCaptchaVerifier {
	return &HCaptchaVerifier{HttpClient: &http.Client{}, VerifyUrl: "https://hcaptcha.com/siteverify", Secret: secret}
}

func (h *HCaptchaVerifier) Verify(request *VerifyRequest) (score *Result, err error) {
	return h.VerifyCtx(context.Background(), request)
}

func (h *HCaptchaVerifier) VerifyCtx(ctx context.Context, request *VerifyRequest) (score *Result, err error) {
	if request.Token == "" {
		return nil, ErrTokenNotFound
	}

	q := url.Values{}
	q.Set("response", request.Token)
	q.Set("secret", h.Secret)
	if request.SiteKey != "" {
		q.Set("sitekey", request.SiteKey)
	}
	if request.RemoteAddress.IsValid() {
		q.Set("remoteip", request.RemoteAddress.String())
	}
	req, _ := http.NewRequestWithContext(ctx, "POST", h.VerifyUrl, bytes.NewBufferString(q.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := h.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}

	raw, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("cannot read captcha response: %w", err)
	}

	respPayload := &hCaptchaResponse{}
	err = json.Unmarshal(raw, respPayload)
	if err != nil {
		return nil, fmt.Errorf("cannot read captcha response: %w", err)
	}

	if respPayload.Success {
		s := float32(1)
		if respPayload.Score > 0 {
			s = respPayload.Score
		}
		return &Result{Score: s}, nil
	}

	return nil, fmt.Errorf("cannot verify captcha: %v", respPayload.ErrorCodes)
}
