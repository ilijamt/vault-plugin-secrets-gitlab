package flags

import (
	"flag"
)

// Flags represent a set of configurable options affecting runtime behavior and configuration visibility.
type Flags struct {
	// ShowConfigToken determines if the configuration token value should be displayed when accessing the configuration endpoint.
	ShowConfigToken bool `json:"show_config_token" mapstructure:"show_config_token"`

	// AllowRuntimeFlagsChange determines whether runtime flags can be dynamically modified during execution.
	AllowRuntimeFlagsChange bool `json:"allow_runtime_flags_change" mapstructure:"allow_runtime_flags_change"`

	// EnableFineGrainedAccessTokens enables fine-grained access tokens, minimal version of gitlab 18.10.
	EnableFineGrainedAccessTokens bool `json:"enable_fine_grained_access_tokens" mapstructure:"enable_fine_grained_access_tokens"`
}

// FlagSet configures the provided FlagSet with flags managed by the Flags struct and returns the updated FlagSet.
func (f *Flags) FlagSet(fs *flag.FlagSet) *flag.FlagSet {
	fs.BoolVar(&f.ShowConfigToken, "show-config-token", false, "Display the token value when reading it's config the configuration endpoint.")
	fs.BoolVar(&f.AllowRuntimeFlagsChange, "allow-runtime-flags-change", false, "Allows you to change the flags dynamically at runtime.")
	fs.BoolVar(&f.EnableFineGrainedAccessTokens, "enable-fine-grained-access-tokens", false, "Enable fine-grained access tokens, minimal version of gitlab 18.10.")
	return fs
}
