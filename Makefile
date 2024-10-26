PROJECT_NAME := "keyhouse"
GIT_COMMIT   := $(shell git describe --dirty=-unsupported --always --tags --long || echo pre-commit)
BUILD_NUMBER ?= 0

DOCKER_REPOSITORY := "docker.io/skriptvalley"
IMAGE_NAME        := $(DOCKER_REPOSITORY)/$(PROJECT_NAME)
IMAGE_VERSION     := $(GIT_COMMIT)-$(BUILD_NUMBER)
IMAGE_TAG         := "$(IMAGE_NAME):$(IMAGE_VERSION)"

# go targets
.PHONY: go-mod
go-mod:
	@go mod tidy
	@go mod vendor

# proto targets
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
.PHONE: docker-dev docker-dev-build docker-dev-push docker docker-build docker-push

docker-dev: docker-dev-build docker-dev-push

docker-dev-build:
	@docker build --platform linux/amd64 -f docker/Dockerfile -t $(IMAGE_TAG)-dev .

docker-dev-push:
	@docker push $(IMAGE_TAG)-dev

docker: docker-build docker-push

docker-build:
	@docker build --platform linux/amd64 -f docker/Dockerfile -t $(IMAGE_TAG)-dev .

docker-push:
	@docker push $(IMAGE_TAG)
