package discovery_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/ramonvermeulen/whosthere/pkg/discovery"
	"github.com/ramonvermeulen/whosthere/pkg/discovery/internal/testkit"
	"github.com/stretchr/testify/require"
)

type deterministicScanner struct {
	scanDuration time.Duration

	mu     sync.Mutex
	starts []time.Time
}

func (s *deterministicScanner) Name() string { return "timing" }

func (s *deterministicScanner) Scan(ctx context.Context, out chan<- *discovery.Device) error {
	_ = out

	s.mu.Lock()
	s.starts = append(s.starts, time.Now())
	s.mu.Unlock()

	t := time.NewTimer(s.scanDuration)
	defer t.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

func (s *deterministicScanner) StartTimes() []time.Time {
	s.mu.Lock()
	defer s.mu.Unlock()

	res := make([]time.Time, len(s.starts))
	copy(res, s.starts)
	return res
}

func waitForScanCompletions(events <-chan discovery.Event, n int, deadline time.Duration) int {
	timer := time.NewTimer(deadline)
	defer timer.Stop()

	var completed int
	for completed < n {
		select {
		case <-timer.C:
			return completed
		case ev, ok := <-events:
			if !ok {
				return completed
			}
			if ev.Type == discovery.EventScanCompleted {
				completed++
			}
		}
	}
	return completed
}

func assertDurationNear(t *testing.T, got, want, tolerance time.Duration) {
	t.Helper()

	require.InDelta(t, float64(want), float64(got), float64(tolerance), "got=%s want=%s tol=%s", got, want, tolerance)
}

func gaps(starts []time.Time) []time.Duration {
	if len(starts) < 2 {
		return nil
	}
	res := make([]time.Duration, 0, len(starts)-1)
	for i := 1; i < len(starts); i++ {
		res = append(res, starts[i].Sub(starts[i-1]))
	}
	return res
}

func TestEngine_Start_WhenScanDelayLessThanScanInterval_SchedulesFromScanStart(t *testing.T) {
	iface := testkit.MustInterfaceInfo(t)

	scanDuration := 80 * time.Millisecond  // how long the fake scanner blocks in Scan() (simulated scan runtime)
	scanInterval := 150 * time.Millisecond // how often the engine should start new scans (measured from scan start)
	scanTimeout := 200 * time.Millisecond  // per-scan context deadline enforced by the engine

	s := &deterministicScanner{scanDuration: scanDuration}

	e, err := discovery.NewEngine(
		discovery.WithInterface(iface),
		discovery.WithScanners(s),
		discovery.WithScanTimeout(scanTimeout),
		discovery.WithScanInterval(scanInterval),
	)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 800*time.Millisecond)
	defer cancel()

	events := e.Start(ctx)
	defer e.Stop()

	require.GreaterOrEqual(t, waitForScanCompletions(events, 3, 700*time.Millisecond), 3)

	starts := s.StartTimes()
	require.GreaterOrEqual(t, len(starts), 3)

	g := gaps(starts)
	require.GreaterOrEqual(t, len(g), 2)

	assertDurationNear(t, g[0], scanInterval, 40*time.Millisecond)
	assertDurationNear(t, g[1], scanInterval, 40*time.Millisecond)
}

func TestEngine_Start_WhenScanTimeoutLongerThanScanInterval_StartsNextScanImmediately(t *testing.T) {
	iface := testkit.MustInterfaceInfo(t)

	scanDuration := 160 * time.Millisecond
	scanInterval := 100 * time.Millisecond
	scanTimeout := 300 * time.Millisecond

	s := &deterministicScanner{scanDuration: scanDuration}

	e, err := discovery.NewEngine(
		discovery.WithInterface(iface),
		discovery.WithScanners(s),
		discovery.WithScanTimeout(scanTimeout),
		discovery.WithScanInterval(scanInterval),
	)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 800*time.Millisecond)
	defer cancel()

	events := e.Start(ctx)
	defer e.Stop()

	require.GreaterOrEqual(t, waitForScanCompletions(events, 2, 700*time.Millisecond), 2)

	starts := s.StartTimes()
	require.GreaterOrEqual(t, len(starts), 2)

	gap := starts[1].Sub(starts[0])
	require.Less(t, gap, scanDuration+60*time.Millisecond)
}

func TestEngine_Start_WhenScanCompletesBeforeScanInterval_WaitsRemainingInterval(t *testing.T) {
	iface := testkit.MustInterfaceInfo(t)

	scanDuration := 30 * time.Millisecond
	scanInterval := 150 * time.Millisecond
	scanTimeout := 200 * time.Millisecond

	s := &deterministicScanner{scanDuration: scanDuration}

	e, err := discovery.NewEngine(
		discovery.WithInterface(iface),
		discovery.WithScanners(s),
		discovery.WithScanTimeout(scanTimeout),
		discovery.WithScanInterval(scanInterval),
	)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 800*time.Millisecond)
	defer cancel()

	events := e.Start(ctx)
	defer e.Stop()

	require.GreaterOrEqual(t, waitForScanCompletions(events, 3, 700*time.Millisecond), 3)

	starts := s.StartTimes()
	require.GreaterOrEqual(t, len(starts), 2)

	gap := starts[1].Sub(starts[0])
	require.Greater(t, gap, scanInterval-40*time.Millisecond)
	require.Less(t, gap, scanInterval+60*time.Millisecond)
}

func TestEngine_Start_WhenScanIntervalIsZero_PerformsSingleScanOnly(t *testing.T) {
	iface := testkit.MustInterfaceInfo(t)

	s := &deterministicScanner{scanDuration: 10 * time.Millisecond}

	e, err := discovery.NewEngine(
		discovery.WithInterface(iface),
		discovery.WithScanners(s),
		discovery.WithScanTimeout(100*time.Millisecond),
		discovery.WithScanInterval(0),
	)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 400*time.Millisecond)
	defer cancel()

	events := e.Start(ctx)
	defer e.Stop()

	got := waitForScanCompletions(events, 1, 300*time.Millisecond)
	require.Equal(t, 1, got)
}
