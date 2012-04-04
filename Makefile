all: dis

test:
	go test ./...

dis:
	go build dis

.PHONY: all test
