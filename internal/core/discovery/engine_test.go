package discovery

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"
)

type fakeScanner struct {
	name    string
	devices []Device
	delay   time.Duration
}

func (f *fakeScanner) Name() string { return f.name }

func (f *fakeScanner) Scan(ctx context.Context, out chan<- Device) error {
	for _, d := range f.devices {
		if f.delay > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(f.delay):
			}
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case out <- d:
		}
	}
	return nil
}

func TestEngineScanOnceMergeAndDedup(t *testing.T) {
	t0 := time.Unix(100, 0)
	t1 := time.Unix(200, 0)
	t2 := time.Unix(300, 0)

	first := Device{
		IP:           net.ParseIP("10.0.0.1"),
		DisplayName:  "one",
		Services:     map[string]int{"svc": 1},
		Sources:      map[string]struct{}{"mdns": {}},
		ExtraData:    map[string]string{"a": "1"},
		Manufacturer: "",
		FirstSeen:    t1,
		LastSeen:     t2,
	}

	second := Device{
		IP:           net.ParseIP("10.0.0.1"),
		MAC:          "aa:bb",
		Manufacturer: "manu",
		Services:     map[string]int{"svc": 0, "svc2": 2},
		Sources:      map[string]struct{}{"ssdp": {}},
		ExtraData:    map[string]string{"b": "2"},
		FirstSeen:    t0,
		LastSeen:     time.Unix(400, 0),
	}

	scanners := []Scanner{
		&fakeScanner{name: "s1", devices: []Device{first}},
		&fakeScanner{name: "s2", devices: []Device{second}},
	}

	eng := NewEngine(scanners, WithTimeout(2*time.Second))

	ctx := context.Background()
	devices, err := eng.ScanOnce(ctx)
	if err != nil {
		t.Fatalf("ScanOnce returned error: %v", err)
	}

	if len(devices) != 1 {
		t.Fatalf("expected 1 device, got %d", len(devices))
	}

	d := devices[0]
	if d.IP.String() != "10.0.0.1" {
		t.Fatalf("unexpected IP: %s", d.IP)
	}
	if d.DisplayName != "one" {
		t.Fatalf("DisplayName not preserved: %s", d.DisplayName)
	}
	if d.MAC != "aa:bb" {
		t.Fatalf("MAC not merged: %s", d.MAC)
	}
	if d.Manufacturer != "manu" {
		t.Fatalf("Manufacturer not merged: %s", d.Manufacturer)
	}
	if d.FirstSeen != t0 {
		t.Fatalf("FirstSeen expected %v got %v", t0, d.FirstSeen)
	}
	if d.LastSeen != time.Unix(400, 0) {
		t.Fatalf("LastSeen not latest: %v", d.LastSeen)
	}
	if d.Services["svc"] != 1 || d.Services["svc2"] != 2 {
		t.Fatalf("services not merged: %+v", d.Services)
	}
	if _, ok := d.Sources["mdns"]; !ok {
		t.Fatalf("mdns source missing")
	}
	if _, ok := d.Sources["ssdp"]; !ok {
		t.Fatalf("ssdp source missing")
	}
	if d.ExtraData["a"] != "1" || d.ExtraData["b"] != "2" {
		t.Fatalf("extra data not merged: %+v", d.ExtraData)
	}
}

func TestEngineTimeoutCancelsScanners(t *testing.T) {
	slow := &fakeScanner{name: "slow", delay: 200 * time.Millisecond, devices: []Device{{IP: net.ParseIP("192.168.1.10")}}}
	eng := NewEngine([]Scanner{slow}, WithTimeout(50*time.Millisecond))

	start := time.Now()
	ctx := context.Background()
	devices, err := eng.ScanOnce(ctx)
	elapsed := time.Since(start)

	if err != nil && !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected context timeout or nil error, got %v", err)
	}
	if len(devices) != 0 {
		t.Fatalf("expected 0 devices due to timeout, got %d", len(devices))
	}
	if elapsed > 300*time.Millisecond {
		t.Fatalf("timeout test took too long: %v", elapsed)
	}
}
