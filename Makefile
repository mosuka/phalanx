GOOS ?=
GOARCH ?=
GO111MODULE ?= on
CGO_ENABLED ?= 0
CGO_CFLAGS ?=
CGO_LDFLAGS ?=
BUILD_TAGS ?=
VERSION ?=
BIN_EXT ?=
DOCKER_REPOSITORY ?= mosuka

PACKAGES = $(shell $(GO) list -tags="$(BUILD_TAGS)" ./... | grep -v '/vendor/')

PROTOBUFS = $(shell find . -name '*.proto' -print0 | xargs -0 -n1 dirname | sort | uniq | grep -v /vendor/)
TARGET_PACKAGES = $(shell find $(CURDIR) -name 'main.go' -print0 | xargs -0 -n1 dirname | sort | uniq | grep -v /vendor/)

ifeq ($(GOOS),)
  GOOS = $(shell go version | awk -F ' ' '{print $$NF}' | awk -F '/' '{print $$1}')
endif

ifeq ($(GOARCH),)
  GOARCH = $(shell go version | awk -F ' ' '{print $$NF}' | awk -F '/' '{print $$2}')
endif

ifeq ($(VERSION),)
  VERSION = latest
endif
LDFLAGS = -ldflags "-s -w -X \"github.com/mosuka/phalanx/version.Version=$(VERSION)\""

ifeq ($(GOOS),windows)
  BIN_EXT = .exe
endif

BUILD_FLAGS := GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=$(CGO_ENABLED) CGO_CFLAGS=$(CGO_CFLAGS) CGO_LDFLAGS=$(CGO_LDFLAGS) GO111MODULE=$(GO111MODULE)

GO := $(BUILD_FLAGS) go

.DEFAULT_GOAL := build

.PHONY: show-env
show-env:
	@echo ">> show env"
	@echo "   GOOS              = $(GOOS)"
	@echo "   GOARCH            = $(GOARCH)"
	@echo "   GO111MODULE       = $(GO111MODULE)"
	@echo "   CGO_ENABLED       = $(CGO_ENABLED)"
	@echo "   CGO_CFLAGS        = $(CGO_CFLAGS)"
	@echo "   CGO_LDFLAGS       = $(CGO_LDFLAGS)"
	@echo "   BUILD_TAGS        = $(BUILD_TAGS)"
	@echo "   VERSION           = $(VERSION)"
	@echo "   BIN_EXT           = $(BIN_EXT)"
	@echo "   LDFLAGS           = $(LDFLAGS)"
	@echo "   PACKAGES          = $(PACKAGES)"
	@echo "   PROTOBUFS         = $(PROTOBUFS)"
	@echo "   TARGET_PACKAGES   = $(TARGET_PACKAGES)"

.PHONY: protoc
protoc: show-env
	@echo ">> generating proto3 code"
	for proto_dir in $(PROTOBUFS); do echo $$proto_dir; protoc --proto_path=. --proto_path=$$proto_dir --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative $$proto_dir/*.proto || exit 1; done

.PHONY: fmt
fmt: show-env
	@echo ">> formatting code"
	$(GO) fmt $(PACKAGES)

.PHONY: mock
mock:
	@echo ">> generating mocks"
	mockgen -source=./metastore/storage.go -destination=./mock/metastore/storage.go

.PHONY: test
test: show-env
	@echo ">> testing all packages"
	$(GO) clean -testcache
	$(GO) test -v -tags="$(BUILD_TAGS)" $(PACKAGES)

.PHONY: clean
clean:
	@echo ">> cleaning repository"
	$(GO) clean

.PHONY: build
build: show-env
	@echo ">> building binaries"
	$(GO) build -tags="$(BUILD_TAGS)" $(LDFLAGS) -o bin/phalanx

.PHONY: docs
docs:
	@echo ">> building document"
	gitbook install ./docs_md
	gitbook build ./docs_md
	cp ./docs_md/README.md ./README.md
	rm -rf docs
	mv ./docs_md/_book docs

.PHONY: tag
tag: show-env
	@echo ">> tagging github"
ifeq ($(VERSION),$(filter $(VERSION),latest master ""))
	@echo "please specify VERSION"
else
	git tag -a $(VERSION) -m "Release $(VERSION)"
	git push origin $(VERSION)
endif

.PHONY: docker-build
docker-build: show-env
	@echo ">> building docker container image"
	docker build -t $(DOCKER_REPOSITORY)/phalanx:latest --build-arg VERSION=$(VERSION) .
	docker tag $(DOCKER_REPOSITORY)/phalanx:latest $(DOCKER_REPOSITORY)/phalanx:$(VERSION)

.PHONY: docker-push
docker-push: show-env
	@echo ">> pushing docker container image"
	docker push $(DOCKER_REPOSITORY)/phalanx:latest
	docker push $(DOCKER_REPOSITORY)/phalanx:$(VERSION)

.PHONY: docker-clean
docker-clean:
	docker rmi -f $(shell docker images --filter "dangling=true" -q --no-trunc)

.PHONY: cert
cert:
	@echo ">> generating certification"
	openssl req -x509 -nodes -newkey rsa:4096 -keyout ./examples/phalanx-key.pem -out ./examples/phalanx-cert.pem -days 365 -subj '/CN=localhost'
