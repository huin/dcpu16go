all: \
    bin/asm \
    bin/dis \

clean:
	rm -f examples/test.{bin,dasm16}
	rm -f bin/dis

examples: \
    examples/test.bin \
    examples/test.dasm16 \

test:
	@go test ./...

fmt:
	@go fmt ./...

bin/%:
	@go build -o bin/$(notdir $@) cmd/$(notdir $@)/*.go

%.bin: %.hex
	xxd -r $< $@

%.dasm16: %.bin
	bin/dis $< $@

.PHONY: all clean examples fmt test
