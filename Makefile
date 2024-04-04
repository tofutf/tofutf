VERSION ?= $(shell git describe --tags --dirty --always)

# Provide some sane defaults for connecting to postgres.
PGPASSWORD ?= $(shell kubectl get secrets postgres-postgresql -oyaml | yq '.data["password"]' -r | base64 -d)
PGUSER ?= tofutf
DBSTRING ?= postgres://$(PGUSER):$(PGPASSWORD)@localhost:5432/postgres

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

.PHONY: dev
dev: 
	./hack/dev.sh

.PHONY: go-tfe-tests
go-tfe-tests: image compose-up
	./hack/go-tfe-tests.bash

.PHONY: watch
watch: tailwind-watch

.PHONY: tailwind
tailwind:
	pnpm exec tailwindcss -i ./internal/http/html/static/css/input.css -o ./internal/http/html/static/css/output.css

.PHONY: tailwind-watch
tailwind-watch:
	+pnpm exec tailwindcss -i ./internal/http/html/static/css/input.css -o ./internal/http/html/static/css/output.css --watch

.PHONY: test
test:
	go test ./...

.PHONY: k3d-up
k3d-up:
	k3d cluster create --config=./hack/k3d.yaml

.PHONY: k3d-down
k3d-down:
	k3d cluster delete tofutf

# Run staticcheck metalinter recursively against code
.PHONY: lint
lint:
	go list ./... | grep -v pggen | xargs staticcheck

# Run go fmt against code
.PHONY: fmt
fmt:
	go fmt ./...

# Run go vet against code
.PHONY: vet
vet:
	go vet ./...

# Install sql code generator
.PHONY: install-pggen
install-pggen:
	@sh -c "which pggen > /dev/null || go install github.com/leg100/pggen/cmd/pggen@latest"

# Generate sql code
.PHONY: sql
sql: install-pggen
	pggen gen go \
		--postgres-connection $(DBSTRING) \
		--query-glob 'internal/sql/queries/*.sql' \
		--output-dir ./internal/sql/pggen \
		--go-type 'text=github.com/jackc/pgtype.Text' \
		--go-type 'int4=github.com/jackc/pgtype.Int4' \
		--go-type 'int8=github.com/jackc/pgtype.Int8' \
		--go-type 'bool=github.com/jackc/pgtype.Bool' \
		--go-type 'bytea=[]byte' \
		--acronym url \
		--acronym cli \
		--acronym sha \
		--acronym json \
		--acronym vcs \
		--acronym html \
		--acronym http \
		--acronym tls \
		--acronym sso \
		--acronym hcl \
		--acronym ip
	goimports -w ./internal/sql/pggen
	go fmt ./internal/sql/pggen

# Install DB migration tool
.PHONY: install-goose
install-goose:
	@sh -c "which goose > /dev/null || go install github.com/pressly/goose/v3/cmd/goose@latest"

# Migrate SQL schema to latest version
.PHONY: migrate
migrate: install-goose
	GOOSE_DBSTRING=$(DBSTRING) GOOSE_DRIVER=postgres goose -dir ./internal/sql/migrations up

# Redo SQL schema migration
.PHONY: migrate-redo
migrate-redo: install-goose
	GOOSE_DBSTRING=$(DBSTRING) GOOSE_DRIVER=postgres goose -dir ./internal/sql/migrations redo

# Rollback SQL schema by one version
.PHONY: migrate-rollback
migrate-rollback: install-goose
	GOOSE_DBSTRING=$(DBSTRING) GOOSE_DRIVER=postgres goose -dir ./internal/sql/migrations down

# Get SQL schema migration status
.PHONY: migrate-status
migrate-status: install-goose
	GOOSE_DBSTRING=$(DBSTRING) GOOSE_DRIVER=postgres goose -dir ./internal/sql/migrations status

.PHONY: doc-screenshots
doc-screenshots: # update documentation screenshots
	OTF_DOC_SCREENSHOTS=true go test ./internal/integration/... -count 1

# Generate path helpers
.PHONY: paths
paths:
	go generate ./internal/http/html/paths
	goimports -w ./internal/http/html/paths

# Re-generate RBAC action strings
.PHONY: actions
actions:
	stringer -type Action ./internal/rbac

# Install staticcheck linter
.PHONY: install-linter
install-linter:
	go install honnef.co/go/tools/cmd/staticcheck@latest

.PHONY: publish
publish:
	KO_DOCKER_REPO=ghcr.io/tofutf/tofutf/ ko resolve --base-import-paths -t $(VERSION) -f ./charts/tofutf/values.yaml.tmpl > ./charts/tofutf/values.yaml 
	yq 'select(di == 0) | .image.tag = .image.override | del(.image.override) | del(.agent) | .image.tag |= sub("ghcr.io/tofutf/tofutf/tofutfd:", "")' -i ./charts/tofutf/values.yaml
	yq ".version=\"$(VERSION)\" | .appVersion=\"$(VERSION)\"" -i ./charts/tofutf/Chart.yaml
	helm package ./charts/tofutf --app-version $(VERSION) --version $(VERSION) --destination=./hack/charts/
	helm push ./hack/charts/tofutf-$(VERSION).tgz oci://ghcr.io/tofutf/tofutf/charts

publish-dev:
	KO_DOCKER_REPO=ghcr.io/tofutf/tofutf/ ko build --base-import-paths -t dev ./cmd/tofutfd
