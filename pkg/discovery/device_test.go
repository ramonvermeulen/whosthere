package discovery

import (
	"net"
	"testing"
	"time"
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

	if base.MAC() != "aa:bb" {
		t.Fatalf("expected MAC merged, got %s", base.MAC())
	}
	if base.DisplayName() != "host" {
		t.Fatalf("DisplayName should remain original when non-empty, got %s", base.DisplayName())
	}
	if base.Manufacturer() != "" {
		t.Fatalf("Manufacturer merge failed, got %s", base.Manufacturer())
	}
	sources := base.Sources()
	if _, ok := sources["a"]; !ok {
		t.Fatalf("source a missing")
	}
	if _, ok := sources["b"]; !ok {
		t.Fatalf("source b missing")
	}
	extra := base.ExtraData()
	if extra["k1"] != "v1" || extra["k2"] != "v2" {
		t.Fatalf("extra data merge failed: %+v", extra)
	}
	if !base.FirstSeen().Equal(time.Unix(50, 0)) {
		t.Fatalf("FirstSeen should be earliest, got %v", base.FirstSeen())
	}
	if !base.LastSeen().Equal(time.Unix(300, 0)) {
		t.Fatalf("LastSeen should be latest, got %v", base.LastSeen())
	}
}

func TestDeviceMergeNilOther(t *testing.T) {
	d := NewDevice(net.IP{})
	d.Merge(nil)
}
