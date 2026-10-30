# 所有 Go 工具鏈都跑在 Docker 容器裡，本機不需要安裝 Go。
# 模組與編譯快取放在具名 volume，`make clean-docker` 可一併清除。

GO_IMAGE   ?= golang:1.25-alpine
BINARY     ?= bin/minesweeper

# 預設交叉編譯成本機（macOS / Linux）可直接執行的二進位檔
GOOS       ?= $(shell uname -s | tr '[:upper:]' '[:lower:]')
UNAME_M    := $(shell uname -m)
GOARCH     ?= $(if $(filter x86_64,$(UNAME_M)),amd64,$(if $(filter aarch64,$(UNAME_M)),arm64,$(UNAME_M)))

MOD_VOLUME   := minesweeper-gomod
BUILD_VOLUME := minesweeper-gobuild

DOCKER_GO = docker run --rm -t \
	-v "$(CURDIR)":/src -w /src \
	-v $(MOD_VOLUME):/go/pkg/mod \
	-v $(BUILD_VOLUME):/root/.cache/go-build \
	-e CGO_ENABLED=0 \
	$(GO_IMAGE)

.PHONY: all build run test vet fmt tidy check clean clean-docker

all: check build

## build: 在容器中交叉編譯出本機可執行檔
build:
	$(DOCKER_GO) sh -c 'GOOS=$(GOOS) GOARCH=$(GOARCH) go build -trimpath -ldflags="-s -w" -o $(BINARY) .'

## run: 編譯後直接在本機終端機執行（遊戲需要真正的 TTY）
run: build
	./$(BINARY)

## test: 執行單元測試
test:
	$(DOCKER_GO) go test ./...

## vet: 靜態檢查
vet:
	$(DOCKER_GO) go vet ./...

## fmt: gofmt 格式化
fmt:
	$(DOCKER_GO) gofmt -l -w .

## tidy: 整理 go.mod / go.sum
tidy:
	$(DOCKER_GO) go mod tidy

## check: tidy + fmt 檢查 + vet + test
check:
	$(DOCKER_GO) sh -c 'go mod tidy && test -z "$$(gofmt -l .)" && go vet ./... && go test ./...'

## clean: 移除編譯產物
clean:
	rm -rf bin

## clean-docker: 移除 Go 模組與編譯快取 volume
clean-docker:
	-docker volume rm $(MOD_VOLUME) $(BUILD_VOLUME)
