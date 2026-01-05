//go:build !linux

package arp

import (
	"context"

	"github.com/ramonvermeulen/whosthere/internal/discovery"
)

// readLinuxARPCache is a no-op on non-Linux platforms; real impl is linux-only.
// this stub keeps the function call valid on other platforms to avoid build errors.
func (s *Scanner) readLinuxARPCache(ctx context.Context, out chan<- discovery.Device) error {
	return nil
}
