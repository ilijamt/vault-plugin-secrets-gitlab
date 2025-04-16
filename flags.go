package gitlab

import (
	"flag"
)

type Flags struct {
	ShowConfigToken         bool   `json:"show_config_token" mapstructure:"show_config_token"`
	AllowRuntimeFlagsChange bool   `json:"allow_runtime_flags_change" mapstructure:"allow_runtime_flags_change"`
	TelemetryCollection     bool   `json:"telemetry_collection" mapstructure:"telemetry_collection"`
	TelemetryEndpoint       string `json:"telemetry_endpoint" mapstructure:"telemetry_endpoint"`
}

// FlagSet returns the flag set for configuring the TLS connection
func (f *Flags) FlagSet(fs *flag.FlagSet) *flag.FlagSet {
	fs.BoolVar(&f.ShowConfigToken, "show-config-token", false, "Display the token value when reading it's config the configuration endpoint.")
	fs.BoolVar(&f.TelemetryCollection, "telemetry-collection", false, "Should we collect telemetry data.")
	fs.StringVar(&f.TelemetryEndpoint, "telemetry-endpoint", "https://vault-plugin-secrets-gitlab-telemetry.matoski.com", "Telemetry endpoint.")
	fs.BoolVar(&f.AllowRuntimeFlagsChange, "allow-runtime-flags-change", false, "Allows you to change the flags dynamically at runtime.")
	return fs
}
