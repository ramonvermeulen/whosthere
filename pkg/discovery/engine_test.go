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

func TestNewEngine_RequiresScannerOrSweeper(t *testing.T) {
	e, err := discovery.NewEngine(discovery.WithInterface(testkit.MustInterfaceInfo(t)))
	require.Nil(t, e)
	require.ErrorIs(t, err, discovery.ErrNoScannersOrSweeper)
}

func TestNewEngine_RequiresInterface(t *testing.T) {
	s := &testkit.FakeScanner{Devices: []*discovery.Device{discovery.NewDevice(testkit.MustIP(t, "10.0.0.1"))}}
	e, err := discovery.NewEngine(discovery.WithScanners(s))
	require.Nil(t, e)
	require.ErrorIs(t, err, discovery.ErrNoInterface)
}

func TestEngine_Scan_MergesDevices(t *testing.T) {
	iface := testkit.MustInterfaceInfo(t)

	base := discovery.NewDevice(testkit.MustIP(t, "10.0.0.2"))
	base.SetDisplayName("host")
	base.AddSource("a")

	other := discovery.NewDevice(testkit.MustIP(t, "10.0.0.2"))
	other.SetMAC("aa:bb:cc:dd:ee:ff")
	other.AddSource("b")

	s1 := &testkit.FakeScanner{NameStr: "s1", Devices: []*discovery.Device{base}}
	s2 := &testkit.FakeScanner{NameStr: "s2", Devices: []*discovery.Device{other}}

	e, err := discovery.NewEngine(
		discovery.WithInterface(iface),
		discovery.WithScanners(s1, s2),
		discovery.WithScanTimeout(250*time.Millisecond),
		discovery.WithScanInterval(time.Millisecond),
	)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	devices, scanErr := e.Scan(ctx)
	require.NoError(t, scanErr)
	require.Len(t, devices, 1)

	d := devices[0]
	require.Equal(t, "10.0.0.2", d.IP().String())
	require.Equal(t, "host", d.DisplayName())
	require.Equal(t, "aa:bb:cc:dd:ee:ff", d.MAC())
}

func TestEngine_Scan_IgnoresInvalidDevice(t *testing.T) {
	iface := testkit.MustInterfaceInfo(t)
	s := &testkit.FakeScanner{Devices: []*discovery.Device{{}}}

	e, err := discovery.NewEngine(
		discovery.WithInterface(iface),
		discovery.WithScanners(s),
		discovery.WithScanTimeout(100*time.Millisecond),
	)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	devices, scanErr := e.Scan(ctx)
	require.NoError(t, scanErr)
	require.Empty(t, devices)
}

func TestEngine_Scan_EmitsDiscoveredEventForEachObservation(t *testing.T) {
	iface := testkit.MustInterfaceInfo(t)

	d1 := discovery.NewDevice(testkit.MustIP(t, "10.0.0.2"))
	d2 := discovery.NewDevice(testkit.MustIP(t, "10.0.0.2"))
	d2.SetMAC("aa:bb:cc:dd:ee:ff")

	s := &testkit.FakeScanner{NameStr: "s", Devices: []*discovery.Device{d1, d2}}

	e, err := discovery.NewEngine(
		discovery.WithInterface(iface),
		discovery.WithScanners(s),
		discovery.WithScanTimeout(250*time.Millisecond),
	)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	events := e.Start(ctx)
	defer e.Stop()

	var discovered int
	for ev := range events {
		if ev.Type == discovery.EventDeviceDiscovered {
			discovered++
			if ev.Device != nil {
				require.Equal(t, "10.0.0.2", ev.Device.IP().String())
			}
		}
		if ev.Type == discovery.EventScanCompleted {
			break
		}
	}

	require.Equal(t, 2, discovered)
}

func TestEngine_StartStop_ClosesEvents(t *testing.T) {
	iface := testkit.MustInterfaceInfo(t)
	s := &testkit.FakeScanner{Devices: []*discovery.Device{discovery.NewDevice(net.IPv4(10, 0, 0, 1))}}
	e, err := discovery.NewEngine(
		discovery.WithInterface(iface),
		discovery.WithScanners(s),
		discovery.WithScanInterval(time.Millisecond),
		discovery.WithScanTimeout(100*time.Millisecond),
	)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	ch := e.Start(ctx)
	e.Stop()

	for {
		_, ok := <-ch
		if !ok {
			break
		}
	}
}
