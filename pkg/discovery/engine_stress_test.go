//go:build stress

package discovery_test

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/ramonvermeulen/whosthere/pkg/discovery"
	"github.com/ramonvermeulen/whosthere/pkg/discovery/internal/testkit"
	"github.com/stretchr/testify/require"
)

func TestEngine_HighVolume_Race(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}

	iface := testkit.MustInterfaceInfo(t)

	n := 100_000
	devices := make([]discovery.Device, 0, n)
	for i := 0; i < n; i++ {
		ip := net.IPv4(10, byte((i>>16)+1), byte(i>>8), byte(i))
		d := discovery.NewDevice(ip)
		devices = append(devices, d)
	}

	s := &testkit.FakeScanner{NameStr: "hi", Devices: devices, DeviceBurst: 128}
	e, err := discovery.NewEngine(
		discovery.WithInterface(iface),
		discovery.WithScanners(s),
		discovery.WithScanTimeout(2*time.Second),
		discovery.WithScanInterval(time.Millisecond),
	)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	events := e.Run(ctx)
	defer e.Stop()

	for ev := range events {
		if ev.Type == discovery.EventScanCompleted {
			require.NotNil(t, ev.Stats)
			require.Equal(t, n, ev.Stats.DeviceCount)
			return
		}
	}

	t.Fatal("engine stopped before scan completed")
}
