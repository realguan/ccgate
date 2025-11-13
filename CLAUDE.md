<!-- OPENSPEC:START -->
# OpenSpec Instructions

These instructions are for AI assistants working in this project.

Always open `@/openspec/AGENTS.md` when the request:
- Mentions planning or proposals (words like proposal, spec, change, plan)
- Introduces new capabilities, breaking changes, architecture shifts, or big performance/security work
- Sounds ambiguous and you need the authoritative spec before coding

Use `@/openspec/AGENTS.md` to learn:
- How to create and apply change proposals
- Spec format and conventions
- Project structure and guidelines

Keep this managed block so 'openspec update' can refresh the instructions.

<!-- OPENSPEC:END -->

# CLAUDE.md

本文件为Claude Code (claude.ai/code)在该代码库中工作时提供指导。

## 前提
网络不通的情况下可以尝试使用代理 export https_proxy=http://127.0.0.1:7890 http_proxy=http://127.0.0.1:7890 all_proxy=socks5://127.0.0.1:7890
## CLI 使用说明（子命令）

ccgate 使用 Cobra 子命令组织命令，主要子命令如下：

- `list` — 列出所有可用平台
- `add` — 交互式添加或更新平台配置
- `delete <name>` — 删除指定名称的平台
- `version` — 显示版本信息
- `completion` — 生成 shell 自动补全脚本（`bash` 或 `zsh`）

全局 flag:

- `--config, -f <path>`: 指定配置文件路径（默认 `~/.ccgate/config.json`）

# CLAUDE.md

概述

此文件说明 ccgate 项目的实现细节、配置格式与运行流程，内容已与当前代码保持一致（Cobra 子命令式 CLI）。

代码文件概览

- `main.go` — 核心实现：配置加载/保存、交互式选择、环境变量映射与启动流程（调用本地 `claude` 可执行文件）。
- `cli.go` — Cobra 命令注册：提供 `list`, `add`, `delete`, `start`, `version`, `completion` 等子命令，以及全局 `--config/-f` flag。

命令行接口摘要

# CLAUDE.md

概述

本文件说明 ccgate 项目的实现细节、配置格式与运行流程，内容已与当前代码保持一致（Cobra 子命令式 CLI，`start` 子命令已移除，相关 flags 已提升为根命令 flags）。

代码文件概览

- `main.go` — 核心实现：配置加载/保存、交互式选择、环境变量映射与启动流程（调用本地 `claude` 可执行文件）。
- `cli.go` — Cobra 命令注册：提供 `list`, `add`, `delete`, `version`, `completion` 子命令，以及全局 `--config/-f`、`--platform/-p`、`--yes/-y`、`--dry-run` flags。

命令行接口摘要

全局选项

- `--config, -f <path>`: 指定配置文件路径（默认 `~/.ccgate/config.json`）。

注：当前版本的 `ccgate` 不再包含 `--platform`, `--yes`, `--dry-run` 这类根级 flags（它们在先前的版本中曾短暂存在或被讨论）。根命令（直接运行 `ccgate`）仍然是交互式启动的入口；如果需要在脚本或 CI 中自动化启动，请参考下面的迁移/自动化建议。

子命令（常用）

- `list` — 列出所有已配置的平台
- `add` — 以交互方式添加或更新平台（提示输入字段）
- `delete <name>` — 删除指定名称的平台
- `completion <bash|zsh>` — 在 stdout 生成 Shell 自动补全脚本
- `version` — 输出工具版本信息

重要变化说明与迁移指南

历史上 `ccgate` 曾以子命令 `start` 暴露交互/非交互两种启动方式；当前代码库已经简化为：

- 直接运行 `ccgate`（根命令）会进入交互式平台选择并执行启动流程。
- 如果你有自动化/脚本化的需求，请不要依赖已删除的 CLI flags；下面给出两种替代方案：

1) 在脚本中解析配置文件并手动设置环境变量后运行 `claude`：

```bash
# 从 ~/.ccgate/config.json 中提取名为 myPlatform 的配置并设置环境变量（示例使用 jq）
cfg=~/.ccgate/config.json
export ANTHROPIC_BASE_URL=$(jq -r '.platforms[] | select(.name=="myPlatform") .ANTHROPIC_BASE_URL' "$cfg")
export ANTHROPIC_AUTH_TOKEN=$(jq -r '.platforms[] | select(.name=="myPlatform") .ANTHROPIC_AUTH_TOKEN' "$cfg")
export ANTHROPIC_MODEL=$(jq -r '.platforms[] | select(.name=="myPlatform") .ANTHROPIC_MODEL' "$cfg")
export ANTHROPIC_SMALL_FAST_MODEL=$(jq -r '.platforms[] | select(.name=="myPlatform") .ANTHROPIC_SMALL_FAST_MODEL' "$cfg")

# 然后运行 claude
claude
```

2) 或者，使用 `ccgate list` + `ccgate add` 在本地生成/修改配置，然后在 CI 中使用自定义脚本（或小工具）读取配置并启动 `claude`。

配置文件格式与查找顺序

ccgate 支持按下列优先级查找配置文件：
1. 由 `--config` 指定的路径
2. 用户目录下的 `~/.ccgate/config.json`
3. 当前目录下的 `platforms.json`

配置内容示例：

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

核心流程（简要）

1. 程序启动并解析命令行参数（Cobra）。
2. 根据命令选择操作：列出、添加、删除或启动某个平台。
3. 启动流程（根命令或通过 `--platform` 指定）：
   - 如果指定了 `--platform`，程序会尝试做精确匹配；若未找到，会基于 Levenshtein 距离给出相似名称建议。
   - 如果未指定平台，程序会进入交互式选择（promptui），列出可用平台供用户选择。
   - 选择平台后，默认会显示确认提示（可用 `--yes` 跳过）。
   - 如果使用 `--dry-run`，程序仅打印要设置的环境变量（并掩码敏感值），不会执行 `claude`。否则程序会设置环境并尝试执行 `claude` 可执行文件。

敏感信息处理

程序在打印认证令牌时会掩码中间部分以避免在终端泄露完整令牌；dry-run 输出也遵循同样的掩码策略。

实现细节要点

- 平台条目是任意键值对集合，程序把每个条目的键导出为环境变量名并赋对应值（示例中使用 ANTHROPIC_* 前缀）。
- `add` 子命令以交互方式提示常见字段（name, vendor, ANTHROPIC_* 等），并将新条目写回配置文件。
- 平台名称匹配与建议使用简单的 Levenshtein 距离算法，当用户输入近似名称时，工具会输出候选名称供用户参考。

开发与测试

```bash
make build
make test
```

建议的改进

- 在 Makefile 中添加 `completion-install`，自动把生成的补全脚本安装到常见路径。
- 在 CI 中加入对 `go test ./...` 与 `make build` 的验证。

参考

- `README.md`：使用示例与快速入门
- `main.go`, `cli.go`：实现细节