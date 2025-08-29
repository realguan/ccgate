# Claude Code Platform Selector (cctool) v1.1.0

这是一个用 Go 语言编写的交互式工具，用于动态选择和启动 Claude Code 平台接口。

## 功能特性

- 🔄 交互式平台选择界面
- 📁 多配置文件支持（JSON格式）
- 🛠️ 命令行参数支持（列表、添加、删除平台等）
- 🌐 多平台支持（Anthropic、Qwen等）
- ✅ 配置验证和错误处理
- 🧪 单元测试覆盖

## 安装

### 方法1：使用 Makefile（推荐）
```bash
# 构建并安装
make install

# 或者分步执行
make build
make install
```

### 方法2：手动构建
```bash
# 构建二进制文件
go build -o build/cctool

# 复制到 ~/bin 目录
mkdir -p ~/bin
cp build/cctool ~/bin/cctool
```

### 方法3：直接安装
```bash
# 直接安装到 GOPATH/bin
go install

# 然后复制到 ~/bin
cp $(go env GOPATH)/bin/cctool ~/bin/cctool
```

## 使用方法

### 交互式选择平台并启动 Claude Code
```bash
cctool
```

运行后会显示可用平台列表，输入数字选择平台，工具将自动设置环境变量并启动 Claude Code。

### 命令行选项

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

## 配置文件

平台配置存储在 JSON 文件中，支持以下查找顺序：

1. `~/.cctool/config.json` (用户特定配置)
2. 当前目录中的 `platforms.json`

配置文件格式如下：

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

可以修改此文件以添加或更改平台配置。

## 开发

### 构建
```bash
# 构建项目
make build

# 构建所有平台版本
make build-all

# 构建特定平台版本
make build-linux
make build-mac
make build-windows
```

### 清理
```bash
make clean
```

### 代码格式化和检查
```bash
# 格式化代码
make fmt

# 检查代码问题
make vet
```

### 运行测试
```bash
# 运行测试
make test

# 运行测试并查看覆盖率
make test-cover
```

### 依赖管理
```bash
# 整理Go模块依赖
make tidy
```

### 查看所有可用命令
```bash
make help
```