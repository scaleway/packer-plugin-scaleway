build-binary:
	make -C .. build

install-plugin:
	packer plugins install -path ../packer-plugin-scaleway "github.com/scaleway/scaleway"

# The plugin is built and installed from the current code by the tests themselves (internal/tests/main_test.go)
test:
	./test.sh
