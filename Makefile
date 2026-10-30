# 踩地雷 Makefile
# 自動偵測：本機有 go 則直接使用本地工具鏈（極速），沒有則自動退回使用 Docker 容器。
# 亦可手動指定 USE_DOCKER=1 強制使用容器編譯。

VERSION      ?= v0.2.1
BINARY       ?= bin/minetui
DIST_DIR     ?= dist
GO_IMAGE     ?= golang:1.25-alpine
MOD_VOLUME   := minetui-gomod
BUILD_VOLUME := minetui-gobuild

LDFLAGS      := -s -w -X main.version=$(VERSION)

# 自動判斷本機是否存在 go
HAS_LOCAL_GO := $(shell which go 2>/dev/null)

ifeq ($(HAS_LOCAL_GO),)
  USE_DOCKER ?= 1
else
  USE_DOCKER ?= 0
endif

# 預設交叉編譯參數（僅 Docker 模式需要指定目標平台）
GOOS    ?= $(shell uname -s | tr '[:upper:]' '[:lower:]')
UNAME_M := $(shell uname -m)
GOARCH  ?= $(if $(filter x86_64,$(UNAME_M)),amd64,$(if $(filter aarch64,$(UNAME_M)),arm64,$(UNAME_M)))

DOCKER_RUN = docker run --rm -t \
	-v "$(CURDIR)":/src -w /src \
	-v $(MOD_VOLUME):/go/pkg/mod \
	-v $(BUILD_VOLUME):/root/.cache/go-build \
	-e CGO_ENABLED=0 \
	$(GO_IMAGE)

ifeq ($(USE_DOCKER),1)
  TOOL_ENV   := [Docker: $(GO_IMAGE)]
  GO_CMD     := $(DOCKER_RUN)
  BUILD_CMD  := $(DOCKER_RUN) sh -c 'GOOS=$(GOOS) GOARCH=$(GOARCH) go build -trimpath -ldflags="$(LDFLAGS)" -o $(BINARY) .'
  CHECK_CMD  := $(DOCKER_RUN) sh -c 'go mod tidy && test -z "$$(gofmt -l .)" && go vet ./... && go test ./...'
else
  TOOL_ENV   := [Local: $(shell go version 2>/dev/null | cut -d' ' -f3)]
  GO_CMD     :=
  BUILD_CMD  := CGO_ENABLED=0 go build -trimpath -ldflags="$(LDFLAGS)" -o $(BINARY) .
  CHECK_CMD  := go mod tidy && test -z "$$(gofmt -l .)" && go vet ./... && go test ./...
endif

.PHONY: all info build run test vet fmt tidy check clean clean-docker dist

all: info check build

## info: 顯示目前使用的 Go 工具鏈模式
info:
	@echo "==> Go toolchain: $(TOOL_ENV)"

## build: 編譯本機可執行檔
build:
	@echo "==> Building $(BINARY) $(TOOL_ENV)"
	$(BUILD_CMD)

## run: 編譯後直接在本機終端機執行
run: build
	./$(BINARY)

## test: 執行單元測試
test:
	@echo "==> Running tests $(TOOL_ENV)"
	$(GO_CMD) go test ./...

## vet: 靜態檢查
vet:
	@echo "==> Running go vet $(TOOL_ENV)"
	$(GO_CMD) go vet ./...

## fmt: 程式碼格式化
fmt:
	@echo "==> Running gofmt $(TOOL_ENV)"
	$(GO_CMD) gofmt -l -w .

## tidy: 整理模組依賴
tidy:
	@echo "==> Running go mod tidy $(TOOL_ENV)"
	$(GO_CMD) go mod tidy

## check: 執行完整檢查 (tidy + fmt + vet + test)
check:
	@echo "==> Running checks $(TOOL_ENV)"
	$(CHECK_CMD)

## dist: 交叉編譯所有主流平台的發布包與校驗檔 (macOS / Linux / Windows)
dist:
	@echo "==> Building cross-platform distribution archives in $(DIST_DIR)/"
	@rm -rf $(DIST_DIR) && mkdir -p $(DIST_DIR)
	@for target in "darwin/arm64" "darwin/amd64" "linux/amd64" "linux/arm64" "windows/amd64"; do \
		os=$${target%/*}; \
		arch=$${target#*/}; \
		ext=""; \
		[ "$$os" = "windows" ] && ext=".exe"; \
		bin="$(DIST_DIR)/minetui$$ext"; \
		echo "    -> Compiling $$os/$$arch..."; \
		CGO_ENABLED=0 GOOS=$$os GOARCH=$$arch go build -trimpath -ldflags="$(LDFLAGS)" -o "$$bin" . || exit 1; \
		if [ "$$os" = "windows" ]; then \
			(cd $(DIST_DIR) && zip -q "minetui_$(VERSION)_$${os}_$${arch}.zip" "minetui$$ext" && rm "minetui$$ext"); \
		else \
			(cd $(DIST_DIR) && tar -czf "minetui_$(VERSION)_$${os}_$${arch}.tar.gz" "minetui" && rm "minetui"); \
		fi; \
	done
	@echo "    -> Generating checksums..."
	@(cd $(DIST_DIR) && shasum -a 256 minetui_* > checksums.txt)
	@echo "==> Distribution archives created in $(DIST_DIR)/:"
	@ls -lh $(DIST_DIR)

## clean: 移除編譯產物
clean:
	rm -rf bin dist

## clean-docker: 移除 Docker 模組與編譯快取 volume
clean-docker:
	-docker volume rm $(MOD_VOLUME) $(BUILD_VOLUME)
