# ccgate - Claude Code 平台管理与透明代理工具

`ccgate` 是一个强大的 Claude Code 平台配置管理工具，同时也是 `claude` 命令的透明代理。它让你可以轻松管理多个 Claude 平台配置，并在不同环境间快速切换。

## ✨ 核心特性

- 🚀 **透明代理** - 完全兼容 `claude` 命令的所有参数和子命令
- 🎯 **平台管理** - 轻松管理多个 Claude API 配置
- 🔄 **快速切换** - 一键切换不同的 API 提供商和环境
- 💡 **智能选择** - 支持交互式选择或命令行指定平台
- 🔒 **安全可靠** - 敏感信息掩码显示，配置验证完善
- 🎨 **用户友好** - 清晰的提示和错误信息

## 📦 安装

### 使用 Makefile

```bash
make build      # 构建二进制到 build/ccgate
make install    # 安装到 ~/bin
```

### 手动编译

```bash
go build -o build/ccgate
```

## 🚀 快速开始

### 1. 添加第一个平台

```bash
ccgate add
```

交互式输入平台信息：
- **平台名称**: `production`
- **厂商** (可选): `Anthropic`
- **ANTHROPIC_BASE_URL**: `https://api.anthropic.com`
- **ANTHROPIC_AUTH_TOKEN**: `sk-ant-...`
- **ANTHROPIC_MODEL**: `claude-sonnet-4-20250514`
- **ANTHROPIC_SMALL_FAST_MODEL** (可选): `claude-3-5-haiku-20241022`

### 2. 查看所有平台

```bash
ccgate list
```

### 3. 使用平台启动 Claude

```bash
# 指定平台
ccgate -p production --continue

# 交互式选择（多平台时）
ccgate --continue
```

## 📖 使用说明

### 平台管理命令

ccgate 提供了完整的平台管理功能：

```bash
ccgate list                    # 列出所有平台
ccgate add                     # 添加或更新平台
ccgate delete <name>           # 删除指定平台
ccgate version                 # 显示版本信息
ccgate completion bash|zsh     # 生成 Shell 自动补全脚本
```

### 透明代理 claude 命令

**核心功能**：`ccgate` 作为 `claude` 的透明代理，支持所有 `claude` 原生命令和参数。

#### 基本用法

```bash
# 继续最近的对话
ccgate -p prod --continue

# 开始新对话
ccgate -p dev chat "帮我优化代码"

# 无参数启动交互式 claude
ccgate -p staging
```

#### 高级用法

```bash
# 使用自定义模型
ccgate -p prod --model sonnet-4 chat "test"

# 指定上下文目录
ccgate -p dev --context ./src --continue

# 任意 claude 参数都会被正确传递
ccgate -p prod <任意 claude 参数...>
```

#### 交互式平台选择

当没有使用 `-p` 指定平台时，会自动触发交互式选择：

```bash
# 会弹出平台选择菜单
ccgate --continue
ccgate chat "hello world"
```

**行为说明**：
- **单平台**：自动使用该平台，无需选择
- **多平台**：显示交互式菜单供用户选择
- **CI/脚本环境**：必须使用 `-p` 显式指定平台

### 全局选项

```bash
-p, --platform <name>    # 指定平台名称
-f, --config <path>      # 指定配置文件路径 (默认: ~/.ccgate/config.json)
-y, --yes                # 跳过确认提示 (适合脚本使用)
-h, --help               # 显示帮助信息
```

## 💡 使用场景

### 场景 1: 开发/测试/生产环境切换

```bash
# 开发环境调试
ccgate -p dev --continue

# 测试环境验证
ccgate -p staging chat "test feature"

# 生产环境使用
ccgate -p production --continue
```

### 场景 2: 多个 API 提供商对比

```bash
# 官方 Anthropic API
ccgate -p anthropic-official chat "比较性能"

# DeepSeek API
ccgate -p deepseek chat "比较性能"

# 本地测试环境
ccgate -p local-dev chat "比较性能"
```

### 场景 3: CI/CD 自动化

```bash
#!/bin/bash
# 在 CI 脚本中使用，跳过交互确认
ccgate -p ci-environment -y --continue
```

### 场景 4: 团队协作配置

```bash
# 使用团队共享的配置文件
ccgate -f ~/.ccgate/team-config.json -p shared-dev chat "hello"
```

## ⚙️ 配置文件

### 默认配置文件路径

```
~/.ccgate/config.json
```

也可以通过 `--config` 或 `-f` 指定自定义路径。

### 配置文件格式

```json
{
  "platforms": [
    {
      "name": "production",
      "vendor": "Anthropic Official",
      "ANTHROPIC_BASE_URL": "https://api.anthropic.com",
      "ANTHROPIC_AUTH_TOKEN": "sk-ant-api03-...",
      "ANTHROPIC_MODEL": "claude-sonnet-4-20250514",
      "ANTHROPIC_SMALL_FAST_MODEL": "claude-3-5-haiku-20241022"
    },
    {
      "name": "deepseek",
      "vendor": "DeepSeek",
      "ANTHROPIC_BASE_URL": "https://api.deepseek.com/anthropic",
      "ANTHROPIC_AUTH_TOKEN": "sk-...",
      "ANTHROPIC_MODEL": "deepseek-chat",
      "ANTHROPIC_SMALL_FAST_MODEL": "deepseek-chat"
    }
  ]
}
```

### 字段说明

| 字段 | 必填 | 说明 |
|------|------|------|
| `name` | ✅ | 平台名称（唯一标识符） |
| `vendor` | ❌ | 厂商名称（用于显示） |
| `ANTHROPIC_BASE_URL` | ✅ | API 基础 URL |
| `ANTHROPIC_AUTH_TOKEN` | ✅ | 认证令牌 |
| `ANTHROPIC_MODEL` | ✅ | 默认模型 |
| `ANTHROPIC_SMALL_FAST_MODEL` | ❌ | 快速模型（可选） |

## 🎯 工作原理

```
用户输入: ccgate -p prod --continue
          ↓
     1. 解析参数
        ├─ ccgate 专有: -p prod
        └─ claude 参数: --continue
          ↓
     2. 加载配置文件
          ↓
     3. 选择平台 (prod)
          ↓
     4. 确认执行 (可用 -y 跳过)
          ↓
     5. 设置环境变量
          ├─ ANTHROPIC_BASE_URL
          ├─ ANTHROPIC_AUTH_TOKEN
          ├─ ANTHROPIC_MODEL
          └─ ANTHROPIC_SMALL_FAST_MODEL
          ↓
     6. 执行 claude --continue
        (使用 syscall.Exec 进程替换)
```

**关键技术**：使用 `syscall.Exec` 进行进程替换，使得 `ccgate` 完全透明，用户体验与直接运行 `claude` 完全一致。

## 🆚 对比原生 claude

| 功能 | 原生 claude | ccgate |
|------|-------------|--------|
| 启动 Claude | `claude` | `ccgate -p prod` |
| 继续对话 | `claude --continue` | `ccgate -p prod --continue` |
| 新对话 | `claude chat "hello"` | `ccgate -p dev chat "hello"` |
| 环境变量管理 | 手动设置 4 个变量 | 自动设置（选择平台） |
| 多环境切换 | 手动修改环境变量 | `-p` 参数一键切换 |
| 配置管理 | 手动编辑配置 | `add/list/delete` 命令 |
| 参数兼容性 | ✅ 所有参数 | ✅ 完全兼容，透明传递 |

## 🔧 高级功能

### Shell 自动补全

#### Bash

```bash
# 生成补全脚本
ccgate completion bash > /etc/bash_completion.d/ccgate

# 或者添加到 ~/.bashrc
ccgate completion bash >> ~/.bashrc
```

#### Zsh

```bash
# 生成补全脚本
ccgate completion zsh > "${fpath[1]}/_ccgate"

# 重新加载补全
autoload -U compinit && compinit
```

### 自定义配置文件

```bash
# 为不同项目使用不同配置
ccgate -f ~/project-a/config.json -p dev --continue
ccgate -f ~/project-b/config.json -p prod --continue

# 团队共享配置
ccgate -f /shared/team-config.json -p shared --continue
```

### 模糊匹配提示

当平台名称输入错误时，会自动提供相似的建议：

```bash
$ ccgate -p prodd --continue
错误：平台 'prodd' 不存在

你是否想使用以下平台？
  - production
  - prod-backup

运行 'ccgate list' 查看所有可用平台
```

## 🔍 故障排查

### 1. 平台不存在

**问题**：
```bash
$ ccgate -p myplatform --continue
错误：平台 'myplatform' 不存在
```

**解决方案**：
```bash
# 查看所有可用平台
ccgate list

# 或添加新平台
ccgate add
```

### 2. 非交互环境错误

**问题**：
```bash
$ ccgate --continue  # 在 CI 中执行
错误：检测到 3 个平台，但当前环境不支持交互式选择
```

**解决方案**：
```bash
# 使用 -p 明确指定平台，并使用 -y 跳过确认
ccgate -p production -y --continue
```

### 3. 找不到 claude 命令

**问题**：
```bash
错误：找不到 claude 可执行文件
请确保 claude 已安装并在 PATH 中
```

**解决方案**：
```bash
# 检查 claude 是否已安装
which claude

# 如果未安装，请访问
# https://claude.ai/download
```

### 4. 确认提示失败

**问题**：
```bash
# 在非 TTY 环境中运行
错误：操作已取消
```

**解决方案**：
```bash
# 使用 -y 跳过确认提示
ccgate -p prod -y --continue
```

## 📚 完整示例

### 示例 1: 日常开发工作流

```bash
# 早上开始工作，继续昨天的对话
ccgate -p dev --continue

# 开始新功能开发
ccgate -p dev chat "帮我实现用户登录功能"

# 准备发布到生产环境
ccgate -p prod chat "review my code changes"
```

### 示例 2: 多模型对比测试

```bash
# 使用不同 API 提供商测试相同问题
ccgate -p anthropic chat "优化这段代码"
ccgate -p deepseek chat "优化这段代码"
ccgate -p kimi chat "优化这段代码"
```

### 示例 3: 自动化脚本

```bash
#!/bin/bash
# deploy.sh - 自动化部署脚本

# 使用 CI 环境的平台配置
ccgate -f /ci/config.json -p ci-prod -y chat "部署到生产环境"

# 检查部署状态
ccgate -f /ci/config.json -p ci-prod -y --continue
```

### 示例 4: 平台管理

```bash
# 添加新平台
ccgate add
# → 输入: production, Anthropic, https://api.anthropic.com, ...

# 查看所有平台
ccgate list

# 删除旧平台
ccgate delete old-staging

# 更新现有平台（使用相同名称）
ccgate add
# → 输入: production (已存在，将更新)
```

## 🔒 安全最佳实践

1. **不要将配置文件提交到版本控制**
   ```bash
   # 添加到 .gitignore
   echo "~/.ccgate/config.json" >> .gitignore
   ```

2. **使用环境变量存储敏感信息**（未来版本支持）
   ```bash
   export ccgate_PROD_TOKEN="sk-ant-..."
   ```

3. **定期轮换 API 令牌**
   ```bash
   # 更新令牌
   ccgate add  # 使用相同平台名更新
   ```

4. **敏感信息自动掩码**
   ```bash
   # 令牌在显示时自动掩码
   → 认证令牌: sk-a****d8c0
   ```

## 🛠️ 开发

### 运行测试

```bash
make test
```

### 格式化代码

```bash
make fmt
```

### 代码检查

```bash
make vet
```

### 构建与安装

```bash
# 构建
make build

# 安装到 ~/bin
make install
```

## 📖 更多帮助

```bash
# 查看总体帮助
ccgate --help

# 查看子命令帮助
ccgate list --help
ccgate add --help
ccgate delete --help
```

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📄 许可证

MIT License

## 🔗 相关文档

- [实现文档](./IMPLEMENTATION.md) - 详细的架构和实现说明
- [Claude.md](./CLAUDE.md) - Claude Code 集成说明
- [Makefile](./Makefile) - 构建脚本

---

**提示**：如果你觉得 ccgate 有用，请给我们一个 ⭐️ Star
