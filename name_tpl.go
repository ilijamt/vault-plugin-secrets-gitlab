package gitlab

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

var tplFuncMap = template.FuncMap{
	"randHexString": randHexString,
	"stringsJoin":   strings.Join,
	"yesNoBool":     yesNoBool,
	"timeNowFormat": timeNowFormat,
}

func TokenName(role *EntryRole) (name string, err error) {
	if role == nil {
		return "", fmt.Errorf("role: %w", errs.ErrNilValue)
	}
	var tpl *template.Template
	tpl, err = template.New("name").Funcs(tplFuncMap).Parse(role.Name)
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
