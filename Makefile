PROJECT_NAME := "keyhouse"
GIT_COMMIT   := $(shell git describe --dirty=-unsupported --always --tags --long || echo pre-commit)
BUILD_NUMBER ?= 0

IMAGE_NAME    := $(PROJECT_NAME)
IMAGE_VERSION := $(GIT_COMMIT)-$(BUILD_NUMBER)
IMAGE_TAG     := "$(IMAGE_NAME):$(IMAGE_VERSION)"

.PHONY: deps
deps: go-mod

.PHONY: go-mod
go-mod:
	@go mod tidy
	@go mod vendor

.PHONY: clean-proto
clean-proto:
	@rm -rf pkg/pb/*

.PHONY:protogen-backend
protogen-backend:
	@mkdir -p pkg/pb/backend
	@mkdir -p pkg/pb/docs
	@protoc -I ./proto --grpc-gateway_out ./pkg/pb/backend \
    --grpc-gateway_opt paths=source_relative \
	--go_out=./pkg/pb --go-grpc_out=./pkg/pb \
	--openapiv2_out=./pkg/pb/docs \
	./proto/*.proto

# docker targets
docker-build:
	@docker build --platform linux/amd64 -f docker/Dockerfile -t $(IMAGE_TAG) .
