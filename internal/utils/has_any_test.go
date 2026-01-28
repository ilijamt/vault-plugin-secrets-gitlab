package utils_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/utils"
)

func TestHasAny_String(t *testing.T) {
	type args struct {
		s    string
		vals []string
		pred func(string, string) bool
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Has prefix, match",
			args: args{"foobar", []string{"foo", "bar", "baz"}, strings.HasPrefix},
			want: true,
		},
		{
			name: "Has prefix, no match",
			args: args{"foobar", []string{"baz", "qux"}, strings.HasPrefix},
			want: false,
		},
		{
			name: "Has suffix, match",
			args: args{"filename.git", []string{".zip", ".git"}, strings.HasSuffix},
			want: true,
		},
		{
			name: "Has suffix, no match",
			args: args{"filename.txt", []string{".git", ".zip"}, strings.HasSuffix},
			want: false,
		},
		{
			name: "Contains, match",
			args: args{"hello", []string{"e", "z"}, strings.Contains},
			want: true,
		},
		{
			name: "Empty vals",
			args: args{"foo", []string{}, strings.HasPrefix},
			want: false,
		},
		{
			name: "Empty s",
			args: args{"", []string{"foo"}, strings.HasPrefix},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := utils.HasAny(tt.args.s, tt.args.vals, tt.args.pred)
			assert.Equal(t, tt.want, got)
		})
	}
}

// Example for a generic case, e.g. with int slices
func TestHasAny_Int(t *testing.T) {
	isMultiple := func(x, y int) bool { return x%y == 0 }
	assert.True(t, utils.HasAny(12, []int{5, 3, 4}, isMultiple)) // 12 % 3 == 0
	assert.False(t, utils.HasAny(7, []int{2, 4, 6}, isMultiple))
}
