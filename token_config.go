package gitlab

type TokenConfig struct {
	TokenWithScopes `json:",inline"`

	UserID int `json:"user_id"`
}
