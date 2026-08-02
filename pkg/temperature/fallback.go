package temperature

import (
	"fmt"
	"os"
	"sync"
	"time"
)

// primaryRetryInterval is how long to stay on the fallback before trying the
// primary again. Long enough that a slow, failing Prometheus does not add its
// timeout to every control cycle; short enough that recovery is unattended.
const primaryRetryInterval = 30 * time.Second

// FallbackSource reads from a primary source and drops to a secondary one only
// while the primary is failing.
//
// It exists because the previous behaviour decided once, at startup: if
// Prometheus was unreachable when the daemon started, it fell back to the local
// thermal file permanently and never retried. On a cluster where the local file
// is the coolest board in the chassis, that means silently regulating on the
// wrong sensor forever, with no error to notice — a boot ordering accident
// turning into a lasting misconfiguration.
type FallbackSource struct {
	primary   Source
	secondary Source

	mu        sync.Mutex
	degraded  bool
	nextRetry time.Time
	now       func() time.Time // injectable for tests
}

// NewFallbackSource wraps primary with a secondary used only during failures.
func NewFallbackSource(primary, secondary Source) *FallbackSource {
	return &FallbackSource{primary: primary, secondary: secondary, now: time.Now}
}

// GetTemperature returns the primary reading when it can, else the secondary.
func (f *FallbackSource) GetTemperature() (float64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if !f.degraded || !f.now().Before(f.nextRetry) {
		temp, err := f.primary.GetTemperature()
		if err == nil {
			if f.degraded {
				fmt.Fprintf(os.Stderr, "Primary temperature source recovered\n")
				f.degraded = false
			}
			return temp, nil
		}

		if !f.degraded {
			fmt.Fprintf(os.Stderr,
				"Primary temperature source failed, using fallback (retrying every %s): %v\n",
				primaryRetryInterval, err)
			f.degraded = true
		}
		f.nextRetry = f.now().Add(primaryRetryInterval)
	}

	return f.secondary.GetTemperature()
}

// Close closes both sources, reporting the first error.
func (f *FallbackSource) Close() error {
	err := f.primary.Close()
	if secErr := f.secondary.Close(); err == nil {
		err = secErr
	}
	return err
}
