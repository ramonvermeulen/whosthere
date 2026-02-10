package core

import (
	"context"
	"log/slog"

	"github.com/ramonvermeulen/whosthere/internal/core/config"
	"github.com/ramonvermeulen/whosthere/internal/core/paths"
	discovery2 "github.com/ramonvermeulen/whosthere/pkg/discovery"
	"github.com/ramonvermeulen/whosthere/pkg/discovery/oui"
	"github.com/ramonvermeulen/whosthere/pkg/discovery/scanners/arp"
	"github.com/ramonvermeulen/whosthere/pkg/discovery/scanners/mdns"
	"github.com/ramonvermeulen/whosthere/pkg/discovery/scanners/ssdp"
	sweeper2 "github.com/ramonvermeulen/whosthere/pkg/discovery/sweeper"
)

func BuildEngine(cfg *config.Config, logger discovery2.Logger) (*discovery2.Engine, error) {
	ctx := context.Background()

	stateDir, err := paths.StateDir()
	if err != nil {
		logger.Log(ctx, slog.LevelWarn, "failed to resolve state dir for OUI cache; continuing with embedded OUI", "error", err)
		stateDir = ""
	}

	ouiDB, err := oui.New(ctx, oui.WithCacheDir(stateDir))
	if err != nil {
		logger.Log(ctx, slog.LevelWarn, "failed to initialize OUI DB; continuing without OUI", "error", err)
		ouiDB = nil
	}

	iface, err := discovery2.NewInterfaceInfo(cfg.NetworkInterface)
	if err != nil {
		return nil, err
	}

	var scanners []discovery2.Scanner

	if cfg.Scanners.SSDP.Enabled {
		scanners = append(scanners, ssdp.New(iface, ssdp.WithLogger(logger)))
	}
	if cfg.Scanners.ARP.Enabled {
		s, err := arp.New(iface, arp.WithLogger(logger))
		if err != nil {
			return nil, err
		}
		scanners = append(scanners, s)
	}
	if cfg.Scanners.MDNS.Enabled {
		s, err := mdns.New(iface, mdns.WithLogger(logger))
		if err != nil {
			return nil, err
		}
		scanners = append(scanners, s)
	}

	opts := []discovery2.Option{
		discovery2.WithInterface(iface),
		discovery2.WithScanners(scanners...),
		discovery2.WithScanTimeout(cfg.ScanTimeout),
		discovery2.WithScanInterval(cfg.ScanInterval),
		discovery2.WithLogger(logger),
	}

	if ouiDB != nil {
		opts = append(opts, discovery2.WithOUIRegistry(ouiDB))
	}

	if cfg.Sweeper.Enabled {
		sweeperOpts := []sweeper2.Option{
			sweeper2.WithSweeperInterface(iface),
			sweeper2.WithSweeperInterval(cfg.Sweeper.Interval),
			sweeper2.WithSweeperTimeout(cfg.Sweeper.Timeout),
			sweeper2.WithSweeperLogger(logger),
		}
		s, _ := sweeper2.New(sweeperOpts...)
		opts = append(opts, discovery2.WithSweeper(s))
	}

	return discovery2.NewEngine(opts...)
}
