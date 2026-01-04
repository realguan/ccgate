# 变量定义
BINARY_NAME=ccgate
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_DIR=build
GO=go
GOFLAGS=-v
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -s -w"

# 构建目标平台
PLATFORMS=darwin/amd64 darwin/arm64 linux/amd64 linux/arm64 windows/amd64

# 颜色输出
GREEN=\033[0;32m
YELLOW=\033[0;33m
RED=\033[0;31m
NC=\033[0m # No Color

.PHONY: help
help: ## 显示帮助信息
	@echo "$(BINARY_NAME) Makefile"
	@echo ""
	@echo "使用方法:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(GREEN)%-15s$(NC) %s\n", $$1, $$2}'

.PHONY: all
all: build ## 构建当前平台

.PHONY: build
build: ## 构建当前平台二进制文件
	@echo "$(GREEN)构建 $(BINARY_NAME) $(VERSION)$(NC)"
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) .

.PHONY: build-all
build-all: ## 构建所有支持平台
	@echo "$(GREEN)构建所有平台$(NC)"
	@mkdir -p $(BUILD_DIR)
	@$(MAKE) $(PLATFORMS)

.PHONY: $(PLATFORMS)
$(PLATFORMS):
	@echo "$(GREEN)构建 $@$(NC)"
	@mkdir -p $(BUILD_DIR)
	$(eval GOOS=$(shell echo $@ | cut -d'/' -f1))
	$(eval GOARCH=$(shell echo $@ | cut -d'/' -f2))
	@GOOS=$(GOOS) GOARCH=$(GOARCH) $(GO) build $(GOFLAGS) $(LDFLAGS) \
		-o $(BUILD_DIR)/$(BINARY_NAME)-$(GOOS)-$(GOARCH)$(if $(filter windows,$(GOOS)),.exe,) .

.PHONY: install
install: ## 安装到 $GOPATH/bin
	@echo "$(GREEN)安装 $(BINARY_NAME)$(NC)"
	$(GO) install $(GOFLAGS) $(LDFLAGS) .

.PHONY: clean
clean: ## 清理构建文件
	@echo "$(YELLOW)清理构建文件$(NC)"
	@rm -rf $(BUILD_DIR)
	@$(GO) clean

.PHONY: test
test: ## 运行测试
	@echo "$(GREEN)运行测试$(NC)"
	$(GO) test -v ./...

.PHONY: lint
lint: ## 运行代码检查
	@echo "$(GREEN)运行代码检查$(NC)"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "$(YELLOW)golangci-lint 未安装，跳过检查$(NC)"; \
	fi

.PHONY: fmt
fmt: ## 格式化代码
	@echo "$(GREEN)格式化代码$(NC)"
	$(GO) fmt ./...

.PHONY: check
check: fmt lint test ## 运行所有检查

# ============ Release 相关 ============

.PHONY: release-check
release-check: ## 检查 release 前的条件
	@echo "$(GREEN)检查 release 准备情况$(NC)"
	@if [ -z "$$VERSION" ]; then \
		echo "$(RED)错误: 请指定 VERSION 参数$(NC)"; \
		echo "示例: make release VERSION=v1.0.0"; \
		exit 1; \
	fi
	@echo "版本: $(VERSION)"
	@if git rev-parse $(VERSION) >/dev/null 2>&1; then \
		echo "$(YELLOW)警告: Tag $(VERSION) 已存在$(NC)"; \
		read -p "是否继续? [y/N] " -n 1 -r; \
		echo; \
		[[ $$REPLY =~ ^[Yy] ]] || exit 1; \
	fi

.PHONY: release-build
release-build: release-check ## 构建所有平台的 release 文件
	@echo "$(GREEN)构建 release $(VERSION)$(NC)"
	@$(MAKE) clean
	@$(MAKE) build-all VERSION=$(VERSION)
	@echo "$(GREEN)生成 checksums$(NC)"
	@cd $(BUILD_DIR) && for file in $(BINARY_NAME)-*; do \
		if [ -f "$$file" ]; then \
			shasum -a 256 "$$file" > "$$file.sha256"; \
		fi; \
	done
	@ls -lh $(BUILD_DIR)/

.PHONY: release-tag
release-tag: release-check ## 创建并推送 git tag
	@echo "$(GREEN)创建 tag $(VERSION)$(NC)"
	@git tag -a $(VERSION) -m "Release $(VERSION)"
	@echo "$(GREEN)推送 tag 到远程$(NC)"
	@git push origin $(VERSION)

.PHONY: release
release: release-build release-tag ## 完整 release 流程（构建 + tag）
	@echo "$(GREEN)==========================================$(NC)"
	@echo "$(GREEN)Release $(VERSION) 创建成功！$(NC)"
	@echo "$(GREEN)==========================================$(NC)"
	@echo ""
	@echo "下一步:"
	@echo "  1. 访问 https://github.com/realguan/ccgate/releases"
	@echo "  2. 找到新创建的 release $(VERSION)"
	@echo "  3. 编辑 release notes"
	@echo "  4. 上传 $(BUILD_DIR)/ 目录下的构建文件"
	@echo ""

.PHONY: release-snapshot
release-snapshot: ## 创建 snapshot release (不创建 git tag)
	@echo "$(YELLOW)构建 snapshot$(NC)"
	@$(MAKE) clean
	@$(MAKE) build-all VERSION=snapshot-$(shell date +%Y%m%d-%H%M%S)
	@ls -lh $(BUILD_DIR)/

# ============ 开发相关 ============

.PHONY: dev
dev: ## 开发模式（运行程序）
	@echo "$(GREEN)运行 $(BINARY_NAME)$(NC)"
	$(GO) run .

.PHONY: update-deps
update-deps: ## 更新依赖
	@echo "$(GREEN)更新依赖$(NC)"
	$(GO) get -u ./...
	$(GO) mod tidy

.PHONY: verify-deps
verify-deps: ## 验证依赖
	@echo "$(GREEN)验证依赖$(NC)"
	$(GO) mod verify
	$(GO) mod tidy
