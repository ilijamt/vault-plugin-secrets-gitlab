package logging

import (
	"crypto/sha1"
	"fmt"
	"maps"
	"slices"
	"time"

	"github.com/hashicorp/go-hclog"
)

type Fields struct {
	ConfigName  string
	RoleName    string
	TokenType   string
	TokenID     int64
	Token       string // logged only as token_sha1_hash
	Name        string
	Path        string
	Username    string
	UserID      int64
	GroupID     string
	ProjectID   string
	AccessLevel string
	Scopes      []string
	ExpiresAt   *time.Time
	Err         error
	// Extra must never hold anything sensitive.
	Extra map[string]any
}

func (f Fields) args() (args []any) {
	add := func(key string, value any, set bool) {
		if set {
			args = append(args, key, value)
		}
	}
	add("config_name", f.ConfigName, f.ConfigName != "")
	add("role_name", f.RoleName, f.RoleName != "")
	add("token_type", f.TokenType, f.TokenType != "")
	add("token_id", f.TokenID, f.TokenID != 0)
	add("token_sha1_hash", Hash(f.Token), f.Token != "")
	add("name", f.Name, f.Name != "")
	add("path", f.Path, f.Path != "")
	add("username", f.Username, f.Username != "")
	add("user_id", f.UserID, f.UserID != 0)
	add("group_id", f.GroupID, f.GroupID != "")
	add("project_id", f.ProjectID, f.ProjectID != "")
	add("access_level", f.AccessLevel, f.AccessLevel != "")
	add("scopes", f.Scopes, len(f.Scopes) > 0)
	add("expires_at", f.ExpiresAt, f.ExpiresAt != nil)
	add("error", f.Err, f.Err != nil)
	for _, key := range slices.Sorted(maps.Keys(f.Extra)) {
		add(key, f.Extra[key], true)
	}
	return args
}

func Hash(value string) string {
	return fmt.Sprintf("%x", sha1.Sum([]byte(value)))
}

// Logger keeps the hclog.Logger private so calls cannot bypass Fields.
type Logger struct {
	logger hclog.Logger
}

func New(logger hclog.Logger) Logger {
	return Logger{logger: logger}
}

func (l Logger) With(f Fields) Logger {
	return Logger{logger: l.logger.With(f.args()...)}
}

func (l Logger) Trace(msg string, fields ...Fields) { l.logger.Trace(msg, flatten(fields)...) }
func (l Logger) Debug(msg string, fields ...Fields) { l.logger.Debug(msg, flatten(fields)...) }
func (l Logger) Info(msg string, fields ...Fields)  { l.logger.Info(msg, flatten(fields)...) }
func (l Logger) Warn(msg string, fields ...Fields)  { l.logger.Warn(msg, flatten(fields)...) }
func (l Logger) Error(msg string, fields ...Fields) { l.logger.Error(msg, flatten(fields)...) }

func flatten(fields []Fields) (args []any) {
	for _, f := range fields {
		args = append(args, f.args()...)
	}
	return args
}
