# Fetch git latest tag
LATEST_GIT_TAG:=$(shell git describe --tags $(git rev-list --tags --max-count=1))
VERSION := $(shell echo $(shell git describe --tags) | sed 's/^v//')
COMMIT := $(shell git log -1 --format='%H')

ldflags = -X github.com/zenanetwork/iris/version.Name=iris \
		  -X github.com/zenanetwork/iris/version.ServerName=irisd \
		  -X github.com/zenanetwork/iris/version.ClientName=iriscli \
		  -X github.com/zenanetwork/iris/version.Version=$(VERSION) \
		  -X github.com/zenanetwork/iris/version.Commit=$(COMMIT) \
		  -X github.com/cosmos/cosmos-sdk/version.Name=iris \
		  -X github.com/cosmos/cosmos-sdk/version.ServerName=irisd \
		  -X github.com/cosmos/cosmos-sdk/version.ClientName=iriscli \
		  -X github.com/cosmos/cosmos-sdk/version.Version=$(VERSION) \
		  -X github.com/cosmos/cosmos-sdk/version.Commit=$(COMMIT)

BUILD_FLAGS := -ldflags '$(ldflags)'

clean:
	rm -rf build

tests:
	# go test  -v ./...

	go test -v ./app/ ./auth/ ./clerk/ ./sidechannel/ ./bank/ ./chainmanager/ ./topup/ ./checkpoint/ ./staking/ -cover -coverprofile=cover.out -parallel 1

# make build
build: clean
	mkdir -p build
	go build $(BUILD_FLAGS) -o build/irisd ./cmd/irisd
	go build $(BUILD_FLAGS) -o build/iriscli ./cmd/iriscli
	@echo "====================================================\n==================Build Successful==================\n===================================================="

# make install
install:
	go install $(BUILD_FLAGS) ./cmd/irisd
	go install $(BUILD_FLAGS) ./cmd/iriscli

contracts:
	abigen --abi=contracts/rootchain/rootchain.abi --pkg=rootchain --out=contracts/rootchain/rootchain.go
	abigen --abi=contracts/stakemanager/stakemanager.abi --pkg=stakemanager --out=contracts/stakemanager/stakemanager.go
	abigen --abi=contracts/slashmanager/slashmanager.abi --pkg=slashmanager --out=contracts/slashmanager/slashmanager.go
	abigen --abi=contracts/statereceiver/statereceiver.abi --pkg=statereceiver --out=contracts/statereceiver/statereceiver.go
	abigen --abi=contracts/statesender/statesender.abi --pkg=statesender --out=contracts/statesender/statesender.go
	abigen --abi=contracts/stakinginfo/stakinginfo.abi --pkg=stakinginfo --out=contracts/stakinginfo/stakinginfo.go
	abigen --abi=contracts/validatorset/validatorset.abi --pkg=validatorset --out=contracts/validatorset/validatorset.go
	abigen --abi=contracts/erc20/erc20.abi --pkg=erc20 --out=contracts/erc20/erc20.go

build-arm: clean
	mkdir -p build
	env CGO_ENABLED=1 GOOS=linux GOARCH=arm64 CC=aarch64-linux-gnu-gcc CXX=aarch64-linux-gnu-g++ go build $(BUILD_FLAGS) -o build/irisd ./cmd/irisd
	env CGO_ENABLED=1 GOOS=linux GOARCH=arm64 CC=aarch64-linux-gnu-gcc CXX=aarch64-linux-gnu-g++ go build $(BUILD_FLAGS) -o build/iriscli ./cmd/iriscli
	@echo "====================================================\n==================Build Successful==================\n===================================================="

#
# Code quality
#

LINT_COMMAND := $(shell command -v golangci-lint 2> /dev/null)
lint:
ifndef LINT_COMMAND
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.63.4
endif
	golangci-lint run --config ./.golangci.yml

.PHONY: vulncheck

vulncheck:
	@go run golang.org/x/vuln/cmd/govulncheck@latest ./...

#
# docker commands
#

build-docker:
	@echo Fetching latest tag: $(LATEST_GIT_TAG)
	git checkout $(LATEST_GIT_TAG)
	docker build -t "maticnetwork/iris:$(LATEST_GIT_TAG)" -f docker/Dockerfile .

push-docker:
	@echo Pushing docker tag image: $(LATEST_GIT_TAG)
	docker push "maticnetwork/iris:$(LATEST_GIT_TAG)"

build-docker-develop:
	docker build -t "maticnetwork/iris:develop" -f docker/Dockerfile.develop .

.PHONY: contracts build

PACKAGE_NAME          := github.com/maticnetwork/iris
GOLANG_CROSS_VERSION  ?= v1.22.1

.PHONY: release-dry-run
release-dry-run:
	@docker run \
		--platform linux/amd64 \
		--rm \
		--privileged \
		-e CGO_ENABLED=1 \
		-e CGO_CFLAGS=-Wno-unused-function \
		-e GITHUB_TOKEN \
		-e DOCKER_USERNAME \
		-e DOCKER_PASSWORD \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-v `pwd`:/go/src/$(PACKAGE_NAME) \
		-w /go/src/$(PACKAGE_NAME) \
		goreleaser/goreleaser-cross:${GOLANG_CROSS_VERSION} \
		--rm-dist --skip-validate --skip-publish

.PHONY: release
release:
	@docker run \
		--rm \
		--privileged \
		-e CGO_ENABLED=1 \
		-e GITHUB_TOKEN \
		-e DOCKER_USERNAME \
		-e DOCKER_PASSWORD \
		-e SLACK_WEBHOOK \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-v $(HOME)/.docker/config.json:/root/.docker/config.json \
		-v `pwd`:/go/src/$(PACKAGE_NAME) \
		-w /go/src/$(PACKAGE_NAME) \
		goreleaser/goreleaser-cross:${GOLANG_CROSS_VERSION} \
		--rm-dist --skip-validate

.PHONY: help
help:
	@echo "Available targets:"
	@echo "  clean               - Removes the build directory."
	@echo "  tests               - Runs Go tests on specific packages."
	@echo "  build               - Compiles the Iris binaries."
	@echo "  install             - Installs the Iris binaries."
	@echo "  contracts           - Generates Go bindings for Ethereum contracts."
	@echo "  build-arm           - Compiles the Iris binaries for ARM64 architecture."
	@echo "  lint                - Runs the GolangCI-Lint tool on the codebase."
	@echo "  build-docker        - Builds a Docker image for the latest Git tag."
	@echo "  push-docker         - Pushes the Docker image for the latest Git tag."
	@echo "  build-docker-develop- Builds a Docker image for the development branch."
	@echo "  release-dry-run     - Performs a dry run of the release process."
	@echo "  release             - Executes the actual release process."
