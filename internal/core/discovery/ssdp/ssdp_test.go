package ssdp

import (
	"testing"

	"github.com/ramonvermeulen/whosthere/internal/core/discovery"
)

func TestNewScanner(t *testing.T) {
	iface := &discovery.InterfaceInfo{}
	scanner := NewScanner(iface)
	if scanner.iface != iface {
		t.Errorf("expected iface to be set")
	}
}

func TestName(t *testing.T) {
	scanner := NewScanner(nil)
	if scanner.Name() != "ssdp" {
		t.Errorf("expected name ssdp, got %s", scanner.Name())
	}
}
