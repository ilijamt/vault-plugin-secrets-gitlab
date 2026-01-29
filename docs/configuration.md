Configuration
=============

## Config

|      Property      | Required | Default value | Sensitive | Description                                                                                                                                   |
|:------------------:|:--------:|:-------------:|:---------:|:----------------------------------------------------------------------------------------------------------------------------------------------|
|       token        |   yes    |      n/a      |    yes    | The token to access Gitlab API, it will not show when you do a read, as it's a sensitive value. Instead it will display it's SHA1 hash value. |
|      base_url      |   yes    |      n/a      |    no     | The address to access Gitlab                                                                                                                  |
| auto_rotate_token  |    no    |      no       |    no     | Should we autorotate the token when it's close to expiry? (Experimental)                                                                      |
| auto_rotate_before |    no    |      24h      |    no     | How much time should be remaining on the token validity before we should rotate it? Minimum can be set to 24h and maximum to 730h             |
|        type        |   yes    |      n/a      |    no     | The type of gitlab instance that we use can be one of saas, self-managed or dedicated                                                         |
