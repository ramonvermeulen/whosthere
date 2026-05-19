package state

import (
	"github.com/ramonvermeulen/whosthere/internal/core/devicemeta"
	"github.com/ramonvermeulen/whosthere/pkg/discovery"
)

func normalizedDeviceMAC(d *discovery.Device) (string, bool) {
	if d == nil {
		return "", false
	}

	return devicemeta.NormalizeMAC(d.MAC())
}
