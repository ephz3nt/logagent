fmt:
	go fmt ./...
build:fmt
	CGO_ENABLED=0 go build -a -ldflags '-extldflags "-static"' .
.PHONY: fmt build