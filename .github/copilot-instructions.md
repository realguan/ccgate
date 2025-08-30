# AI 助手指南 - cctool

此文档为 AI 助手提供在 cctool（Claude Code 平台选择器）代码库中工作的关键信息。

## 项目概述

cctool 是一个基于 Go 语言的交互式工具，用于动态选择和启动不同的 Claude Code 平台接口。主要功能包括平台配置管理、环境变量设置和应用程序启动。

## 架构要点

### 核心组件

1. **配置管理**
   - 配置文件: 优先级顺序为 `~/.cctool/config.json` > `./platforms.json`
   - 示例: 见 `main.go` 中的 `loadPlatforms()` 函数

2. **平台管理**
   - 位置: `main.go` 中的平台相关函数
   - 关键操作: 添加、删除、列出平台配置

3. **环境变量处理**
   - 关键函数: `setEnvironment()`
   - 用途: 根据选定平台动态设置 Claude Code 运行环境

### 数据流

配置加载 → 命令行处理 → 平台管理 → 用户选择 → 环境设置 → 应用启动

## 开发工作流

### 关键命令

```bash
make build      # 构建项目
make test       # 运行测试
make install    # 安装到 ~/bin
make fmt        # 格式化代码
make vet        # 代码检查
```

### 测试策略

- 单元测试位于 `main_test.go`
- 重点测试: 平台验证、配置验证、平台查找功能
- 运行测试覆盖率: `make test-cover`

## 项目特定约定

1. **配置文件格式**
```json
{
  "platforms": [
    {
      "name": "平台名称",
      "vendor": "厂商名称（可选）",
      "ANTHROPIC_BASE_URL": "API基础URL",
      "ANTHROPIC_AUTH_TOKEN": "认证令牌",
      "ANTHROPIC_MODEL": "模型规格",
      "ANTHROPIC_SMALL_FAST_MODEL": "快速小模型名称"
    }
  ]
}
```

2. **错误处理模式**
- 配置验证在 `main.go` 中集中处理
- 使用标准错误返回进行错误传播

3. **命令行接口约定**
- 所有命令行参数使用连字符前缀 (如 `-list`, `-platform`)
- 交互模式为默认行为 (无参数时)

## 集成点

1. **外部依赖**
- `github.com/manifoldco/promptui`: 用于交互式命令行界面

2. **环境变量接口**
- `ANTHROPIC_BASE_URL`
- `ANTHROPIC_AUTH_TOKEN`
- `ANTHROPIC_MODEL`
- `ANTHROPIC_SMALL_FAST_MODEL`

## 常见任务示例

1. **添加新平台**:
```go
// 在 platforms.json 中添加新条目
{
  "name": "新平台",
  "ANTHROPIC_BASE_URL": "https://api.example.com",
  "ANTHROPIC_AUTH_TOKEN": "token",
  "ANTHROPIC_MODEL": "model-name"
}
```

2. **修改环境变量处理**:
- 在 `main.go` 中的 `setEnvironment()` 函数中添加新的环境变量映射

3. **扩展验证规则**:
- 在 `main.go` 中的验证相关函数中添加新的验证逻辑
