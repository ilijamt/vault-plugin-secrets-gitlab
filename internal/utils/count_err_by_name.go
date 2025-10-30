package utils

import (
	"errors"

	"github.com/hashicorp/go-multierror"
)

func CountErrByName(err *multierror.Error) map[string]int {
	data := make(map[string]int)

	if err == nil || err.Errors == nil {
		return data
	}

	for _, e := range err.Errors {
		if e == nil {
			continue
		}

		var name string
		if unwrapped := errors.Unwrap(e); unwrapped != nil {
			name = unwrapped.Error()
		} else {
			name = e.Error()
		}

		data[name]++
	}

	return data
}
