# CLAUDE.md

本文件为Claude Code (claude.ai/code)在该代码库中工作时提供指导。

## 前提
网络不通的情况下可以尝试使用代理 export https_proxy=http://127.0.0.1:7890 http_proxy=http://127.0.0.1:7890 all_proxy=socks5://127.0.0.1:7890

## 项目概述

这是一个基于Go语言的交互式工具，名为"Claude Code平台选择器"(cctool) v1.1.0，允许用户动态选择和启动不同的Claude Code平台接口。该工具管理不同平台的环境变量，并使用适当的配置启动Claude Code应用程序。

## 功能特性

- 🔄 交互式平台选择界面
- 📁 多配置文件支持（JSON格式）
- 🛠️ 命令行参数支持（列表、添加、删除平台等）
- 🌐 多平台支持（Anthropic、Qwen等）
- ✅ 配置验证和错误处理
- 🧪 单元测试覆盖

## 代码架构

- **main.go**: 核心应用逻辑，包含以下功能函数：
  - 从JSON文件加载平台配置
  - 使用promptui进行交互式平台选择
  - 根据选定平台设置环境变量
  - 启动Claude Code应用程序
  - 配置验证和错误处理
  - 命令行参数解析
  
- **main_test.go**: 单元测试文件，包含平台验证、配置验证和平台查找功能的测试

- **platforms.json**: 配置文件，包含平台定义：
  - 平台名称
  - 厂商信息（可选）
  - API基础URL
  - 认证令牌
  - 模型规格

- **Makefile**: 构建和开发命令，支持多平台构建

- **go.mod/go.sum**: Go模块依赖，主要使用github.com/manifoldco/promptui进行交互式提示

## 常用开发命令

### 构建
```bash
# 构建项目
make build

# 安装二进制文件到~/bin
make install

# 安装二进制文件到GOPATH/bin
make install-go

# 清理构建产物
make clean

# 整理Go模块依赖
make tidy

# 格式化代码
make fmt

# 检查代码问题
make vet

# 构建所有平台版本
make build-all

# 构建特定平台版本
make build-linux
make build-mac
make build-windows
```

### 测试
```bash
# 运行测试
make test

# 运行测试并查看覆盖率
make test-cover

# 运行特定测试
go test -v ./... -run TestName
```

### 开发
```bash
# 直接运行应用程序
go run main.go

# 构建并运行
make build && ./build/cctool

# 格式化代码
go fmt ./...

# 检查代码问题
go vet ./...

# 查看依赖
go list -m all
```

### 帮助
```bash
# 显示可用的make目标
make help
```

## 配置

平台配置存储在JSON文件中，支持以下查找顺序：

1. `~/.cctool/config.json` (用户特定配置)
2. 当前目录中的`platforms.json`
3. `/Users/guan/git/cc-repo/ccer/platforms.json` (备用路径)

配置文件结构如下：
```json
{
  "platforms": [
    {
      "name": "平台名称",
      "vendor": "厂商名称（可选）",
      "ANTHROPIC_BASE_URL": "API基础URL",
      "ANTHROPIC_AUTH_TOKEN": "认证令牌",
      "ANTHROPIC_MODEL": "模型名称",
      "ANTHROPIC_SMALL_FAST_MODEL": "快速小模型名称"
    }
  ]
}
```

## 关键组件

1. **配置层**: 负责从JSON文件加载和解析平台配置，支持多文件位置回退机制
2. **平台管理**: 支持添加、删除、列出平台配置
3. **表示层**: 使用promptui创建交互式菜单以从可用平台中选择
4. **环境管理**: 根据选定的平台配置设置环境变量
5. **应用启动**: 使用配置的环境变量执行`claude`命令
6. **验证层**: 对平台配置进行验证，确保必要字段不为空

## 架构详情

应用程序遵循简单的线性流程：

1. **初始化**: 应用程序通过解析命令行参数启动
2. **配置加载**: 使用`loadPlatforms()`函数从JSON文件加载平台配置，该函数在多个位置搜索
3. **参数处理**: 根据命令行参数执行相应操作（列表、添加、删除平台等）
4. **用户交互**: 使用`selectPlatformInteractive()`向用户呈现交互式菜单以选择平台
5. **环境设置**: 使用`setEnvironment()`将选定的平台配置应用到环境变量
6. **应用启动**: 使用`launchClaudeCode()`在配置的环境中启动Claude Code应用程序

配置加载机制支持多个文件位置以提高灵活性：
- 用户特定配置位于`~/.cctool/config.json`
- 本地配置位于`./platforms.json`
- 备用配置位于固定系统路径

平台选择使用promptui库提供干净的交互式终端界面，用于在可用平台之间选择。

## 命令行参数

cctool支持以下命令行参数：

```bash
cctool [选项]

选项:
  -list          列出所有可用平台
  -platform name 直接使用指定平台
  -f path        指定配置文件路径
  -add           添加新平台
  -delete name   按名称删除平台
  -help, -h      显示帮助信息
  -version, -v   显示版本信息
```

### 参数说明

- `-list`: 显示所有配置的平台列表
- `-platform <name>`: 直接使用指定名称的平台配置
- `-f <path>`: 指定配置文件路径
- `-add`: 交互式添加新平台配置
- `-delete <name>`: 删除指定名称的平台配置
- `-help`, `-h`: 显示帮助信息
- `-version`, `-v`: 显示工具版本信息

### 示例用法

```bash
# 列出所有平台
cctool -list

# 直接使用指定平台
cctool -platform qwen

# 添加新平台
cctool -add

# 删除平台
cctool -delete default

# 使用自定义配置文件
cctool -f /path/to/custom-config.json
```

## 高级架构

cctool应用程序设计具有清晰的关注点分离：

1. **配置层**: 负责从JSON文件加载和解析平台配置。它实现了一个回退机制，确保应用程序可以在多个位置找到配置文件。

2. **命令行处理层**: 解析和处理命令行参数，执行相应的操作。

3. **平台管理层**: 提供平台的增删改查功能。

4. **表示层**: 使用promptui库创建用于平台选择的交互式命令行界面。这为在不同Claude Code平台配置之间选择提供了用户友好的体验。

5. **环境管理层**: 处理平台配置到环境变量的映射。这种抽象允许应用程序根据用户选择动态配置Claude Code运行时环境。

6. **执行层**: 负责在正确配置的环境变量中启动Claude Code应用程序。这一层确保在启动应用程序时正确应用选定的平台配置。

7. **验证层**: 对平台配置进行验证，确保必要字段不为空。

这些层之间的数据流是单向的：
配置 → 命令行处理 → 平台管理 → 用户选择 → 环境设置 → 应用程序启动

这种架构使得通过简单地向配置文件添加条目来扩展新平台变得容易，而无需更改代码。同时，通过添加验证层和错误处理，提高了应用程序的健壮性。