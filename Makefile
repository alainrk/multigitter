PHONY: run test

run:
	go run cmd/multigitter/main.go

test:
	go test -v -cover -race ./...
