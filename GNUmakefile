NAME=scaleway
BINARY=packer-plugin-${NAME}
PLUGIN_FQN="$(shell grep -E '^module' <go.mod | sed -E 's/module *//')"

COUNT?=1
E2E_TEST?=$(shell go list ./internal/tests)
TEST?=$(filter-out $(E2E_TEST),$(shell go list ./...))

.PHONY: dev

build:
	@go build -o ${BINARY}

dev:
	@go build -ldflags="-X '${PLUGIN_FQN}/version.VersionPrerelease=dev'" -o '${BINARY}'
	packer plugins install --path ${BINARY} "$(shell echo "${PLUGIN_FQN}" | sed 's/packer-plugin-//')"

test:
	@go test -count $(COUNT) $(TEST) -timeout=3m

e2e_test:
	make -C e2e_tests test

plugin-check: build
	@go tool packer-sdc plugin-check ${BINARY}

testacc: dev
	@PACKER_ACC=1 go test -count $(COUNT) -v $(TEST) -timeout=120m

generate: install-packer-sdc
	@go generate ./...
	@if [ -d ".docs" ]; then rm -r ".docs"; fi
	@go tool packer-sdc renderdocs -src "docs" -partials docs-partials/ -dst ".docs/"
	@./.web-docs/scripts/compile-to-webdocs.sh "." ".docs" ".web-docs" "scaleway"
	@rm -r ".docs"
	# checkout the .docs folder for a preview of the docs

install-plugin:
	@packer plugins install github.com/scaleway/scaleway
