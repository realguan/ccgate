# Claude Code Platform Selector (cctool)

这是一个用 Go 语言编写的交互式工具，用于管理并启动 Claude Code 平台配置。

## 功能特性

### 命令行选项

命令基于子命令（Cobra）。全局 flag:

- `--config, -f <path>`: 指定配置文件路径（默认 `~/.cctool/config.json`）。

常用子命令:

```bash
cctool list           # 列出所有可用平台
cctool add            # 交互式添加或更新平台
cctool delete <name>  # 删除指定名称的平台
cctool version        # 显示版本信息
cctool completion bash|zsh  # 生成 shell 自动补全脚本
```

### 示例

```bash
# 列出所有平台
cctool list

# 交互式选择并启动（默认：直接运行根命令）
cctool

# 使用指定平台直接启动（flags 在根命令上）
cctool --platform qwen

# 交互选择后跳过确认并直接启动
cctool --yes

# 仅打印将设置的环境变量（dry-run）
cctool --platform qwen --dry-run

# 添加新平台
cctool add

# 删除平台
cctool delete default

# 生成 bash 自动补全脚本
cctool completion bash > /etc/bash_completion.d/cctool

# 使用自定义配置文件
cctool --config /path/to/custom-config.json list
```

## 使用方法

cctool 是一个用 Go 编写的轻量命令行工具，用来管理并启动不同的 Claude Code 平台配置（通过设置环境变量并执行 `claude` 可执行文件）。

主要特点:
- 使用交互式选择（promptui）或命令行子命令启动平台
- 平台配置使用 JSON 文件（默认 `~/.cctool/config.json`）
- 基于 Cobra 提供子命令、help 与自动补全支持

### 安装

推荐使用 Makefile：

```bash
make build    # 构建二进制到 build/cctool
make install  # 安装到 ~/bin
```

### 快速使用说明（CLI 概览）

全局 flag:

- `--config, -f <path>`: 指定配置文件路径（默认 `~/.cctool/config.json`）

主要子命令:

- `list` — 列出所有可用平台
- `add` — 交互式添加或更新平台配置
- `delete <name>` — 删除指定名称的平台
- `version` — 输出工具版本
- `completion <bash|zsh>` — 在 stdout 生成自动补全脚本

### 示例

```bash
# 列出所有平台
cctool list

# 交互式选择并启动（默认行为 — 直接运行根命令）
cctool

# 使用指定平台直接启动（flags 在根命令上）
cctool --platform myPlatform

# 指定配置文件并列出
cctool --config ./platforms.json list

# 仅打印环境变量（dry-run）
cctool --platform myPlatform --dry-run

# 生成 bash 自动补全脚本
cctool completion bash > /etc/bash_completion.d/cctool
```

## 配置文件

默认路径：`~/.cctool/config.json`（也可使用 `--config` 指定）

格式示例：

```json
{
  "platforms": [
    {
      "name": "default",
      "vendor": "Anthropic",
      "ANTHROPIC_BASE_URL": "https://api.anthropic.com",
      "ANTHROPIC_AUTH_TOKEN": "<REDACTED>",
      "ANTHROPIC_MODEL": "claude-2",
      "ANTHROPIC_SMALL_FAST_MODEL": "claude-2.1-small"
    }
  ]
}
```

## 安全与注意事项

- 输出/日志中会掩码认证令牌的一部分（dry-run 也会掩码）
- 不要将敏感令牌以明文提交到版本控制

## 开发

```bash
# 运行测试
make test

# 格式化 & 检查
make fmt
make vet
```

如需帮助，运行 `cctool --help` 或 `cctool <subcommand> --help` 获取自动生成的使用说明。