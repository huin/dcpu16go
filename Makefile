all: dis

examples: \
    examples/test.bin \
    examples/test.dasm16

test:
	@go test ./...

fmt:
	@go fmt ./...

dis:
	@go build dis

%.bin: %.hex
	xxd -r $< $@

%.dasm16: %.bin
	"./dis" $< $@

.PHONY: all examples fmt test
