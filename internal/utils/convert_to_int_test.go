//go:build unit

package utils_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/errs"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/utils"
)

func TestConvertToInt(t *testing.T) {
	var tests = []struct {
		in     any
		outVal int
		outErr error
	}{
		{int(52), int(52), nil},
		{int8(13), int(13), nil},
		{int16(612), int(612), nil},
		{int32(56236), int(56236), nil},
		{int64(23462346), int(23462346), nil},
		{float32(62346.62), int(62346), nil},
		{float64(263467.26), int(263467), nil},
		{"1", int(0), errs.ErrInvalidValue},
	}

	for _, tst := range tests {
		t.Logf("ConvertToInt(%T(%v))", tst.in, tst.in)
		val, err := utils.ConvertToInt(tst.in)
		assert.Equal(t, tst.outVal, val)
		if tst.outErr != nil {
			assert.ErrorIs(t, err, tst.outErr)
		}
	}
}
