// The breaker package implements multiple types of circuit breakers.
package breaker

import (
	"sync"
	"time"

	"github.com/if-itb/siasn-libs-backend/logutil"
)

// A circuit breaker based on error rate.
// A circuit breaker signals the program using Tripped channel when the breaker error count
// has reached the preconfigured limit. You can add new error for the breaker to track using AddError.
// The breaker count will reset back to 0 after no errors have been received after a certain amount of time.
type RateCircuitBreaker struct {
	// Error counter limit until the breaker is tripped.
	Limit int
	// The reset cooldown. See also AddError.
	Cooldown time.Duration
	isActive bool
	timer    *time.Timer
	// Cancellation signal channel. Value is added to it in Deactivate.
	cancel chan struct{}
	// Counter mutex, preventing race condition with reset go
	// routine in resetTimer, Current and AddError, also making AddError thread safe.
	countMu sync.RWMutex
	// Timer mutex, preventing stopping a nil timer (race condition between resetTimer and Deactivate).
	timerMu sync.Mutex
	errs    []error
	Logger  logutil.Logger

	// Closed when the breaker is tripped.
	Tripped chan struct{}
}

// Activate starts the breaker.
// Subsequent calls to AddError will now trigger the breaker.
func (b *RateCircuitBreaker) Activate() {
	b.isActive = true
	b.cancel = make(chan struct{})
	b.Tripped = make(chan struct{})
}

func (b *RateCircuitBreaker) IsActive() bool {
	return b.isActive
}

// Get the current breaker counter.
// Current is thread safe to use.
func (b *RateCircuitBreaker) Current() int {
	b.countMu.RLock()
	defer b.countMu.RUnlock()
	return len(b.errs)
}

func (b *RateCircuitBreaker) resetTimer() {
	b.timerMu.Lock()
	defer b.timerMu.Unlock()
	if b.timer == nil {
		b.timer = time.NewTimer(b.Cooldown)
		go func() {
			select {
			case <-b.timer.C:
				logutil.Debug("breaker error count has been reset")
			case <-b.cancel:
			}

			b.countMu.Lock()
			defer b.countMu.Unlock()

			b.errs = nil
			b.timer = nil
		}()
	} else {
		b.timer.Reset(b.Cooldown)
	}
}

// AddErrors adds a new error to the breaker, adding the breaker error count and resetting the breaker internal timer.
// During this time, if more errors are added before the cooldown is reset, the breaker limit will be
// reached and the breaker will be tripped.
//
// If the breaker is not active, AddError won't do anything.
//
// AddError is thread safe to use.
func (b *RateCircuitBreaker) AddError(err error) {
	if !b.IsActive() {
		return
	}

	b.countMu.Lock()

	b.errs = append(b.errs, err)
	b.resetTimer()

	if len(b.errs) >= b.Limit {
		b.countMu.Unlock()
		close(b.Tripped)
		logutil.Warnf("breaker tripped, logged errors: %v", b.errs)
		b.Deactivate()
		return
	}

	b.countMu.Unlock()
}

// Deactivate deactivates the timer.
// This also stops the breaker internal timer and resets its count to 0.
func (b *RateCircuitBreaker) Deactivate() {
	if !b.isActive {
		return
	}

	b.isActive = false

	b.timerMu.Lock()
	defer b.timerMu.Unlock()
	if b.timer != nil {
		b.timer.Stop()
		b.cancel <- struct{}{}
	}
}
