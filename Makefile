.PHONY: default examples test

default:
	./validate.sh
	go build github.com/uoregon-libraries/gopkg/...

examples:
	go build -o ./bin/bagit ./examples/bagit

test:
	go test ./...
