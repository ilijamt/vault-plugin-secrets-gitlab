Flags
=====

There are some flags we can specify to enable/disable some functionality in the vault plugin.

|               Flag                | Default value | Changable during runtime if `allow-runtime-flags-change` is set to `true` | Description                                                                            |
|:---------------------------------:|:-------------:|:-------------------------------------------------------------------------:|----------------------------------------------------------------------------------------|
|         show-config-token         |     false     |                                   true                                    | Display the token value when reading a config on it's endpoint like `/config/default`. |
|    allow-runtime-flags-change     |     false     |                                   false                                   | Allows you to change the flags at runtime                                              |
