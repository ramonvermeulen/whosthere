package discovery

import (
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestDeviceMerge(t *testing.T) {
	base := NewDevice(net.ParseIP("10.0.0.1"))
	base.SetDisplayName("host")
	base.AddSource("a")
	base.AddExtraData("k1", "v1")
	base.SetFirstSeen(time.Unix(100, 0))
	base.SetLastSeen(time.Unix(200, 0))

	other := NewDevice(net.ParseIP("10.0.0.1"))
	other.SetMAC("aa:bb")
	other.AddSource("b")
	other.AddExtraData("k2", "v2")
	other.SetFirstSeen(time.Unix(50, 0))
	other.SetLastSeen(time.Unix(300, 0))

	base.Merge(other)

	require.Equal(t, "aa:bb", base.MAC(), "expected MAC merged")
	require.Equal(t, "host", base.DisplayName(), "DisplayName should remain original when non-empty")
	require.Empty(t, base.Manufacturer(), "Manufacturer merge failed")
	sources := base.Sources()
	require.Contains(t, sources, "a", "source a missing")
	require.Contains(t, sources, "b", "source b missing")
	extra := base.ExtraData()
	require.Equal(t, "v1", extra["k1"], "extra data k1")
	require.Equal(t, "v2", extra["k2"], "extra data k2")
	require.True(t, base.FirstSeen().Equal(time.Unix(50, 0)), "FirstSeen should be earliest, got %v", base.FirstSeen())
	require.True(t, base.LastSeen().Equal(time.Unix(300, 0)), "LastSeen should be latest, got %v", base.LastSeen())
}

func TestDeviceMergeNilOther(t *testing.T) {
	d := NewDevice(net.IP{})
	d.Merge(nil)
}
