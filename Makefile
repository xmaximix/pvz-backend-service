.PHONY: prereq generate build test test-integ

generate: prereq
	@oapi-codegen -generate types,gin \
		-package api \
		-o internal/api/types/types.gen.go \
		api/openapi.yaml
	@protoc \
    		--go_out=. --go_opt=paths=source_relative \
    		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
    		api/pvz/v1/pvz.proto

build:
	go build -o bin/pvz-server cmd/server/main.go

test:
	go test ./internal/auth \
            ./internal/logger \
            ./internal/repo \
            ./internal/service \
            ./internal/api \
            -cover

test-integ:
	go test ./tests/integration