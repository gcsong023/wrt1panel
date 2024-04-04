GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOARCH=$(shell go env GOARCH)
GOOS=$(shell go env GOOS)
BASE_PATH := $(shell pwd)
BUILD_PATH = $(BASE_PATH)/build
WEB_PATH=$(BASE_PATH)/frontend
SERVER_PATH=$(BASE_PATH)/backend
MAIN= $(BASE_PATH)/cmd/server/main.go
APP_NAME=1panel
ASSETS_PATH= $(BASE_PATH)/cmd/server/web/assets

# 清理前端和后端的构建产物及资源
clean: clean_assets clean_backend

clean_assets:
	rm -rf $(ASSETS_PATH)

clean_backend:
	rm -f $(BUILD_PATH)/$(APP_NAME)

upx_bin:
	upx $(BUILD_PATH)/$(APP_NAME)

# 前端构建
build_frontend:
	cd $(WEB_PATH) && npm install && npm run build:pro

# 后端构建（统一目标，根据GOOS动态调整）
build_backend: build_backend_$(GOOS)

build_backend_linux: GOOS=linux
build_backend_linux:
	cd $(SERVER_PATH) && $(GOBUILD) -trimpath -ldflags '$(LDFLAGS)' -o $(BUILD_PATH)/$(APP_NAME) $(MAIN)

build_backend_darwin: GOOS=darwin
build_backend_darwin:
	cd $(SERVER_PATH) && $(GOBUILD) -trimpath -ldflags '$(LDFLAGS)' -o $(BUILD_PATH)/$(APP_NAME) $(MAIN)

# 根据当前操作系统，选择合适的后端构建目标
build_backend_$(GOOS): $(if $(filter linux,$(GOOS)),build_backend_linux,$(if $(filter darwin,$(GOOS)),build_backend_darwin))

# 定义构建模式
MODE ?= dev

ifeq ($(MODE), stable)
	LDFLAGS += -s -w
else ifeq ($(MODE), dev)
	LDFLAGS +=
endif

# 构建前端和所有支持的后端平台（默认dev模式）
build_all: MODE=stable
build_all: build_frontend build_backend

# 在本地执行完整的构建流程（包括对macOS的后端构建）
build_on_local: clean build_frontend build_backend_darwin upx_bin

# 新增模式相关的构建目标
build_stable: MODE=stable
build_stable: build_all

build_dev: MODE=dev
build_dev: build_all

# 重命名了一些目标以提高其描述性
.PHONY: clean clean_assets clean_backend upx_bin build_frontend build_backend build_backend_linux build_backend_darwin build_all build_on_local build_stable build_dev