build-binary:
	make -C .. build

link-binary:
	ln -fs ../packer-plugin-scaleway

test: build-binary link-binary
	./test.sh
