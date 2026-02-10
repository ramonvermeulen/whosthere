package discovery_test

import (
	"context"
	"testing"
	"time"

	"github.com/ramonvermeulen/whosthere/pkg/discovery"
	"github.com/ramonvermeulen/whosthere/pkg/discovery/internal/testkit"
	"github.com/stretchr/testify/require"
)

func TestEngine_StartStop_CleansUpAndNoSendOnClosed(t *testing.T) {
	iface := testkit.MustInterfaceInfo(t)

	s1 := &testkit.FakeScanner{NameStr: "s1", Devices: []*discovery.Device{discovery.NewDevice(testkit.MustIP(t, "10.0.0.10"))}}
	s2 := &testkit.FakeScanner{NameStr: "s2", Devices: []*discovery.Device{discovery.NewDevice(testkit.MustIP(t, "10.0.0.11"))}}

	e, err := discovery.NewEngine(
		discovery.WithInterface(iface),
		discovery.WithScanners(s1, s2),
		discovery.WithScanTimeout(200*time.Millisecond),
		discovery.WithScanInterval(time.Millisecond),
	)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	events := e.Start(ctx)

	require.Eventually(t, func() bool {
		for {
			select {
			case ev, ok := <-events:
				if !ok {
					return false
				}
				if ev.Type == discovery.EventScanCompleted {
					return true
				}
			default:
				return false
			}
		}
	}, time.Second, 5*time.Millisecond)

	e.Stop()
	cancel()

	_, ok := <-events
	if ok {
		for range events {
		}
	}
}
