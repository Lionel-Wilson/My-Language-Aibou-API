deps:
	@echo "Installing dependencies..."
	go mod tidy
	go mod download
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

start:
	docker build -t my-language-aibou-api .
	docker run --name my-language-aibou-api -p 8080:8080 my-language-aibou-api
.PHONY: start

delete:
	docker rm my-language-aibou-api
.PHONY: delete

test:
	go test ./...
.PHONY: test
