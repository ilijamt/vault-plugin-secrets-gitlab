package backend

const Help = `
The Gitlab Access token auth Backend dynamically generates private 
and group tokens.

After mounting this Backend, credentials to manage Gitlab tokens must be configured 
with the "^config/(?P<config_name>\w(([\w-.@]+)?\w)?)$" endpoints.
`
