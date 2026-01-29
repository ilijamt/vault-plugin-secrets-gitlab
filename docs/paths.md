Paths
=====

For a list of the available endpoints you can check below or by running the command `vault path-help gitlab` for your version after you've mounted it.

```shell
$ vault path-help gitlab
## DESCRIPTION

The Gitlab Access token auth Backend dynamically generates private
and group tokens.

After mounting this Backend, credentials to manage Gitlab tokens must be configured
with the "^config/(?P<config_name>\w(([\w-.@]+)?\w)?)$" endpoints.

## PATHS

The following paths are supported by this backend. To view help for
any of the paths below, use the help command with any route matching
the path pattern. Note that depending on the policy of your auth token,
you may or may not be able to access certain paths.

    ^config/(?P<config_name>\w(([\w-.]+)?\w)?)$
        Configure the Gitlab Access Tokens Backend.

    ^config/(?P<config_name>\w(([\w-.]+)?\w)?)/rotate$
        Rotate the gitlab token for this configuration.

    ^config?/?$
        Lists existing configs

    ^flags$
        Flags for the plugin.

    ^roles/(?P<role_name>\w(([\w-.]+)?\w)?)$
        Create a role with parameters that are used to generate a various access tokens.

    ^roles?/?$
        Lists existing roles

    ^token/(?P<role_name>\w(([\w-.]+)?\w)?)(/(?P<path>.+))?$
        Generate an access token based on the specified role
```
