package main

import (
	"os"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/vault/api"
	"github.com/hashicorp/vault/sdk/plugin"
	gat "github.com/ilijamt/vault-plugin-secrets-gitlab"
)

var (
	logger = hclog.New(&hclog.LoggerOptions{})
)

func main() {
	apiClientMeta := &api.PluginAPIClientMeta{}
	flags := apiClientMeta.FlagSet()

	fatalIfError(flags.Parse(os.Args[1:]))

	tlsConfig := apiClientMeta.GetTLSConfig()
	tlsProviderFunc := api.VaultPluginTLSProvider(tlsConfig)

	fatalIfError(plugin.ServeMultiplex(&plugin.ServeOpts{
		BackendFactoryFunc: gat.Factory,
		TLSProviderFunc:    tlsProviderFunc,
	}))
}

func fatalIfError(err error) {
	if err != nil {
		logger.Error("plugin shutting down", "error", err)
		os.Exit(1)
	}
}
