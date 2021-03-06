PROJECT = github.com/nj-eka/shurl
CMD := $(PROJECT)/cmd/shurl

GOOS = linux
GOARCH = amd64
CGO_ENABLED = 0

RELEASE := $(shell git tag -l | tail -1 | grep -E "v.+"|| echo devel)
COMMIT := git-$(shell git rev-parse --short HEAD)
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')

format:
	go fmt $(shell go list ./...)

check: format
	golangci-lint run -c golangci-lint.yaml

test: check
	go test ./...

.PHONY: build
build:
	mkdir -p bin
	GOOS=${GOOS} GOARCH=${GOARCH} CGO_ENABLED=${CGO_ENABLED} go build -v -x \
		-ldflags "-s -w \
		-X ${PROJECT}/app.Version=${RELEASE} \
		-X ${PROJECT}/app.Commit=${COMMIT} \
		-X ${PROJECT}/app.BuildTime=${BUILD_TIME}" \
		-o bin ${CMD}

run: build
	bin/shurl -config "config/local/config.yml"

clean:
	rm -rf bin
