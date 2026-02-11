package discovery

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// to ensure error messages do not change for homebrew-core
// see https://github.com/Homebrew/homebrew-core/blob/main/Formula/w/whosthere.rb
func TestNewInterfaceInfo_InvalidInterface(t *testing.T) {
	_, err := NewInterfaceInfo("non_existing")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no such network interface")
}
