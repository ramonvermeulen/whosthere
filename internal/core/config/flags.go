package config

// Flags represents CLI flags provided by the user.
type Flags struct {
	ConfigFile string
	PprofPort  string

	Overrides map[string]string
}
