//go:build !windows

package arp

import (
	"context"

	"github.com/ramonvermeulen/whosthere/internal/core/discovery"
)

// readWindowsARPCache is a no-op on non-Windows platforms.
func (s *Scanner) readWindowsARPCache(ctx context.Context, out chan<- discovery.Device) error {
	return nil
}
