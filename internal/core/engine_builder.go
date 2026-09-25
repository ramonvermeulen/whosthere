package core

import (
	"context"
	"fmt"
	"log/slog"
	"net"

	"github.com/ramonvermeulen/whosthere/internal/core/config"
	"github.com/ramonvermeulen/whosthere/internal/core/paths"
	"github.com/ramonvermeulen/whosthere/pkg/discovery"
	"github.com/ramonvermeulen/whosthere/pkg/discovery/oui"
	"github.com/ramonvermeulen/whosthere/pkg/discovery/scanners/arp"
	"github.com/ramonvermeulen/whosthere/pkg/discovery/scanners/mdns"
	"github.com/ramonvermeulen/whosthere/pkg/discovery/scanners/ssdp"
	"github.com/ramonvermeulen/whosthere/pkg/discovery/sweeper"
)

func BuildEngine(cfg *config.Config, logger discovery.Logger) (*discovery.Engine, error) {
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

	iface, err := discovery.NewInterfaceInfo(cfg.NetworkInterface)
	if err != nil {
		return nil, err
	}

	targetSubnets, err := parseTargetSubnets(cfg.TargetSubnets)
	if err != nil {
		return nil, err
	}

	var scanners []discovery.Scanner

	if cfg.Scanners.SSDP.Enabled {
		s, err := ssdp.New(iface, ssdp.WithLogger(logger), ssdp.WithTargetSubnets(targetSubnets))
		if err != nil {
			return nil, err
		}
		scanners = append(scanners, s)
	}
	if cfg.Scanners.ARP.Enabled {
		s, err := arp.New(
			iface,
			arp.WithLogger(logger),
			arp.WithAllInterfaces(cfg.AllInterfaces),
			arp.WithTargetSubnets(targetSubnets),
		)
		if err != nil {
			return nil, err
		}
		scanners = append(scanners, s)
	}
	if cfg.Scanners.MDNS.Enabled {
		s, err := mdns.New(iface, mdns.WithLogger(logger), mdns.WithTargetSubnets(targetSubnets))
		if err != nil {
			return nil, err
		}
		scanners = append(scanners, s)
	}

	opts := []discovery.Option{
		discovery.WithInterface(iface),
		discovery.WithScanners(scanners...),
		discovery.WithScanTimeout(cfg.ScanTimeout),
		discovery.WithScanInterval(cfg.ScanInterval),
		discovery.WithLogger(logger),
		discovery.WithTargetSubnets(targetSubnets),
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
			sweeper.WithTargetSubnets(targetSubnets),
			sweeper.WithAllowLargeSubnets(cfg.ScanLargeSubnets),
		}
		s, err := sweeper.New(sweeperOpts...)
		if err != nil {
			return nil, err
		}
		opts = append(opts, discovery.WithSweeper(s))
	}

	return discovery.NewEngine(opts...)
}

func parseTargetSubnets(rawSubnets []string) ([]*net.IPNet, error) {
	if len(rawSubnets) == 0 {
		return []*net.IPNet{}, nil
	}

	subnets := make([]*net.IPNet, 0, len(rawSubnets))
	for _, raw := range rawSubnets {
		ip, ipNet, err := net.ParseCIDR(raw)
		if err != nil {
			return nil, fmt.Errorf("parse target subnet %q: %w", raw, err)
		}
		if ip.To4() == nil {
			return nil, fmt.Errorf("target subnet %q must be an IPv4 CIDR", raw)
		}
		ipNet.IP = ip.Mask(ipNet.Mask)
		subnets = append(subnets, ipNet)
	}

	return subnets, nil
}
