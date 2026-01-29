//go:build linux

package arp

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/ramonvermeulen/whosthere/internal/core/discovery"
	"go.uber.org/zap"
)

// readLinuxARPCache reads /proc/net/arp and emits completed entries.
// see https://man7.org/linux/man-pages/man5/proc_pid_net.5.html for more information about /proc/net/arp.
func (s *Scanner) readLinuxARPCache(ctx context.Context, out chan<- discovery.Device) error {
	log := zap.L().With(zap.String("scanner", s.Name()))

	entries, err := parseProcNetARP(ctx, "/proc/net/arp")
	if err != nil {
		log.Debug("failed to parse /proc/net/arp", zap.Error(err))
		return err
	}
	return s.emitARPEntries(ctx, out, entries)
}

// parseProcNetARP parses the ARP table file at the given path.
// It returns a slice of completed ARP entries.
// The standard format for the arp table is:
// IP address     HW type   Flags     HW address          Mask   Device
// 192.168.0.50   0x1       0x2       00:50:BF:25:68:F3   *      eth0
// 192.168.0.250  0x1       0xc       00:00:00:00:00:00   *      eth0
func parseProcNetARP(ctx context.Context, path string) ([]Entry, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	defer func() {
		_ = f.Close()
	}()

	scanner := bufio.NewScanner(f)
	if !scanner.Scan() {
		return nil, fmt.Errorf("empty %s", path)
	}

	var entries []Entry
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		fields := strings.Fields(scanner.Text())
		if len(fields) < 6 {
			continue
		}

		ip := net.ParseIP(fields[0])
		if ip == nil {
			continue
		}

		// Flags column is hex; require 0x2 (completed)
		flagsStr := fields[2]
		flags, err := strconv.ParseInt(strings.TrimPrefix(flagsStr, "0x"), 16, 64)
		if err != nil || flags&0x2 == 0 {
			continue
		}

		mac, err := net.ParseMAC(fields[3])
		if err != nil || len(mac) == 0 || isBroadcastMAC(mac) || isMulticastMAC(mac) {
			continue
		}

		interfaceName := fields[5]

		entries = append(entries, Entry{IP: ip, MAC: mac, InterfaceName: interfaceName})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan %s: %w", path, err)
	}

	return entries, nil
}
