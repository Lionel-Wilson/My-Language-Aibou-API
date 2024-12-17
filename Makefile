deps:
	go install go.uber.org/mock/mockgen@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install mvdan.cc/gofumpt@latest
.PHONY: deps

mock:
	go generate ./...
.PHONY: mock

lint:
	go mod tidy
	go vet ./...
	gci write -s standard -s default -s "prefix(github.com/Lionel-Wilson/My-Language-Aibou-API)" .
	gofumpt -l -w .
	wsl -fix ./... 2> /dev/null || true
	golangci-lint run $(p)
	go fmt ./...
.PHONY: lint