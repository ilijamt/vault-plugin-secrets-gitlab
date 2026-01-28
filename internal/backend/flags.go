package backend

import (
	"github.com/hashicorp/vault/sdk/framework"
)

const FlagsPath = "flags"
const FlagsHelpSynopsis = `Flags for the plugin.`
const FlagsHelpDescription = ``

var FlagsFieldSchema = map[string]*framework.FieldSchema{
	"show_config_token": {
		Type:         framework.TypeBool,
		Description:  "Should we display the token value for the roles?",
		Default:      false,
		DisplayAttrs: &framework.DisplayAttributes{Name: "Show Config Token"},
	},
}
