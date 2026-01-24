//go:build windows

package arp

import (
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"syscall"
	"unsafe"

	"github.com/ramonvermeulen/whosthere/internal/core/discovery"
	"go.uber.org/zap"
)

// Windows API definitions for GetIpNetTable
// https://learn.microsoft.com/en-us/windows/win32/api/iphlpapi/nf-iphlpapi-getipnettable

const (
	// MAXLEN_PHYSADDR is standard length for physical address in MIB_IPNETROW
	MAXLEN_PHYSADDR = 8
)

// MIB_IPNETROW structure contains information for an ARP table entry.
// We align fields to match Windows packing (usually 4-byte boundaries).
type MIB_IPNETROW struct {
	Index       uint32
	PhysAddrLen uint32
	PhysAddr    [MAXLEN_PHYSADDR]byte
	Addr        uint32
	Type        uint32
}

// MIB_IPNETTABLE structure contains a table of ARP entries.
// The NumEntries field tells us how many rows follow.
// In C this is `DWORD dwNumEntries; MIB_IPNETROW table[ANY_SIZE];`
type MIB_IPNETTABLE struct {
	NumEntries uint32
}

var (
	modiphlpapi       = syscall.NewLazyDLL("iphlpapi.dll")
	procGetIpNetTable = modiphlpapi.NewProc("GetIpNetTable")
)

// readWindowsARPCache retrieves ARP entries using the Windows IP Helper API.
func (s *Scanner) readWindowsARPCache(ctx context.Context, out chan<- discovery.Device) error {
	log := zap.L().With(zap.String("scanner", "arp"))

	entries, err := s.getIpNetTable(ctx)
	if err != nil {
		log.Debug("failed to get windows arp table via API", zap.Error(err))
		return err
	}

	return s.emitARPEntries(ctx, out, entries)
}

// getIpNetTable calls the Windows GetIpNetTable API and converts the result to our generic Entry struct.
func (s *Scanner) getIpNetTable(ctx context.Context) ([]Entry, error) {
	// First call to determine size
	var size uint32
	// Return value 122 (ERROR_INSUFFICIENT_BUFFER) is expected
	r1, _, _ := procGetIpNetTable.Call(
		0,
		uintptr(unsafe.Pointer(&size)),
		0,
	)

	// If it's not ERROR_INSUFFICIENT_BUFFER (122) and not NO_ERROR (0), something is wrong.
	if r1 != 0 && r1 != 122 {
		// Just in case it failed catastrophically, but usually it returns 122 or 87 (invalid param) if buffer is needed
	}

	// Just to be safe, if size is 0, give it some room (e.g. 15kb)
	if size == 0 {
		size = 15000
	}

	buf := make([]byte, size)
	r1, _, _ = procGetIpNetTable.Call(
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(unsafe.Pointer(&size)),
		0, // ensure sorted
	)

	if r1 != 0 {
		if r1 == 122 {
			// Buffer still too small? Try again with new size.
			buf = make([]byte, size)
			r1, _, _ = procGetIpNetTable.Call(
				uintptr(unsafe.Pointer(&buf[0])),
				uintptr(unsafe.Pointer(&size)),
				0,
			)
			if r1 != 0 {
				return nil, fmt.Errorf("GetIpNetTable failed with error code %d", r1)
			}
		} else {
			return nil, fmt.Errorf("GetIpNetTable failed with error code %d", r1)
		}
	}

	// Parse the buffer
	// Number of entries is at the beginning
	numEntries := *(*uint32)(unsafe.Pointer(&buf[0]))

	// MIB_IPNETROW size:
	// Index (4) + PhysAddrLen (4) + PhysAddr (8) + Addr (4) + Type (4) = 24 bytes
	rowSize := uintptr(24) // unsafe.Sizeof(MIB_IPNETROW{}) might have padding - check alignment
	// In Go, struct alignment might add padding.
	//
	// MIB_IPNETROW is:
	// DWORD dwIndex; (4 bytes)
	// DWORD dwPhysAddrLen; (4 bytes)
	// BYTE bPhysAddr[MAXLEN_PHYSADDR]; (8 bytes)
	// DWORD dwAddr; (4 bytes)
	// DWORD dwType; (4 bytes)
	// Total 24 bytes EXACTLY. No padding needed between 4-byte aligned fields.

	// Offset of the first row is after NumEntries (4 bytes).
	// But we must be careful about alignment. NumEntries is 4 bytes.
	// Does the first struct start at offset 4?
	// On 32-bit and 64-bit Windows, alignof(DWORD) is 4.
	// So offset 4 is valid.
	var entries []Entry
	startPtr := uintptr(unsafe.Pointer(&buf[0])) + 4

	for i := uint32(0); i < numEntries; i++ {
		// Check context
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		rowPtr := (*MIB_IPNETROW)(unsafe.Pointer(startPtr + (uintptr(i) * rowSize)))

		// Index must match our interface index
		if int(rowPtr.Index) != s.iface.Interface.Index {
			continue
		}

		// Row Addr is IPv4 address as DWORD (little-endian usually)
		ipBytes := make([]byte, 4)
		binary.LittleEndian.PutUint32(ipBytes, rowPtr.Addr)
		ip := net.IP(ipBytes)

		// Row PhysAddr
		if rowPtr.PhysAddrLen > MAXLEN_PHYSADDR {
			continue // Should not happen
		}

		// Copy valid bytes of MAC
		mac := make(net.HardwareAddr, rowPtr.PhysAddrLen)
		for j := uint32(0); j < rowPtr.PhysAddrLen; j++ {
			mac[j] = rowPtr.PhysAddr[j]
		}

		// According to docs:
		// Type:
		// 1 = Other
		// 2 = Invalid (deleted)
		// 3 = Dynamic
		// 4 = Static
		//
		// We usually want dynamic and static (except multicast/broadcast which we execute filter logic on later).
		// Type 2 is invalid.
		if rowPtr.Type == 2 {
			continue
		}

		entries = append(entries, Entry{
			IP:            ip,
			MAC:           mac,
			InterfaceName: s.iface.Interface.Name,
			Age:           0,
		})
	}

	return entries, nil
}
