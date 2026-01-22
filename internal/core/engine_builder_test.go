package core

import (
	"testing"
	"time"

	"github.com/ramonvermeulen/whosthere/internal/core/config"
	"github.com/ramonvermeulen/whosthere/internal/core/discovery"
	"github.com/ramonvermeulen/whosthere/internal/core/discovery/arp"
	"github.com/ramonvermeulen/whosthere/internal/core/discovery/mdns"
	"github.com/ramonvermeulen/whosthere/internal/core/discovery/ssdp"
)

func TestBuildScanners(t *testing.T) {
	iface := &discovery.InterfaceInfo{}
	enabled := []string{"ssdp", "arp", "mdns"}
	scanners, sweeper := BuildScanners(iface, enabled)

	if len(scanners) != 3 {
		t.Errorf("expected 3 scanners, got %d", len(scanners))
	}
	if sweeper == nil {
		t.Errorf("expected sweeper")
	}

	if _, ok := scanners[0].(*ssdp.Scanner); !ok {
		t.Errorf("expected ssdp scanner")
	}
	if _, ok := scanners[1].(*arp.Scanner); !ok {
		t.Errorf("expected arp scanner")
	}
	if _, ok := scanners[2].(*mdns.Scanner); !ok {
		t.Errorf("expected mdns scanner")
	}
}

func TestGetEnabledFromCfg(t *testing.T) {
	cfg := &config.Config{
		Scanners: config.ScannerConfig{
			SSDP: config.ScannerToggle{Enabled: true},
			ARP:  config.ScannerToggle{Enabled: false},
			MDNS: config.ScannerToggle{Enabled: true},
		},
	}
	enabled := GetEnabledFromCfg(cfg)
	expected := []string{"ssdp", "mdns"}
	if len(enabled) != len(expected) {
		t.Errorf("expected %v, got %v", expected, enabled)
	}
	for i, e := range enabled {
		if e != expected[i] {
			t.Errorf("expected %v, got %v", expected, enabled)
		}
	}
}

func TestBuildEngine(t *testing.T) {
	iface := &discovery.InterfaceInfo{}
	enabled := []string{"ssdp"}
	timeout := 10 * time.Second
	engine := BuildEngine(iface, nil, enabled, timeout)

	if len(engine.Scanners) != 1 {
		t.Errorf("expected 1 scanner")
	}
	if engine.Timeout != timeout {
		t.Errorf("expected timeout %v, got %v", timeout, engine.Timeout)
	}
}
