package gitlab

type TokenGroup struct {
	*TokenWithScopesAndAccessLevel `json:",inline"`
}
