package gitlab

import "flag"

type Flags struct {
	ShowConfigToken bool
}

// FlagSet returns the flag set for configuring the TLS connection
func (f *Flags) FlagSet(fs *flag.FlagSet) *flag.FlagSet {
	fs.BoolVar(&f.ShowConfigToken, "show-config-token", false, "")
	return fs
}
