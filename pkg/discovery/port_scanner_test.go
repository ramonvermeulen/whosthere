package discovery

import (
	"context"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// mockConn is a minimal implementation of net.Conn for testing.
type mockConn struct{}

func (m *mockConn) Read(b []byte) (n int, err error)   { return 0, nil }
func (m *mockConn) Write(b []byte) (n int, err error)  { return len(b), nil }
func (m *mockConn) Close() error                       { return nil }
func (m *mockConn) LocalAddr() net.Addr                { return nil }
func (m *mockConn) RemoteAddr() net.Addr               { return nil }
func (m *mockConn) SetDeadline(t time.Time) error      { return nil }
func (m *mockConn) SetReadDeadline(t time.Time) error  { return nil }
func (m *mockConn) SetWriteDeadline(t time.Time) error { return nil }

// mockDialer is a mock implementation of Dialer for testing.
type mockDialer struct {
	openPorts map[string]bool
}

func (m *mockDialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	if m.openPorts[address] {
		return &mockConn{}, nil // simulate open
	}
	return nil, net.ErrClosed // simulate closed
}

func TestPortScanner_Stream(t *testing.T) {
	mock := &mockDialer{
		openPorts: map[string]bool{
			"127.0.0.1:80":  true,
			"127.0.0.1:443": true,
		},
	}
	ps := &PortScanner{
		workers: 2,
		dialer:  mock,
	}

	ports := []int{22, 80, 443, 8080}
	openPorts := make(map[int]struct{})
	var mu sync.Mutex

	err := ps.Stream(context.Background(), "127.0.0.1", ports, 100*time.Millisecond, func(port int) {
		mu.Lock()
		openPorts[port] = struct{}{}
		mu.Unlock()
	})
	require.NoError(t, err)

	require.Contains(t, openPorts, 80)
	require.Contains(t, openPorts, 443)
	require.Len(t, openPorts, 2)
}

func TestPortScanner_Stream_EmptyPorts(t *testing.T) {
	ps := NewPortScanner(1, nil)
	openPorts := make(map[int]struct{})
	var mu sync.Mutex

	err := ps.Stream(context.Background(), "127.0.0.1", nil, 100*time.Millisecond, func(port int) {
		mu.Lock()
		openPorts[port] = struct{}{}
		mu.Unlock()
	})
	require.NoError(t, err)
	require.Empty(t, openPorts)
}
