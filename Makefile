FILENAME=logagent
fmt:
	go fmt ./...
build:fmt
	CGO_ENABLED=0 go build -a -ldflags '-extldflags "-static"' -o $(FILENAME) . && upx $(FILENAME)
.PHONY: fmt build