TAG = $$(git rev-parse --short HEAD)
IMG ?= ghcr.io/legobeat/git-auth-proxy:$(TAG)

assets:
	draw.io -b 10 -x -f png -p 0 -o assets/architecture.png assets/diagram.drawio
.PHONY: assets

build:
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o git-auth-proxy .

lint:
	golangci-lint run ./...

fmt:
	go fmt ./...

vet:
	go vet ./...

test: fmt vet
	go test --cover ./...

run: fmt vet
	go run main.go

docker-build:
	docker build -t ${IMG} .

e2e: docker-build kind-load
	./e2e/e2e.sh $(TAG)
.PHONY: e2e


