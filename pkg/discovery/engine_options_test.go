package discovery_test

import (
	"testing"
	"time"

	"github.com/ramonvermeulen/whosthere/pkg/discovery"
	"github.com/ramonvermeulen/whosthere/pkg/discovery/internal/testkit"
	"github.com/stretchr/testify/require"
)

func TestWithScanInterval_AllowsZero(t *testing.T) {
	s := &testkit.FakeScanner{Devices: []*discovery.Device{discovery.NewDevice(testkit.MustIP(t, "10.0.0.1"))}}
	e, err := discovery.NewEngine(
		discovery.WithInterface(testkit.MustInterfaceInfo(t)),
		discovery.WithScanners(s),
		discovery.WithScanTimeout(50*time.Millisecond),
		discovery.WithScanInterval(0),
	)
	require.NoError(t, err)
	require.NotNil(t, e)
}

func TestWithScanInterval_RejectsNegative(t *testing.T) {
	s := &testkit.FakeScanner{Devices: []*discovery.Device{discovery.NewDevice(testkit.MustIP(t, "10.0.0.1"))}}
	e, err := discovery.NewEngine(
		discovery.WithInterface(testkit.MustInterfaceInfo(t)),
		discovery.WithScanners(s),
		discovery.WithScanTimeout(50*time.Millisecond),
		discovery.WithScanInterval(-1*time.Millisecond),
	)
	require.Error(t, err)
	require.Nil(t, e)
}
