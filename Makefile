PROJECT_NAME := "keyhouse"
GIT_COMMIT   := $(shell git describe --dirty=-unsupported --always --tags --long || echo pre-commit)
BUILD_NUMBER ?= 0

DOCKER_REPOSITORY := "docker.io/skriptvalley"
IMAGE_NAME        := $(DOCKER_REPOSITORY)/$(PROJECT_NAME)
IMAGE_VERSION     := $(GIT_COMMIT)-$(BUILD_NUMBER)
IMAGE_TAG         := "$(IMAGE_NAME):$(IMAGE_VERSION)"

# app targets
.PHONY: app-docker-build app-run
app-docker-build:
	@docker build -f docker/Dockerfile.dev -t $(IMAGE_TAG)-local .

app-run: 
	@if ! docker image inspect $(IMAGE_TAG)-local > /dev/null 2>&1; then \
		echo "Image not found. Building..."; \
		$(MAKE) app-docker-build; \
	fi
	@docker run --name $(PROJECT_NAME) -d -p 30100:8080 -p 30101:8081 $(IMAGE_TAG)-local \
	/app/keyhouse \
	--app-version=$(IMAGE_VERSION) \
	--log-level=debug \
	--redis-host=localhost \
	--redis-port=6379 \
	--redis-password=admin

# clean docker container for app if running and remove -local image
app-docker-clean:
	@docker rm -f $(PROJECT_NAME) 2>/dev/null || true
	@docker rmi $(IMAGE_TAG)-local 2>/dev/null || true

# go targets
.PHONY: go-mod
go-mod:
	@go mod tidy
	@go mod vendor

# proto targets
.PHONY: proto-clean
proto-clean:
	@rm -rf pkg/pb/*

.PHONY: proto-gen-app
proto-gen-app:
	@mkdir -p pkg/pb/app
	@mkdir -p pkg/pb/docs
	@protoc -I ./proto --grpc-gateway_out ./pkg/pb/app \
    --grpc-gateway_opt paths=source_relative \
	--go_out=./pkg/pb --go-grpc_out=./pkg/pb \
	--openapiv2_out=./pkg/pb/docs \
	./proto/*.proto

# docker targets
.PHONE: docker-dev docker-dev-build docker-dev-push docker docker-build docker-push docker-clean-all

docker-dev: docker-dev-build docker-dev-push

docker-dev-build:
	@docker build --platform linux/amd64 -f docker/Dockerfile -t $(IMAGE_TAG)-dev .

docker-dev-push:
	@docker push $(IMAGE_TAG)-dev

docker-dev-clean:
	@docker rmi $(IMAGE_TAG)-dev

docker: docker-build docker-push

docker-build:
	@docker build --platform linux/amd64 -f docker/Dockerfile -t $(IMAGE_TAG)-dev .

docker-push:
	@docker push $(IMAGE_TAG)

docker-clean-all:
	@docker rmi -f $(shell docker images skriptvalley/keyhouse -q)

# docker-compose targets
.PHONY: docker-compose-up docker-compose-down
docker-compose-up:
	@awk -v tag="$(IMAGE_VERSION)-local" '/^IMAGE_VERSION=/ {print "IMAGE_VERSION=" tag; next} {print}' docker/.env > docker/.env.tmp
	@mv docker/.env.tmp docker/.env
	@docker compose -f docker/docker-compose.yaml --env-file docker/.env up -d

docker-compose-down:
	@docker compose -f docker/docker-compose.yaml down
