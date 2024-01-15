package captcha

import "context"

type MockVerifier struct {
	Score float32
}

func (m *MockVerifier) Verify(request *VerifyRequest) (score *Result, err error) {
	return m.VerifyCtx(context.Background(), request)
}

func (m *MockVerifier) VerifyCtx(ctx context.Context, request *VerifyRequest) (score *Result, err error) {
	return &Result{Score: m.Score}, nil
}
