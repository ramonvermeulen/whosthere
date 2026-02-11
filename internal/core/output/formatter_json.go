package output

import (
	"encoding/json"
	"io"

	"github.com/ramonvermeulen/whosthere/pkg/discovery"
)

var _ Formatter = (*JSONFormatter)(nil)

// JSONFormatter implements Formatter for JSON output
type JSONFormatter struct {
	pretty bool
}

func NewJSONFormatter(pretty bool) *JSONFormatter {
	return &JSONFormatter{pretty: pretty}
}

func (f *JSONFormatter) Format(w io.Writer, results *discovery.ScanResults) error {
	encoder := json.NewEncoder(w)
	if f.pretty {
		encoder.SetIndent("", "  ")
	}
	return encoder.Encode(results)
}
