# CLI

`tofutf` is the CLI for TOFUTF.

Download a [release](https://github.com/tofutf/tofutf/releases). Ensure you select the client component, `tofutf`. The release is a zip file. Extract the `tofutf` binary to a directory in your system PATH.

Run `tofutf` with no arguments to receive usage instructions:

```bash
Usage:
  tofutf [command]

Available Commands:
  agents        Agent management
  help          Help about any command
  organizations Organization management
  runs          Runs management
  state         State version management
  teams         Team management
  users         User account management
  workspaces    Workspace management

Flags:
      --address string   Address of tofutf server (default "localhost:8080")
  -h, --help             help for tofutf
      --token string     API authentication token

Use "tofutf [command] --help" for more information about a command.
```

Credentials are sourced from the same file the terraform CLI uses (`~/.terraform.d/credentials.tfrc.json`). To populate credentials, run:

```bash
terraform login <tofutfd_hostname>
```
