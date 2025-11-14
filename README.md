# ccgate

一个简洁高效的 Claude Code 平台管理与透明代理工具。

## 特性

- 🔧 **多平台配置管理** - 轻松管理多个 Claude 平台配置
- 🚀 **透明代理** - 自动设置环境变量，无缝代理到 claude 命令
- 🎯 **智能选择** - 支持命令行指定或交互式选择平台
- ✨ **简单易用** - 基于 Cobra 的现代 CLI 设计

## 安装

### 方式一：二进制文件（推荐）

从 [Releases](https://github.com/realguan/ccgate/releases) 下载对应平台的二进制文件。

### 方式二：源码编译

```bash
git clone https://github.com/realguan/ccgate.git
cd ccgate
make build
```

## 快速开始

### 1. 添加平台配置

```bash
# 交互式添加平台
ccgate add
```

系统会提示你输入：
- 平台名称
- 厂商
- Anthropic API Base URL
- 认证令牌
- 模型配置

### 2. 使用平台

```bash
# 交互式选择平台并启动
ccgate

# 指定平台启动
ccgate -p myplatform

# 开始新对话
ccgate chat "你好，Claude！"
```

### 3. 管理平台

```bash
# 列出所有平台
ccgate list

# 删除平台
ccgate delete myplatform

# 查看版本
ccgate version
```

## 配置文件

配置文件默认位于 `~/.ccgate/config.json`，格式如下：

```json
{
  "platforms": [
    {
      "name": "default",
      "vendor": "Anthropic",
      "ANTHROPIC_BASE_URL": "https://api.anthropic.com",
      "ANTHROPIC_AUTH_TOKEN": "your-token-here",
      "ANTHROPIC_MODEL": "claude-3-5-sonnet-20241022",
      "ANTHROPIC_SMALL_FAST_MODEL": "claude-3-5-haiku-20241022"
    }
  ]
}
```

## 命令帮助

```
ccgate [flags] [claude-args...]

Flags:
  -f, --config string   指定配置文件路径
  -p, --platform string 指定平台名称
  -y, --yes            跳过确认提示
  -h, --help           帮助信息

Subcommands:
  list      列出所有平台
  add       添加或更新平台配置
  delete    删除指定平台
  version   显示版本信息
```

## 工作原理

ccgate 通过以下方式工作：

1. 加载用户配置的平台信息
2. 根据参数或交互式选择确定目标平台
3. 设置对应的环境变量（ANTHROPIC_*）
4. 透明代理到本地的 `claude` 可执行文件

## 系统要求

- Go 1.21+
- 已安装并配置好的 Claude CLI 工具

## 开发

```bash
# 运行测试
make test

# 构建项目
make build
```

## 许可证

MIT License

## 贡献

欢迎提交 Issue 和 Pull Request！

## 致谢

感谢 Claude 和 Anthropic 提供的优秀 AI 服务。
