package devicemeta

import (
	"net"
	"strings"
)

// NormalizeMAC returns a canonical MAC address in lowercase colon-separated form.
func NormalizeMAC(mac string) (string, bool) {
	parsed, err := net.ParseMAC(strings.TrimSpace(mac))
	if err != nil {
		return "", false
	}

	return parsed.String(), true
}
