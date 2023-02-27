# Project metadata
GOVERSION:=1.19
SERVER?=ghcr.io
OWNER?=gabrielopesantos
NAME?=eco
GIT_COMMIT?=$(shell git rev-parse HEAD)
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null)

# Current system information
#    != go env GOOS
GOOS?=$(shell go env GOOS)
GOARCH?=$(shell go env GOARCH)

BIN_PATH := ./bin/${GOOS}_${GOARCH}/${NAME}

# List of ldflags
LD_FLAGS:=\
	-s \
	-w \
	-X 'main.Name=${NAME}' \
	-X 'main.Version=${VERSION}' \
	-X 'main.GitCommit=${GIT_COMMIT}' \

# Builds project locally
.PHONY: dev
dev:
	@echo "Building ${NAME} for ${GOOS}/${GOARCH}"
	@rm -f "${BIN_PATH}"
	@env \
		CGO_ENABLED="0" \
		go build \
			-ldflags "${LD_FLAGS}" \
			-o "${BIN_PATH}"


.PHONY: run
# Build and run project locally
run: dev
	@echo "Running ${BIN_PATH}"
	@${BIN_PATH}

.PHONY: build
build-docker:
	@echo "+ $@"
	@docker image build \
		--force-rm --no-cache \
		--build-arg TARGETOS=${GOOS} \
		--build-arg TARGETARCH=${GOARCH} \
		--build-arg BIN_NAME=${NAME} \
		-t $(SERVER)/$(OWNER)/$(NAME):$(VERSION) .
