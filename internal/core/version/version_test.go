package version

import (
	"bytes"
	"strings"
	"testing"
)

func TestFprint(t *testing.T) {
	var buf bytes.Buffer
	Fprint(&buf)
	output := buf.String()

	if !strings.Contains(output, "OS:") {
		t.Errorf("expected OS in output")
	}
	if !strings.Contains(output, "Version:") {
		t.Errorf("expected Version in output")
	}
	if !strings.Contains(output, "Commit:") {
		t.Errorf("expected Commit in output")
	}
	if !strings.Contains(output, "Date:") {
		t.Errorf("expected Date in output")
	}
	if !strings.Contains(output, Version) {
		t.Errorf("expected version value %s", Version)
	}
}
