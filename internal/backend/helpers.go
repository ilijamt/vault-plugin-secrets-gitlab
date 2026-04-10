package backend

import "cmp"

func configName(name string) string {
	return cmp.Or(name, DefaultConfigName)
}
