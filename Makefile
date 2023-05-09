.PHONY: default examples test

default:
	./validate.sh
	go build github.com/uoregon-libraries/gopkg/...

examples:
	go build -o ./bin/bagit ./examples/bagit
	go build -o ./bin/copydir ./examples/copydir
	go build -o ./bin/manifest ./examples/manifest
	go build -o ./bin/syncdir ./examples/syncdir

test:
	go test ./...
