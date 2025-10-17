package utils

import (
	"crypto/rand"
	"fmt"
	"strings"
	"text/template"
	"time"
	_ "unsafe"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/errs"
)

func yesNoBool(in bool) string {
	if in {
		return "yes"
	}
	return "no"
}
func randHexString(bytes int) string {
	buf := make([]byte, bytes)
	_, _ = rand.Read(buf)
	return fmt.Sprintf("%x", buf)
}

func timeNowFormat(layout string) string {
	return time.Now().UTC().Format(layout)
}

func stringsJoin(elems []string, sep string) string {
	for i := range elems {
		elems[i] = strings.TrimSpace(elems[i])
	}
	return strings.Join(elems, sep)
}

func stringsSplit(s, sep string) (out []string) {
	out = strings.Split(s, sep)
	for i := range out {
		out[i] = strings.TrimSpace(out[i])
	}
	return out
}

var tplFuncMap = template.FuncMap{
	"randHexString":  randHexString,
	"stringsJoin":    stringsJoin,
	"yesNoBool":      yesNoBool,
	"timeNowFormat":  timeNowFormat,
	"trimSpace":      strings.TrimSpace,
	"stringsSplit":   stringsSplit,
	"stringsReplace": strings.Replace,
}

// TokenNameData defines an interface for objects that contain a token name and
// methods for obtaining data relevant to token-based operations.
//
// This interface provides a contract for structures that need to offer
// token name data and conversion capabilities. It is used to ensure consistent
// handling of token names and associated logic.
type TokenNameData interface {
	// GetName returns the token's name as a string
	GetName() string
	// LogicalResponseData returns a map containing relevant data that can be used in template operations or logical evaluations
	LogicalResponseData() map[string]any
	// IsNil returns a boolean indicating whether the instance is considered nil or invalid
	IsNil() bool
}

// ValidateTokenNameName validates the template syntax of a token name.
//
// This function checks if the provided TokenNameData instance is non-nil and executes
// basic validation of the token name's syntax by parsing it as a template. This helps
// ensure the token name format adheres to expected patterns and contains no syntax errors.
func ValidateTokenNameName(role TokenNameData) (err error) {
	if role == nil || role.IsNil() {
		return fmt.Errorf("role: %w", errs.ErrNilValue)
	}
	_, err = template.New("name").Funcs(tplFuncMap).Parse(role.GetName())
	return err
}

// TokenName generates a token name by executing the template defined in TokenNameData.
//
// This function retrieves the template string from the TokenNameData, parses it, and
// then executes it while substituting placeholders with the logical response data
// provided by the token role. An additional "unix_timestamp_utc" field is added to the
// data map, representing the current UTC Unix timestamp.
func TokenName(role TokenNameData) (name string, err error) {
	if role == nil || role.IsNil() {
		return "", fmt.Errorf("role: %w", errs.ErrNilValue)
	}
	var tpl *template.Template
	tpl, err = template.New("name").Funcs(tplFuncMap).Parse(role.GetName())
	if err != nil {
		return "", err
	}
	buf := new(strings.Builder)
	var data = role.LogicalResponseData()
	data["unix_timestamp_utc"] = time.Now().UTC().Unix()
	delete(data, "name")
	err = tpl.Execute(buf, data)
	name = buf.String()
	return name, err
}
