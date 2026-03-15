VERSION := 0.3.0
LDFLAGS := -s -w -X main.version=$(VERSION)
PLATFORMS := linux-amd64 linux-arm64 darwin-amd64 darwin-arm64 windows-amd64

.PHONY: build test lint bench coverage install clean build-all package $(addprefix build-,$(PLATFORMS))

build:
	CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o bin/veilkey-cli .

test:
	go test -v -race ./...

lint:
	golangci-lint run ./...

bench:
	go test -bench=. -benchmem ./...

coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out
	@echo ""
	@echo "HTML report: go tool cover -html=coverage.out"

install: build
	cp bin/veilkey-cli /usr/local/bin/veilkey-cli

clean:
	rm -rf bin/ dist/ coverage.out

build-linux-amd64:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o bin/veilkey-cli-linux-amd64 .

build-linux-arm64:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o bin/veilkey-cli-linux-arm64 .

build-darwin-amd64:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o bin/veilkey-cli-darwin-amd64 .

build-darwin-arm64:
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o bin/veilkey-cli-darwin-arm64 .

build-windows-amd64:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o bin/veilkey-cli-windows-amd64.exe .

build-all: $(addprefix build-,$(PLATFORMS))

package: build-all
	@mkdir -p dist
	@for p in linux-amd64 linux-arm64 darwin-amd64 darwin-arm64; do \
		echo "Packaging veilkey-cli-$$p..."; \
		tar -czf dist/veilkey-cli-$(VERSION)-$$p.tar.gz \
			-C bin veilkey-cli-$$p \
			-C ../scripts vk vk_wrap; \
	done
	@echo "Packaging veilkey-cli-windows-amd64..."
	@cd bin && zip -q ../dist/veilkey-cli-$(VERSION)-windows-amd64.zip veilkey-cli-windows-amd64.exe
	@echo "Done: dist/"
	@ls -lh dist/
