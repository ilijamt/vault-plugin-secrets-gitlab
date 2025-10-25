package utils

import (
	"errors"

	"github.com/hashicorp/go-multierror"
)

func CountErrByName(err *multierror.Error) map[string]int {
	var data = make(map[string]int)

	for _, e := range err.Errors {
		name := errors.Unwrap(e).Error()
		if _, ok := data[name]; !ok {
			data[name] = 0
		}
		data[name]++
	}

	return data
}
