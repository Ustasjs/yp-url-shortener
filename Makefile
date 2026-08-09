# Загружаем переменные из .env, если файл существует
ifneq (,$(wildcard ./.env))
    include .env
    export
endif

.PHONY: run
run:
	go run ./cmd/shortener/main.go

.PHONY: test
test:
	go test ./... -v -count=1

.PHONY: lint
lint:
	golangci-lint run ./...

.PHONY: mocks
mocks:
	go run github.com/vektra/mockery/v3@v3.7.2

.PHONY: proto-tools
proto-tools:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.11
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.6.2

.PHONY: proto
proto:
	protoc -I api/proto \
		--go_out=.      --go_opt=module=Ustasjs/yp-url-shortener \
		--go-grpc_out=. --go-grpc_opt=module=Ustasjs/yp-url-shortener \
		api/proto/shortener/v1/shortener.proto