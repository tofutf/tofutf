# Environment variables

tofutf can be configured from environment variables. Arguments can be converted to the equivalent env var by prefixing
it with `tofutf_`, replacing all `-` with `_`, and upper-casing it. For example:

- `--secret` becomes `tofutf_SECRET`
- `--site-token` becomes `tofutf_SITE_TOKEN`

Env variables can be suffixed with `_FILE` to tell tofutf to read the values from a file. This is useful for container
environments where secrets are often mounted as files.
