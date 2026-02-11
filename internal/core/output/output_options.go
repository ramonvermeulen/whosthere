package output

import "github.com/ramonvermeulen/whosthere/pkg/discovery"

type Option func(output *Output) error

func WithPretty() Option {
	return func(o *Output) error {
		o.pretty = true
		return nil
	}
}

func WithSort(sortFunc func(a, b *discovery.Device) bool) Option {
	return func(o *Output) error {
		o.sortFunc = sortFunc
		return nil
	}
}
