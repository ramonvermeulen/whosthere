package version

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFprint(t *testing.T) {
	var buf bytes.Buffer
	Fprint(&buf)
	output := buf.String()

	assert.Contains(t, output, "OS:", "expected OS in output")
	assert.Contains(t, output, "Version:", "expected Version in output")
	assert.Contains(t, output, "Commit:", "expected Commit in output")
	assert.Contains(t, output, "Date:", "expected Date in output")
	assert.Contains(t, output, Version, "expected version value")
}
