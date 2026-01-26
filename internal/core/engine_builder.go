package core

import (
	"time"

	"github.com/ramonvermeulen/whosthere/internal/core/config"
	"github.com/ramonvermeulen/whosthere/internal/core/discovery"
	"github.com/ramonvermeulen/whosthere/internal/core/discovery/arp"
	"github.com/ramonvermeulen/whosthere/internal/core/discovery/mdns"
	"github.com/ramonvermeulen/whosthere/internal/core/discovery/ssdp"
	"github.com/ramonvermeulen/whosthere/internal/core/oui"
)

func BuildScanners(iface *discovery.InterfaceInfo, enabled []string) ([]discovery.Scanner, *arp.Sweeper) {
	sweeper := arp.NewSweeper(iface, 5*time.Minute, time.Minute)
	var scanners []discovery.Scanner
	for _, e := range enabled {
		switch e {
		case "ssdp":
			scanners = append(scanners, ssdp.NewScanner(iface))
		case "arp":
			scanners = append(scanners, arp.NewScanner(iface, sweeper))
		case "mdns":
			scanners = append(scanners, mdns.NewScanner(iface))
		}
	}
	return scanners, sweeper
}

func GetEnabledFromCfg(cfg *config.Config) []string {
	var enabled []string
	if cfg.Scanners.SSDP.Enabled {
		enabled = append(enabled, "ssdp")
	}
	if cfg.Scanners.ARP.Enabled {
		enabled = append(enabled, "arp")
	}
	if cfg.Scanners.MDNS.Enabled {
		enabled = append(enabled, "mdns")
	}
	return enabled
}

func BuildEngine(iface *discovery.InterfaceInfo, ouiDB *oui.Registry, enabled []string, timeout time.Duration) *discovery.Engine {
	scanners, _ := BuildScanners(iface, enabled)

	if ouiDB != nil {
		return discovery.NewEngine(scanners, discovery.WithTimeout(timeout), discovery.WithOUIRegistry(ouiDB))
	}

	return discovery.NewEngine(scanners, discovery.WithTimeout(timeout))
}
