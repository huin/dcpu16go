all: bin/dis

examples: \
    examples/test.bin \
    examples/test.dasm16

test:
	@go test ./...

fmt:
	@go fmt ./...

bin/dis:
	@go build -o bin/dis cmd/dis/*.go

%.bin: %.hex
	xxd -r $< $@

%.dasm16: %.bin
	"./dis" $< $@

.PHONY: all examples fmt test
