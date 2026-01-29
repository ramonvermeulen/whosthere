//go:build windows

package arp

import (
	"context"
	"encoding/binary"
	"fmt"
	"log/slog"
	"net"
	"syscall"
	"unsafe"

	"github.com/ramonvermeulen/whosthere/pkg/discovery"
	"golang.org/x/sys/windows"
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

// MIB_IPNETTABLE structure contains the table of ARP entries.
type MIB_IPNETTABLE struct {
	NumEntries uint32
	Table      [1]MIB_IPNETROW // Variable length array
}

var (
	modiphlpapi       = syscall.NewLazyDLL("iphlpapi.dll")
	procGetIpNetTable = modiphlpapi.NewProc("GetIpNetTable")
)

// readWindowsARPCache retrieves ARP entries using the Windows IP Helper API.
func (s *Scanner) readWindowsARPCache(ctx context.Context, out chan<- *discovery.Device) error {
	entries, err := s.getIpNetTable(ctx)
	if err != nil {
		s.logger.Log(ctx, slog.LevelDebug, "failed to get windows arp table via API", "error", err, "scanner", s.Name())
		return err
	}

	return s.emitARPEntries(ctx, out, entries)
}

// getIpNetTable calls the Windows GetIpNetTable API and converts the result to our generic Entry struct.
func (s *Scanner) getIpNetTable(ctx context.Context) ([]Entry, error) {
	// First call to determine size
	var size uint32
	// Use syscall to call the procedure
	r1, _, _ := procGetIpNetTable.Call(
		0,
		uintptr(unsafe.Pointer(&size)),
		0,
	)

	if r1 != 0 && syscall.Errno(r1) != windows.ERROR_INSUFFICIENT_BUFFER {
		return nil, fmt.Errorf("GetIpNetTable failed obtaining size: %d", r1)
	}

	// Just to be safe, if size is 0, give it some room (e.g. 15kb)
	if size == 0 {
		size = 15000
	}

	buf := make([]byte, size)
	r1, _, _ = procGetIpNetTable.Call(
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(unsafe.Pointer(&size)),
		0, // unsorted
	)

	if r1 != 0 {
		if syscall.Errno(r1) == windows.ERROR_INSUFFICIENT_BUFFER {
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
	if len(buf) < 4 {
		return nil, fmt.Errorf("buffer too small for MIB_IPNETTABLE header: %d", len(buf))
	}

	// Cast to MIB_IPNETTABLE pointer
	table := (*MIB_IPNETTABLE)(unsafe.Pointer(&buf[0]))
	numEntries := table.NumEntries

	if numEntries == 0 {
		return []Entry{}, nil
	}

	// Verify buffer size
	// Offset of Table field is 4 bytes.
	// Each row is sizeof(MIB_IPNETROW).
	// We use unsafe.Sizeof to be robust to struct layout.
	rowSize := unsafe.Sizeof(MIB_IPNETROW{})
	expectedSize := uintptr(4) + uintptr(numEntries)*rowSize

	if uintptr(len(buf)) < expectedSize {
		return nil, fmt.Errorf("buffer too small for %d entries: expected %d, got %d", numEntries, expectedSize, len(buf))
	}

	var entries []Entry

	// Safe to create slice now
	rows := unsafe.Slice(&table.Table[0], numEntries)

	for _, row := range rows {
		// Check context
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		// Index must match our interface index
		if int(row.Index) != s.iface.Interface.Index {
			continue
		}

		// Row Addr is IPv4 address as DWORD (little-endian usually)
		ipBytes := make([]byte, 4)
		binary.LittleEndian.PutUint32(ipBytes, row.Addr)
		ip := net.IP(ipBytes)

		// Row PhysAddr
		if row.PhysAddrLen > MAXLEN_PHYSADDR {
			continue // Should not happen
		}

		// Copy valid bytes of MAC
		mac := make(net.HardwareAddr, row.PhysAddrLen)
		for j := uint32(0); j < row.PhysAddrLen; j++ {
			mac[j] = row.PhysAddr[j]
		}

		// Windows MIB_IPNETROW Type constants:
		// 1 = Other
		// 2 = Invalid (deleted)
		// 3 = Dynamic
		// 4 = Static
		// Type 2 is invalid.
		if row.Type == 2 {
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
