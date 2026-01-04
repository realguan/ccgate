# ccgate

<div align="center">

**Claude Code 透明代理与配置管理工具**

[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

</div>

## 简介

ccgate 是一个用于管理多个 Claude Code 平台配置的工具。它允许你在官方 Anthropic API 和各种兼容第三方服务之间轻松切换，并通过透明代理的方式启动 `claude` 客户端，自动注入相应的环境变量。

### 主要特性

- **多平台管理** - 支持配置多个 Claude Code 平台（官方及第三方兼容接口）
- **TUI 交互界面** - 美观的终端用户界面，支持键盘快捷操作
- **模版向导** - 内置主流供应商配置模版，快速添加新平台
- **置顶功能** - 常用平台可置顶显示，优先选择
- **透明代理** - 自动注入环境变量，无感知切换平台
- **配置安全** - Token 脱敏显示，防止意外泄露

## 支持的供应商

ccgate 内置了以下供应商的配置模版：

| 供应商 | 描述 |
|--------|------|
| [Anthropic](https://www.anthropic.com) | Claude 官方接口 |
| [MiniMax](https://www.minimax.chat) | 海螺大模型 |
| [DeepSeek](https://www.deepseek.com) | 深度求索 |
| [Moonshot](https://www.moonshot.cn) | Kimo 月之暗面 |
| [ZhipuAI](https://www.bigmodel.cn) | 智谱 GLM |
| Custom | 自定义配置 |

## 安装

### 前置要求

- Go 1.24 或更高版本
- 已安装 [Claude Code CLI](https://docs.anthropic.com/en/docs/claude-code)

### 从源码安装

```bash
git clone https://github.com/realguan/ccgate.git
cd ccgate
go install
```

### 使用 Go 安装

```bash
go install github.com/realguan/ccgate@latest
```

确保 `$GOPATH/bin` 已添加到你的 `PATH` 环境变量中。

## 快速开始

### 1. 首次运行

首次运行时，ccgate 会自动生成示例配置文件：

```bash
ccgate
```

配置文件位于 `~/.cc-proxy/config.json`，格式如下：

```json
{
  "platforms": [
    {
      "name": "anthropic-official",
      "vendor": "Anthropic",
      "description": "Claude 官方接口",
      "ANTHROPIC_BASE_URL": "https://api.anthropic.com",
      "ANTHROPIC_AUTH_TOKEN": "sk-ant-...",
      "ANTHROPIC_MODEL": "claude-3-5-sonnet-20240620",
      "pinned": true
    },
    {
      "name": "minimax",
      "vendor": "MiniMax",
      "ANTHROPIC_BASE_URL": "https://api.minimax.chat/v1",
      "ANTHROPIC_AUTH_TOKEN": "your-api-key",
      "ANTHROPIC_MODEL": "abab6.5-chat"
    }
  ]
}
```

### 2. 编辑配置

用你的真实 API Key 替换示例中的 `ANTHROPIC_AUTH_TOKEN`。

### 3. 启动 Claude Code

运行 ccgate 后，通过 TUI 界面选择平台，即可自动启动 `claude` 客户端：

```bash
ccgate
```

## 使用方法

### 命令行参数

```bash
# 直接指定平台启动
ccgate -p anthropic-official

# 传递参数给 claude
ccgate -p minimax --help

# 仅打印环境变量，不实际执行
ccgate --dry-run
```

### TUI 快捷键

| 按键 | 功能 |
|------|------|
| `↑` / `↓` | 导航列表 |
| `Enter` | 选择平台并启动 |
| `a` | 添加新配置 |
| `e` | 编辑选中配置 |
| `p` | 置顶/取消置顶 |
| `x` | 删除配置 |
| `q` / `Ctrl+C` | 退出 |

### 添加新配置

1. 按 `a` 进入添加向导
2. 选择供应商模版（或选择 Custom 自定义）
3. 填写配置信息：
   - Name：配置名称（如 `dev-minimax`）
   - Vendor：厂商名称
   - Base URL：API 地址
   - Token：API 认证令牌
   - Model：模型 ID
4. 按 Enter 保存

## 配置文件

配置文件支持以下字段：

```json
{
  "platforms": [
    {
      "name": "配置名称",
      "vendor": "厂商名称",
      "description": "配置描述",
      "pinned": false,
      "ANTHROPIC_BASE_URL": "https://api.example.com",
      "ANTHROPIC_AUTH_TOKEN": "your-api-token",
      "ANTHROPIC_MODEL": "model-id",
      "ANTHROPIC_SMALL_FAST_MODEL": "optional-small-model",
      "extra_env": {
        "CUSTOM_VAR": "custom-value"
      }
    }
  ]
}
```

### 配置文件路径

ccgate 按以下优先级查找配置文件：

1. `~/.cc-proxy/config.json`
2. `~/.ccgate/config.json`

## 开发

### 项目结构

```
ccgate/
├── main.go      # 主程序入口和命令行处理
├── config.go    # 配置加载和保存逻辑
├── tui.go       # TUI 界面实现
├── go.mod       # Go 模块定义
└── go.sum       # 依赖校验和
```

### 依赖

- [cobra](https://github.com/spf13/cobra) - 命令行框架
- [bubbletea](https://github.com/charmbracelet/bubbletea) - TUI 框架
- [bubbles](https://github.com/charmbracelet/bubbles) - TUI 组件库
- [lipgloss](https://github.com/charmbracelet/lipgloss) - 样式库

### 构建

```bash
go build -o ccgate .
```

## 许可证

本项目采用 MIT 许可证 - 详见 [LICENSE](LICENSE) 文件。

## 贡献

欢迎提交 Issue 和 Pull Request！

## 相关链接

- [Claude Code 官方文档](https://docs.anthropic.com/en/docs/claude-code)
- [Claude API 文档](https://docs.anthropic.com/en/api/index)

---

<div align="center">

Made with ❤️ by the community

</div>
