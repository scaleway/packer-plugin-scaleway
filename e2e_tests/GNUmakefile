build-binary:
	make -C .. build

link-binary:
	ln -fs ../packer-plugin-scaleway

test-simple: build-binary link-binary
	packer build ./simple/build_scaleway.pkr.hcl
	go run ./simple/test.go

test:
	./test.sh
