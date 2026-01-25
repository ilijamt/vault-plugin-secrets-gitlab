package flags

import (
	"flag"
)

// Flags represent a set of configurable options affecting runtime behavior and configuration visibility.
type Flags struct {
	// ShowConfigToken determines if the configuration token value should be displayed when accessing the configuration endpoint.
	ShowConfigToken bool `json:"show_config_token" mapstructure:"show_config_token"`

	// AllowPathOverride determines if the path can be overridden in a defined role.
	AllowPathOverride bool `json:"allow_path_override" mapstructure:"allow_path_override"`

	// AllowRuntimeFlagsChange determines whether runtime flags can be dynamically modified during execution.
	AllowRuntimeFlagsChange bool `json:"allow_runtime_flags_change" mapstructure:"allow_runtime_flags_change"`
}

// FlagSet configures the provided FlagSet with flags managed by the Flags struct and returns the updated FlagSet.
func (f *Flags) FlagSet(fs *flag.FlagSet) *flag.FlagSet {
	fs.BoolVar(&f.ShowConfigToken, "show-config-token", false, "Display the token value when reading it's config the configuration endpoint.")
	fs.BoolVar(&f.AllowRuntimeFlagsChange, "allow-runtime-flags-change", false, "Allows you to change the flags dynamically at runtime.")
	fs.BoolVar(&f.AllowPathOverride, "allow-path-override", false, "Allows you to override the path for a specific role.")
	return fs
}
