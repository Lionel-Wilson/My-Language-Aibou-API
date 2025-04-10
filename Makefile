deps:
	@echo "Installing dependencies..."
	go mod tidy
	go mod download
.PHONY: deps

mock:
	go generate ./...
.PHONY: mock

entity:
	@sqlboiler psql -c ./sqlboiler.toml
.PHONY: entity

#DSN="postgres://mylanguageaibouuser:MyAibou25@db:5432/my-language-aibou-db?sslmode=disable"
DSN="postgres://mylanguageaibouuser:MyAibou25@localhost:5432/my-language-aibou-db?sslmode=disable"

migrate-up:
	@goose -dir ./migrations postgres ${DSN} up
.PHONY: migrate-up

migrate-down:
	@goose -dir ./migrations postgres ${DSN} down
.PHONY: migrate-down

migrate-create:
	@cd ./migrations && goose create update_payment_transactions_payment_intent sql
.PHONY: migrate-create

lint:
	go mod tidy
	go vet ./...
	gci write -s standard -s default -s "prefix(github.com/Lionel-Wilson/My-Language-Aibou-API)" .
	gofumpt -l -w .
	wsl -fix ./... 2> /dev/null || true
	golangci-lint run $(p)
	go fmt ./...
.PHONY: lint

build:
	docker-compose up --build
.PHONY: build

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
