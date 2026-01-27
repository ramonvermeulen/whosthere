package core

import (
	"github.com/ramonvermeulen/whosthere/internal/core/config"
	"github.com/ramonvermeulen/whosthere/internal/core/discovery"
	"github.com/ramonvermeulen/whosthere/internal/core/discovery/arp"
	"github.com/ramonvermeulen/whosthere/internal/core/discovery/mdns"
	"github.com/ramonvermeulen/whosthere/internal/core/discovery/ssdp"
	"github.com/ramonvermeulen/whosthere/internal/core/oui"
)

func BuildEngine(iface *discovery.InterfaceInfo, ouiDB *oui.Registry, cfg *config.Config) *discovery.Engine {
	var scanners []discovery.Scanner

	if cfg.Scanners.SSDP.Enabled {
		scanners = append(scanners, ssdp.NewScanner(iface))
	}
	if cfg.Scanners.ARP.Enabled {
		scanners = append(scanners, arp.NewScanner(iface))
	}
	if cfg.Scanners.MDNS.Enabled {
		scanners = append(scanners, mdns.NewScanner(iface))
	}

	opts := []discovery.EngineOption{discovery.WithTimeout(cfg.ScanDuration)}
	if ouiDB != nil {
		opts = append(opts, discovery.WithOUIRegistry(ouiDB))
	}

	engine := discovery.NewEngine(scanners, opts...)

	if cfg.Sweeper.Enabled {
		engine.Sweeper = discovery.NewSweeper(iface, cfg.Sweeper.Interval)
	}

	return engine
}
