build-binary:
	make -C .. build

install-plugin:
	packer plugins install -path ../packer-plugin-scaleway "github.com/scaleway/scaleway"

test: build-binary install-plugin
	./test.sh
