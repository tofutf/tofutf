# Flags

## `--address`

* System: `tofutfd`, `tofutf-agent`
* Default: `localhost:8080`

Sets the listening address of an `tofutfd` node.

Set the port to an empty string or to `0` to choose a random available port.

Set the address to an empty string to listen on all interfaces. For example, the
following listens on all interfaces using a random port:

```
tofutfd --address :0
```

## `--cache-expiry`

* System: `tofutfd`
* Default: `10 minutes`

Set the TTL for cache entries.

## `--cache-size`

* System: `tofutfd`
* Default: `0` (unlimited)

Cache size in MB. The cache is stored in RAM. Default is `0` which means it'll use an unlimited amount of RAM.

It is recommended that you set this to an appropriate size in a production
deployment, taking into consideration the [cache expiry](#-cache-expiry).

## `--concurrency`

* System: `tofutfd`, `tofutf-agent`
* Default: 5

Sets the number of workers that can process runs concurrently.

## `--dev-mode`

* System: `tofutfd`
* Default: `false`

Enables developer mode:

1. Static files are loaded from disk rather than from those embedded within the `tofutfd` binary.
2. Enables [livereload](https://github.com/livereload/livereload-js).

This means you can make changes to CSS, templates, etc, and you automatically see the changes in the browser in real-time.

If developer mode were disabled, you would need to re-build the `tofutfd` binary and then manually reload the page in your browser.

!!! note
    Ensure you have cloned the git repository to your local filesystem and that you have started `tofutfd` from the root of the repository, otherwise it will not be able to locate the static files.

## `--github-client-id`

* System: `tofutfd`
* Default: ""

Github OAuth Client ID. Set this flag along with [--github-client-secret](#-github-client-secret) to enable [Github authentication](../auth/providers/github.md).

## `--github-client-secret`

* System: `tofutfd`
* Default: ""

Github OAuth client secret. Set this flag along with [--github-client-id](#-github-client-id) to enable [Github authentication](../auth/providers/github.md).

## `--gitlab-client-id`

* System: `tofutfd`
* Default: ""

Gitlab OAuth Client ID. Set this flag along with [--gitlab-client-secret](#-gitlab-client-secret) to enable [Gitlab authentication](../auth/providers/gitlab.md).

## `--gitlab-client-secret`

* System: `tofutfd`
* Default: ""

Gitlab OAuth client secret. Set this flag along with [--gitlab-client-id](#-gitlab-client-id) to enable [Gitlab authentication](../auth/providers/gitlab.md).

## `--google-jwt-audience`

* System: `tofutfd`
* Default: ""

The Google JWT audience claim for validation. If unspecified then the audience
claim is not validated. See the [Google IAP](../auth/providers/iap.md#verification) document for more details.

## `--hostname`

* System: `tofutfd`
* Default: `localhost:8080` or `--address` if specified.

Sets the hostname that clients can use to access the tofutf cluster. This value is
used within links sent to various clients, including:

* The `terraform` CLI when it is streaming logs for a remote `plan` or `apply`.
* Pull requests on VCS providers, e.g. the link beside the status check on a
Github pull request.

It is highly advisable to set this flag in a production deployment.

## `--webhook-hostname`

* System: `tofutfd`
* Default: `localhost:8080` or `--address` if specified.

Sets the hostname that VCS providers can use to access the tofutf webhooks.

## `--log-format`

* System: `tofutfd`, `tofutf-agent`
* Default: `default`

Set the logging format. Can be one of:

* `default`: human-friendly, not easy to parse, writes to stderr
* `text`: sequence of key=value pairs, writes to stdout
* `json`: json format, writes to stdout

## `--max-config-size`

* System: `tofutfd`
* Default: `104865760` (10MiB)

Maximum permitted configuration upload size. This refers to the size of the (compressed) configuration tarball that `terraform` uploads to tofutf at the start of a remote plan/apply.

## `--oidc-client-id`

* System: `tofutfd`
* Default: ""

OIDC Client ID. Set this flag along with [--oidc-client-secret](#-oidc-client-secret) to enable [OIDC authentication](../auth/providers/oidc.md).

## `--oidc-client-secret`

* System: `tofutfd`
* Default: ""

OIDC Client Secret. Set this flag along with [--oidc-client-id](#-oidc-client-id) to enable [OIDC authentication](../auth/providers/oidc.md).

## `--oidc-issuer-url`

* System: `tofutfd`
* Default: ""

OIDC Issuer URL for OIDC authentication.

## `--oidc-name`

* System: `tofutfd`
* Default: ""

User friendly OIDC name - this is the name of the OIDC provider shown on the login prompt on the web UI.

## `--oidc-scopes`

* System: `tofutfd`
* Default: [openid,profile]

OIDC scopes to request from OIDC provider.

## `--oidc-username-claim`

* System: `tofutfd`
* Default: "name"

OIDC claim for mapping to an tofutf username. Must be one of `name`, `email`, or `sub`.

## `--otel`

* System: `tofutfd`
* Default: false

Enable open telemetry integration. The integration is configured via the normal otel environment variables.

## `--restrict-org-creation`

* System: `tofutfd`
* Default: false

Restricts the ability to create organizations to users possessing the site admin role. By default _any_ user can create organizations.

## `--sandbox`

* System: `tofutfd`
* Default: false

Enable sandbox box; isolates `terraform apply` using [bubblewrap](https://github.com/containers/bubblewrap) for additional security.

## `--secret`

* **Required**
* System: `tofutfd`
* Default: ""

Hex-encoded 16-byte secret for performing cryptographic work. You should use a cryptographically secure random number generator, e.g. `openssl`:

```bash
> openssl rand -hex 16
6b07b57377755b07cf61709780ee7484
```

!!! note
    The secret is required. It must be exactly 16 bytes in size, and it must be hex-encoded.

## `--site-admins`

* System: `tofutfd`
* Default: []

Promote users to the role of site admin. Specify their usernames, separated by a
comma. For example:

```
tofutfd --site-admins bob@example.com,alice@example.com
```

Users are automatically created if they don't exist already.

## `--site-token`

* System: `tofutfd`
* Default: ""

The site token for authenticating with the built-in [`site-admin`](../auth/site_admins.md) user, e.g.:

```bash
tofutfd --site-token=643f57a1016cdde7e7e39914785d36d61fd
```

The default, an empty string, disables the site admin account.

## `--v`, `-v`

* System: `tofutfd`, `tofutf-agent`
* Default: `0`

Set logging verbosity. The higher the number the more verbose the logs. Each number translates to a `level` log field like so:

|verbosity|level|
|-|-|
|0|INFO|
|1|DEBUG|
|2|DEBUG-1|
|3|DEBUG-2|
|n|DEBUG-(n+1)|
