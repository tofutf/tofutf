# Development

Code contributions are welcome.

## Setup your machine

* Clone the repo:

```
git clone git@github.com:tofutf/tofutf.git
```

The repository uses [devcontainers](https://containers.dev/) in order to provide a consistent development environment. If you have not developed using a devcontainer before, then [these](https://code.visualstudio.com/docs/devcontainers/containers) docs should be quite useful. 

## Documentation

The documentation pages are maintained in the `./docs` directory of the repository. 

```
make serve-docs
```

That builds and runs the documentation site on your workstation at `http://localhost:9999`. Any changes you make to the documentation are reflected in real-time in the browser.

Before running the make task you'll need to run `pnpm install`

Screenshots in the documentation are largely automated. The browser-based integration tests produce screenshots at various steps. If the environment variable `TOFUTF_DOC_SCREENSHOTS=true` is present then such a test also writes the screenshot into the documentation directory. The following make task runs the tests along with the aforementioned environment variable:

```
make doc-screenshots
```

## SQL migrations

The database schema is migrated using [goose](https://github.com/pressly/goose). The SQL migration files are kept in the repo in `./sql/migrations`. Upon startup `tofutfd` automatically migrates the DB to the latest version.

If you're developing a SQL migration you may want to migrate the database manually. Use the `make` tasks to assist you:

* `make migrate`
* `make migrate-redo`
* `make migrate-rollback`
* `make migrate-status`

## SQL queries

SQL queries are handwritten in `./internal/sql/queries` and turned into Go using [pggen](https://github.com/jschaf/pggen).

After you make changes to the queries run the following make task to invoke `pggen`:

* `make sql`

## HTML path helpers

Rails-style path helpers are generated using `go generate`. The path specifications are maintained in `./internal/http/html/paths/gen.go`. After making changes to the specs run the following make task to generate the helpers:

* `make paths`

## Web development

If you're making changes to web templates then you may want to enable [developer mode](config/flags.md/#-dev-mode). Once enabled you will be able to see changes without restarting `tofutfd`.

tofutf uses [Tailwind CSS](https://tailwindcss.com/) to generate CSS classes. Run the following make task to generate the CSS:

* `make tailwind`

!!! note
    To install tailwind first ensure you've installed `pnpm` and then run `pnpm install`

