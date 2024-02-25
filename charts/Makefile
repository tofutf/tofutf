.PHONY: all
all: readme lint deploy test

.PHONY: lint
lint:
	helm lint ./charts/otf

.PHONY: deploy
deploy:
	helm upgrade -i --create-namespace -n otf-test -f ./charts/otf/test-values.yaml otf ./charts/otf --wait

.PHONY: test
test: lint deploy
	helm test -n otf-test otf

.PHONY: readme
readme:
	helm-docs -c ./charts/otf -t ../../README.md.gotmpl -o ../../README.md

# bump patch in chart version
.PHONY: bump
bump:
	yq -i '.version |= (split(".") | .[-1] |= ((. tag = "!!int") + 1) | join("."))' ./charts/otf/Chart.yaml
