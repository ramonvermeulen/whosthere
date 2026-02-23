package mdns

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"testing"
	"time"

	hashimdns "github.com/hashicorp/mdns"
	"github.com/ramonvermeulen/whosthere/pkg/discovery"
	"github.com/stretchr/testify/require"
)

func TestNewScanner_Defaults(t *testing.T) {
	iface := &discovery.InterfaceInfo{}
	s, err := New(iface)
	require.NoError(t, err)
	require.NotNil(t, s)
	require.Equal(t, iface, s.iface)
	require.IsType(t, discovery.NoOpLogger{}, s.logger)
}

func TestNewScanner_WithLogger(t *testing.T) {
	iface := &discovery.InterfaceInfo{}
	logger := testLogger{}
	s, err := New(iface, WithLogger(logger))
	require.NoError(t, err)
	require.Equal(t, logger, s.logger)
}

func TestNewScanner_WithLoggerNil(t *testing.T) {
	iface := &discovery.InterfaceInfo{}
	_, err := New(iface, WithLogger(nil))
	require.Error(t, err)
	require.Contains(t, err.Error(), "logger cannot be nil")
}

func TestScanner_Name(t *testing.T) {
	s, _ := New(nil)
	require.Equal(t, "mdns", s.Name())
}

type testLogger struct{}

func (testLogger) Log(_ context.Context, _ slog.Level, _ string, _ ...any) {}

func Test_splitKeyValue(t *testing.T) {
	tests := []struct {
		in   string
		want []string
	}{
		{"foo=bar", []string{"foo", "bar"}},
		{"foo=", []string{"foo", ""}},
		{"=bar", []string{"", "bar"}},
		{"foo", nil},
		{"", nil},
	}
	for _, tt := range tests {
		got := splitKeyValue(tt.in)
		if tt.want == nil {
			require.Nil(t, got)
		} else {
			require.Equal(t, tt.want, got)
		}
	}
}

// withTestQueryFunc is a test-only option for injecting a mock query function.
func withTestQueryFunc(f func(params *hashimdns.QueryParam) error) Option {
	return func(s *Scanner) error {
		s.queryFunc = f
		return nil
	}
}

func TestScan_ContextCancel(t *testing.T) {
	iface := &discovery.InterfaceInfo{}
	s, _ := New(iface, withTestQueryFunc(func(params *hashimdns.QueryParam) error {
		time.Sleep(10 * time.Millisecond)
		return nil
	}))
	ctx, cancel := context.WithCancel(context.Background())
	results := make(chan *discovery.Device, 1)
	cancel()
	err := s.Scan(ctx, results)
	require.ErrorIs(t, err, context.Canceled)
}

func TestScan_NoEntries(t *testing.T) {
	iface := &discovery.InterfaceInfo{}
	s, _ := New(iface, withTestQueryFunc(func(params *hashimdns.QueryParam) error {
		return nil
	}))
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	results := make(chan *discovery.Device, 1)
	err := s.Scan(ctx, results)
	require.NoError(t, err)
}

func TestScan_ErrorPropagation(t *testing.T) {
	testErr := errors.New("query failed")
	iface := &discovery.InterfaceInfo{}
	s, _ := New(iface, withTestQueryFunc(func(params *hashimdns.QueryParam) error {
		return testErr
	}))
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	results := make(chan *discovery.Device, 1)
	err := s.Scan(ctx, results)
	require.ErrorIs(t, err, testErr)
}

func TestScan_EntryWithInfoFields(t *testing.T) {
	iface := &discovery.InterfaceInfo{}
	s, _ := New(iface, withTestQueryFunc(func(params *hashimdns.QueryParam) error {
		params.Entries <- &hashimdns.ServiceEntry{
			AddrV4:     net.ParseIP("1.2.3.4"),
			Name:       "testdev",
			Info:       "foo=bar",
			InfoFields: []string{"foo=bar", "baz"},
		}
		close(params.Entries)
		return nil
	}))
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	results := make(chan *discovery.Device, 1)
	err := s.Scan(ctx, results)
	require.NoError(t, err)
	dev := <-results
	require.Equal(t, "1.2.3.4", dev.IP().String())
	require.Equal(t, "testdev", dev.DisplayName())
	require.Equal(t, "bar", dev.ExtraData()["foo"])
	require.Equal(t, "true", dev.ExtraData()["baz"])
}
