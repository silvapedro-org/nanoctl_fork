package temperature

import (
	"errors"
	"testing"
	"time"
)

type stubSource struct {
	temp  float64
	err   error
	reads int
}

func (s *stubSource) GetTemperature() (float64, error) {
	s.reads++
	return s.temp, s.err
}

func (s *stubSource) Close() error { return nil }

func TestFallbackUsesPrimaryWhenHealthy(t *testing.T) {
	primary := &stubSource{temp: 70}
	secondary := &stubSource{temp: 40}
	f := NewFallbackSource(primary, secondary)

	got, err := f.GetTemperature()
	if err != nil || got != 70 {
		t.Fatalf("got %v, %v; want 70, nil", got, err)
	}
	if secondary.reads != 0 {
		t.Errorf("secondary was read %d times while primary was healthy", secondary.reads)
	}
}

func TestFallbackRecoversInsteadOfLatching(t *testing.T) {
	primary := &stubSource{err: errors.New("prometheus down")}
	secondary := &stubSource{temp: 40}
	f := NewFallbackSource(primary, secondary)

	now := time.Now()
	f.now = func() time.Time { return now }

	// Primary is down: we fall back.
	if got, err := f.GetTemperature(); err != nil || got != 40 {
		t.Fatalf("degraded read: got %v, %v; want 40, nil", got, err)
	}

	// Within the retry window the primary is not hammered.
	readsAfterFirst := primary.reads
	if got, _ := f.GetTemperature(); got != 40 {
		t.Fatalf("expected to stay on the fallback, got %v", got)
	}
	if primary.reads != readsAfterFirst {
		t.Errorf("primary retried inside the cooldown: %d -> %d", readsAfterFirst, primary.reads)
	}

	// Once it recovers and the window passes, we must go back to it — this is
	// the whole point: the old code latched to the fallback forever.
	primary.err = nil
	primary.temp = 70
	now = now.Add(primaryRetryInterval + time.Second)

	if got, err := f.GetTemperature(); err != nil || got != 70 {
		t.Fatalf("after recovery: got %v, %v; want 70, nil", got, err)
	}

	// And stays there without consulting the fallback again.
	readsBefore := secondary.reads
	if got, _ := f.GetTemperature(); got != 70 {
		t.Fatalf("expected primary, got %v", got)
	}
	if secondary.reads != readsBefore {
		t.Errorf("secondary consulted after recovery")
	}
}
