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
		ScanInterval: 20 * time.Second,
		Scanners: config.ScannerConfig{
			SSDP: config.ScannerToggle{Enabled: true},
			ARP:  config.ScannerToggle{Enabled: false},
			MDNS: config.ScannerToggle{Enabled: false},
		},
		Sweeper: config.SweeperConfig{Enabled: true, Interval: 5 * time.Minute},
	}

	engine := BuildEngine(iface, nil, cfg)

	if engine == nil {
		t.Errorf("expected engine to be created")
	}
	if engine.Devices == nil {
		t.Errorf("expected Devices channel to be created")
	}
	if engine.Events == nil {
		t.Errorf("expected Events channel to be created")
	}
}

func TestBuildEngineSweeperDisabled(t *testing.T) {
	iface := &discovery.InterfaceInfo{}
	cfg := &config.Config{
		ScanDuration: 10 * time.Second,
		ScanInterval: 20 * time.Second,
		Scanners: config.ScannerConfig{
			SSDP: config.ScannerToggle{Enabled: true},
			ARP:  config.ScannerToggle{Enabled: false},
			MDNS: config.ScannerToggle{Enabled: false},
		},
		Sweeper: config.SweeperConfig{Enabled: false},
	}

	engine := BuildEngine(iface, nil, cfg)

	if engine == nil {
		t.Errorf("expected engine to be created")
	}
}
