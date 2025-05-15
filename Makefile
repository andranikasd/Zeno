.PHONY: lint test build container

lint:
	go fmt ./...
	golangci-lint run

test:
	go test ./... -cover

build:
	go build -o bin/zeno ./cmd/zeno

docker-container:
	@echo "Building image $(DOCKER_TAG)…"
	# if you want BuildKit enabled, make sure it's installed; otherwise set BUILDKIT=0
	DOCKER_BUILDKIT=1 docker build \
	  -f docker/Dockerfile \
	  -t $(DOCKER_TAG) \
	  .