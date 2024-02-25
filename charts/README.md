# otf-charts

![Version: 0.3.11](https://img.shields.io/badge/Version-0.3.11-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: 0.2.3](https://img.shields.io/badge/AppVersion-0.2.3-informational?style=flat-square)

[OTF](https://github.com/leg100/otf) Helm charts.

## Instructions

Add the helm repository:

```bash
helm repo add otf https://leg100.github.io/otf-charts
```

To install the chart you need at the very minimum:

* A PostgreSQL database up and running.
* A hex-coded 16 byte [secret](https://docs.otf.ninja/latest/config/flags#-secret).
* Either setup an [identity provider](https://docs.otf.ninja/latest/auth#identity-providers) or set a [site admin token](https://docs.otf.ninja/latest/auth#site-admin).

For example, if a PostgreSQL server is accessible via the hostname `postgres`, has a database named `otf` accessible to a user with username `postgres` and password `postgres`:

```
helm install otf otf/otf --set secret=2876cb147697052eec5b3cdb56211681 --set site-token=my-token --set database=postgres://postgres:postgres@postgres/otf
```

Alternatively, you can use the [test-values.yaml](./charts/otf/test-values.yaml) from this repo:

```
helm install otf otf/otf -f ./charts/otf/test-values.yaml
```

This will:

* Install PostgreSQL on the cluster
* Set secret
* Set a site token

Note: you should only use this for testing purposes.

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity | object | `{}` |  |
| caCerts.enabled | bool | `false` | Mount a secret containing CA certificates and make them available to both terraform and otfd, allowing them to communicate with API endpoints that use custom CA certificates. |
| caCerts.secretItems | list | `[]` | Specify individual items in secret containing CA certificates. Use the [KeyToPath](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.28/#keytopath-v1-core) schema for each item. If unspecified, all items are mounted from the secret. |
| caCerts.secretName | string | `"certs"` | Name of secret containing the CA certificates to mount. |
| database | string | `""` | Postgres connection string |
| databasePasswordFromSecret | object | `nil` | Source database password from a secret |
| databaseUsernameFromSecret | object | `nil` | Source database username from a secret |
| envsFromSecret | string | `""` | Environment variables to be passed to the deployment from the named kubernetes secret. |
| extraEnvs | list | `[]` | Extra environment variables to be passed to the deployment. |
| fullnameOverride | string | `""` |  |
| github.clientID | string | `""` | Github OAuth client ID. See [docs](https://docs.otf.ninja/latest/config/flags/#-github-client-id). |
| github.clientSecret | string | `""` | Github OAuth client secret. See [docs](https://docs.otf.ninja/latest/config/flags/#-github-client-secret). |
| github.hostname | string | `"github.com"` | Github hostname to use for all interactions with Github. |
| gitlab.clientID | string | `""` | Gitlab OAuth client ID. See [docs](https://docs.otf.ninja/latest/config/flags/#-gitlab-client-id). |
| gitlab.clientSecret | string | `""` | Gitlab OAuth client secret. See [docs](https://docs.otf.ninja/latest/config/flags/#-gitlab-client-secret). |
| gitlab.hostname | string | `"gitlab.com"` | Gitlab hostname to use for all interactions with Gitlab. |
| google.audience | string | `""` | The Google JWT audience claim for validation. Validation is skipped if empty. See [docs](https://docs.otf.ninja/latest/config/flags/#-google-jwt-audience). |
| hostname | string | `""` | Set client-accessible hostname. See [docs](https://docs.otf.ninja/latest/config/flags/#-hostname). |
| image.pullPolicy | string | `"IfNotPresent"` |  |
| image.repository | string | `"leg100/otfd"` |  |
| image.tag | string | `""` |  |
| imagePullSecrets | list | `[]` |  |
| ingress.annotations | object | `{}` |  |
| ingress.enabled | bool | `false` |  |
| ingress.hosts | list | `[]` |  |
| ingress.path | string | `"/"` |  |
| ingress.pathType | string | `"Prefix"` |  |
| ingress.tls | list | `[]` |  |
| logging.format | string | `"default"` | Logging format: default, text, or json. See [docs](https://docs.otf.ninja/latest/config/flags/#-log-format) |
| logging.http | bool | `false` | Log http requests. |
| logging.verbosity | int | `0` | Logging verbosity, the higher the number the more verbose the logs. See [docs](https://docs.otf.ninja/latest/config/flags/#-v). |
| maxConfigSize | string | `""` | Max config upload size in bytes. See [docs](https://docs.otf.ninja/latest/config/flags/#-max-config-size). |
| nameOverride | string | `""` |  |
| no_proxy | string | `nil` | Specify hosts for which outbound connections should not use the proxy. |
| nodeSelector | object | `{}` |  |
| oidc.clientID | string | `""` | OIDC client ID. See [docs](https://docs.otf.ninja/latest/auth/providers/oidc/). |
| oidc.clientSecretFromSecret | object | `nil` | Source OIDC client secret from a k8s secret. See [docs](https://docs.otf.ninja/latest/auth/providers/oidc/). |
| oidc.issuerURL | string | `""` | OIDC issuer URL. See [docs](https://docs.otf.ninja/latest/auth/providers/oidc/). |
| oidc.name | string | `""` | OIDC provider name. See [docs](https://docs.otf.ninja/latest/auth/providers/oidc/). |
| oidc.scopes | list | `[]` | Override OIDC scopes. See [docs](https://docs.otf.ninja/latest/auth/providers/oidc/). |
| oidc.usernameClaim | string | `""` | Override OIDC claim used for username. See [docs](https://docs.otf.ninja/latest/auth/providers/oidc/). |
| podAnnotations | object | `{}` | Add annotations to otfd pod |
| podSecurityContext | object | `{}` | Set security context for otfd pod |
| postgres.enabled | bool | `false` | Install postgres chart dependency. |
| proxy | string | `nil` | Specify an http(s) proxy for outbound connections. |
| replicaCount | int | `1` | Number of otfd nodes in cluster |
| resources | object | `{}` |  |
| sandbox | bool | `false` | Enable sandboxing of terraform apply - note, this will run pods as privileged |
| secret | string | `""` | Cryptographic secret. Must be a hex-encoded 16-byte string. See [docs](https://docs.otf.ninja/latest/config/flags/#-secret). |
| secretFromSecret | object | `{}` | Source cryptographic secret from a kubernetes secret instead. |
| service.annotations | object | `{}` |  |
| service.port | int | `80` | Service port for otf |
| service.type | string | `"ClusterIP"` | Service type for otf |
| serviceAccount.annotations | object | `{}` | Annotations to add to the service account |
| serviceAccount.create | bool | `true` | Specifies whether a service account should be created |
| serviceAccount.name | string | `""` | The name of the service account to use. If not set and create is true, a name is generated using the fullname template |
| serviceMonitor | object | `{"enabled":false}` | Collect prometheus metrics |
| siteAdmins | list | `[]` | Site admins - list of user accounts promoted to site admin. See [docs](https://docs.otf.ninja/latest/config/flags/#-site-admins). |
| siteToken | string | `""` | Site admin token - empty string disables the site admin account. See [docs](https://docs.otf.ninja/latest/config/flags/#-site-token). |
| tolerations | list | `[]` |  |

## Development

If you make any changes to the chart you need to bump its version. You can either do that by hand in `Chart.yaml`, or using `make`:

```bash
# requires `yq`
make bump
```

To generate the `README.md` from its template, `README.md.gotmpl`:

```bash
# requires `helm-docs`
make readme
```

Any changes to the version or to the `values.yaml` file are automatically reflected in the generated `README.md`.

To lint the chart:

```bash
make lint
```

To deploy the chart to a cluster to the namespace `otf-test` with pre-configured defaults along with PostgreSQL:

```bash
make deploy
```

To test the chart (assumes release is named `otf`):

```bash
make test
```

And to lint, deploy, and then test the chart:

```bash
make all
```
