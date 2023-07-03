.PHONY: all
all: clean generate tidy vet fmt lint test gosec

clean:
	$(call print-target)
	@go clean
	@rm -f coverage.*

.PHONY: generate
generate:
	$(call print-target)
	@go generate ./...

.PHONY: tidy
tidy:
	$(call print-target)
	@go mod tidy

.PHONY: vet
vet:
	$(call print-target)
	@go vet ./...

.PHONY: fmt
fmt:
	$(call print-target)
	@go fmt ./...

.PHONY: lint
lint:
	$(call print-target)
	@golangci-lint run

.PHONY: test
test:
	$(call print-target)
	@go test -race -covermode=atomic -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html

.PHONY: gosec
gosec:
	$(call print-target)
	@gosec ./...

.PHONY: build
build:
	$(call print-target)
	@go build -o bin/ ./...

define print-target
    @printf "Executing target: \033[36m$@\033[0m\n"
endef
