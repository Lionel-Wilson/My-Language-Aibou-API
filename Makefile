deps:
	go install go.uber.org/mock/mockgen@latest

.PHONY: deps

mock:
	go generate ./...
.PHONY: mock