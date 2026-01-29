package core

import (
	"context"
	"log/slog"

	"github.com/ramonvermeulen/whosthere/internal/core/config"
	"github.com/ramonvermeulen/whosthere/internal/core/discovery"
	"github.com/ramonvermeulen/whosthere/internal/core/discovery/oui"
	"github.com/ramonvermeulen/whosthere/internal/core/discovery/scanners/arp"
	"github.com/ramonvermeulen/whosthere/internal/core/discovery/scanners/mdns"
	"github.com/ramonvermeulen/whosthere/internal/core/discovery/scanners/ssdp"
	"github.com/ramonvermeulen/whosthere/internal/core/discovery/sweeper"
)

func BuildEngine(cfg *config.Config, logger discovery.Logger) (*discovery.Engine, error) {
	ctx := context.Background()

	ouiDB, err := oui.Init(ctx)
	if err != nil {
		logger.Log(context.Background(), slog.LevelWarn, "failed to initialize OUI DB; continuing without OUI", "error", err)
		ouiDB = nil
	}

	iface, err := discovery.NewInterfaceInfo(cfg.NetworkInterface)
	if err != nil {
		return nil, err
	}

	var scanners []discovery.Scanner

	if cfg.Scanners.SSDP.Enabled {
		scanners = append(scanners, ssdp.NewScanner(iface, logger))
	}
	if cfg.Scanners.ARP.Enabled {
		scanners = append(scanners, arp.NewScanner(iface, logger))
	}
	if cfg.Scanners.MDNS.Enabled {
		scanners = append(scanners, mdns.NewScanner(iface, logger))
	}

	opts := []discovery.Option{
		discovery.WithInterface(iface),
		discovery.WithScanners(scanners...),
		discovery.WithScanTimeout(cfg.ScanTimeout),
		discovery.WithScanInterval(cfg.ScanInterval),
		discovery.WithLogger(logger),
	}

	if ouiDB != nil {
		opts = append(opts, discovery.WithOUIRegistry(ouiDB))
	}

	if cfg.Sweeper.Enabled {
		sweeperOpts := []sweeper.Option{
			sweeper.WithSweeperInterface(iface),
			sweeper.WithSweeperInterval(cfg.Sweeper.Interval),
			sweeper.WithSweeperTimeout(cfg.Sweeper.Timeout),
			sweeper.WithSweeperLogger(logger),
		}
		s, _ := sweeper.NewSweeper(sweeperOpts...)
		opts = append(opts, discovery.WithSweeper(s))
	}

	return discovery.New(opts...)
}
