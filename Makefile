.PHONY: build test test-compat test-external clean

build:
	go build -o shdoc-ng .

test:
	go test ./...

test-compat:
	SHDOC_CMD="gawk -f ./shdoc-awk/shdoc" go test -v -run TestExternal/compat

test-external:
	SHDOC_CMD="gawk -f ./shdoc-awk/shdoc" go test -v -run TestExternal

clean:
	rm -f shdoc-ng
