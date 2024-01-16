package breaker_test

import (
	"errors"
	"github.com/if-itb/siasn-libs-backend/breaker"
	"github.com/if-itb/siasn-libs-backend/logutil"
	"testing"
	"time"
)

func TestBreaker_AddError(t *testing.T) {
	b := &breaker.RateCircuitBreaker{
		Limit:    3,
		Cooldown: 3 * time.Second,
		Logger:   logutil.NewStdLogger(false, "breaker"),
	}

	b.Activate()
	b.AddError(errors.New("test"))
	if b.Current() != 1 {
		t.Fatal("current counter is not 1 after error is added once")
	}

	time.Sleep(4 * time.Second)
	if b.Current() != 0 {
		t.Fatal("current counter is not reset after 4 seconds")
	}

	b.AddError(errors.New("test"))
	time.Sleep(4 * time.Second)
	b.AddError(errors.New("test"))
	b.AddError(errors.New("test"))
}

func TestBreaker_AddErrorTripped(t *testing.T) {
	b := &breaker.RateCircuitBreaker{
		Limit:    3,
		Cooldown: 3 * time.Second,
		Logger:   logutil.NewStdLogger(false, "breaker"),
	}

	b.Activate()
	b.AddError(errors.New("test"))
	b.AddError(errors.New("test"))
	b.AddError(errors.New("test"))
	select {
	case <-b.Tripped:
	default:
		t.Fatal("breaker is not tripped")
	}
}

func TestTestBreaker_Deactivate(t *testing.T) {
	b := &breaker.RateCircuitBreaker{
		Limit:    3,
		Cooldown: 3 * time.Second,
		Logger:   logutil.NewStdLogger(false, "breaker"),
	}

	b.Activate()
	b.AddError(errors.New("test"))
	if b.Current() != 1 {
		t.Fatal("current counter is not 1 after error is added once")
	}

	b.Deactivate()
	time.Sleep(1 * time.Second)
	if b.Current() != 0 {
		t.Fatal("current counter is not 0 after breaker is deactivated")
	}
}
