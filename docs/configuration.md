Configuration
=============

## Config

|      Property      | Required | Default value | Sensitive | Description                                                                                                                                   |
|:------------------:|:--------:|:-------------:|:---------:|:----------------------------------------------------------------------------------------------------------------------------------------------|
|       token        |   yes    |      n/a      |    yes    | The token to access Gitlab API, it will not show when you do a read, as it's a sensitive value. Instead it will display its SHA1 hash value. |
|      base_url      |   yes    |      n/a      |    no     | The address to access Gitlab                                                                                                                  |
| auto_rotate_token  |    no    |      no       |    no     | Should we autorotate the token when it's close to expiry? (Experimental)                                                                      |
| auto_rotate_before |    no    |      24h      |    no     | How much time should be remaining on the token validity before we should rotate it? Minimum can be set to 24h and maximum to 730h             |
|        type        |   yes    |      n/a      |    no     | The type of gitlab instance that we use can be one of saas, self-managed or dedicated                                                         |

## GitLab information

Writing a config stores what the plugin can learn about the GitLab instance, and reading the config shows it as `gitlab_version`, `gitlab_instance` and `last_refreshed_at`. It is refreshed about once an hour, so a GitLab upgrade is picked up without writing the config again.

`gitlab_instance` holds a few settings, such as whether fine-grained tokens are enforced. Reading them needs an administrator, or a fine-grained token with `read_application_setting`. Without that, `settings_known` is `false` and the plugin works as normal. If `token_owner_is_admin` is `true` and `settings_known` is `false`, Admin Mode is probably on and the config token needs the `admin_mode` scope.

`personal_tokens_disabled` is empty when GitLab doesn't return that setting, which happens on instances without a paid license.

These values are a hint and can be up to an hour old. GitLab always has the final say.
