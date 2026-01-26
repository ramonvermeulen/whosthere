package core

import (
	"testing"
	"time"

	"github.com/ramonvermeulen/whosthere/internal/core/config"
	"github.com/ramonvermeulen/whosthere/internal/core/discovery"
)

func TestBuildEngine(t *testing.T) {
	iface := &discovery.InterfaceInfo{}
	cfg := &config.Config{
		ScanDuration: 10 * time.Second,
		Scanners: config.ScannerConfig{
			SSDP: config.ScannerToggle{Enabled: true},
			ARP:  config.ScannerToggle{Enabled: false},
			MDNS: config.ScannerToggle{Enabled: false},
		},
		Sweeper: config.SweeperConfig{Enabled: true, Interval: 5 * time.Minute},
	}

	engine := BuildEngine(iface, nil, cfg)

	if len(engine.Scanners) != 1 {
		t.Errorf("expected 1 scanner, got %d", len(engine.Scanners))
	}
	if engine.Timeout != 10*time.Second {
		t.Errorf("expected timeout 10s, got %v", engine.Timeout)
	}
	if engine.Sweeper == nil {
		t.Errorf("expected sweeper to be created")
	}
}

func TestBuildEngineSweeperDisabled(t *testing.T) {
	iface := &discovery.InterfaceInfo{}
	cfg := &config.Config{
		ScanDuration: 10 * time.Second,
		Scanners: config.ScannerConfig{
			SSDP: config.ScannerToggle{Enabled: true},
			ARP:  config.ScannerToggle{Enabled: false},
			MDNS: config.ScannerToggle{Enabled: false},
		},
		Sweeper: config.SweeperConfig{Enabled: false},
	}

	engine := BuildEngine(iface, nil, cfg)

	if engine.Sweeper != nil {
		t.Errorf("expected sweeper to be nil when disabled")
	}
}
