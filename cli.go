package main

import (
	"fmt"
	"os"
	"time"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var (
	// 全局 flags
	cfgFile      string
	platformName string
	skipConfirm  bool

	// 版本信息（通过 ldflags 在构建时注入）
	Version   = "v0.0.0"
	Commit    = "unknown"
	BuildDate = "unknown"
)

// rootCmd 根命令
var rootCmd = &cobra.Command{
	Use:   "ccgate [flags] [claude-args...]",
	Short: "Claude Code 平台管理与透明代理工具",
	Long: `ccgate 是一个 Claude Code 平台配置管理工具，同时也是 claude 命令的透明代理。

它提供以下功能：
1. 管理多个 Claude 平台配置（list, add, delete）
2. 透明代理 claude 命令，自动设置环境变量

示例:
  ccgate list                    # 列出所有平台
  ccgate add                     # 添加新平台
  ccgate -p prod --continue      # 使用 prod 平台继续对话
  ccgate --continue              # 交互式选择平台后继续对话
  ccgate chat "hello"            # 交互式选择平台后开始新对话`,

	// 禁用默认的 completion 命令（我们自己实现）
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: false,
	},

	// 关键：允许未知 flags 和参数
	FParseErrWhitelist: cobra.FParseErrWhitelist{
		UnknownFlags: true,
	},
	DisableFlagParsing: false, // 保持 flag 解析，但允许未知 flags

	// 禁用参数验证，允许任意参数
	Args: cobra.ArbitraryArgs,

	RunE: handleRootCommand,
}

func init() {
	// 全局 flags（这些不会传递给 claude）
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "f", "", "指定配置文件路径")
	rootCmd.Flags().StringVarP(&platformName, "platform", "p", "", "指定平台名称")
	rootCmd.Flags().BoolVarP(&skipConfirm, "yes", "y", false, "跳过确认提示")

	// 添加子命令
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(deleteCmd)
	rootCmd.AddCommand(versionCmd)
}

// handleRootCommand 处理根命令（透明代理逻辑）
func handleRootCommand(cmd *cobra.Command, args []string) error {
	// 提取 claude 参数（排除 ccgate 专有的 flags）
	claudeArgs := extractClaudeArgs(os.Args[1:])

	// 加载配置
	config, err := loadConfig(cfgFile)
	if err != nil {
		return fmt.Errorf("加载配置失败: %w", err)
	}

	// 验证配置
	if len(config.Platforms) == 0 {
		theme := DefaultTheme()
		DisplayWarning("没有配置任何平台", theme)
		fmt.Println("请先运行 'ccgate add' 添加平台")
		return nil
	}

	// 选择平台（-p 指定 或 交互式）
	// 多平台交互式选择时内部会处理确认循环（支持 ESC 返回）
	// 其他情况（-p 指定或单平台）在外部确认
	platform, err := selectPlatform(config, platformName, claudeArgs, skipConfirm)
	if err != nil {
		return err
	}

	// 当使用 -p 指定平台 或 单平台自动选择时，需要确认（除非 --yes）
	// 交互式多平台选择时内部已经处理了确认
	if platformName != "" || len(config.Platforms) == 1 {
		if err := confirmExecution(platform, claudeArgs, skipConfirm); err != nil {
			return err
		}
	}

	// 透明代理到 claude
	return proxyToClaude(platform, claudeArgs)
}

// extractClaudeArgs 从原始参数中提取要传递给 claude 的参数
func extractClaudeArgs(rawArgs []string) []string {
	var claudeArgs []string
	skip := false

	for _, arg := range rawArgs {
		if skip {
			skip = false
			continue
		}

		// 跳过 ccgate 专有 flags
		switch arg {
		case "--config", "-f", "--platform", "-p":
			skip = true // 跳过下一个参数（值）
			continue
		case "--yes", "-y":
			continue // 仅跳过当前参数
		case "-h", "--help":
			// help 不传递，由 cobra 处理
			continue
		default:
			// 检查是否为组合短选项（如 -py）
			if len(arg) > 1 && arg[0] == '-' && arg[1] != '-' {
				// 可能是组合短选项，需要分解
				filtered := filterCombinedShortFlags(arg)
				if filtered != "" {
					claudeArgs = append(claudeArgs, filtered)
				}
			} else {
				claudeArgs = append(claudeArgs, arg)
			}
		}
	}

	return claudeArgs
}

// filterCombinedShortFlags 过滤组合短选项中的 ccgate 专有 flags
func filterCombinedShortFlags(combined string) string {
	if len(combined) < 2 || combined[0] != '-' {
		return combined
	}

	// 移除 ccgate 专有的短选项: p, y, f
	filtered := "-"
	for i := 1; i < len(combined); i++ {
		c := combined[i]
		if c != 'p' && c != 'y' && c != 'f' {
			filtered += string(c)
		}
	}

	if filtered == "-" {
		return ""
	}
	return filtered
}

// list 子命令
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "列出所有可用平台",
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := loadConfig(cfgFile)
		if err != nil {
			return fmt.Errorf("加载配置失败: %w", err)
		}
		listPlatforms(config.Platforms)
		return nil
	},
}

// add 子命令
var addCmd = &cobra.Command{
	Use:   "add",
	Short: "添加或更新平台配置",
	Long: `交互式添加新平台或更新现有平台配置。

如果平台名称已存在，将更新该平台的配置。`,
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := loadConfig(cfgFile)
		if err != nil {
			return fmt.Errorf("加载配置失败: %w", err)
		}

		newPlatform, err := addPlatform()
		if err != nil {
			return err
		}

		// 检查是否已存在
		existingIndex := -1
		for i, p := range config.Platforms {
			if p.Name == newPlatform.Name {
				existingIndex = i
				break
			}
		}

		theme := DefaultTheme()
		if existingIndex != -1 {
			config.Platforms[existingIndex] = newPlatform
			DisplayWarning(fmt.Sprintf("平台 '%s' 已存在，已更新配置", newPlatform.Name), theme)
		} else {
			config.Platforms = append(config.Platforms, newPlatform)
			DisplaySuccess(fmt.Sprintf("平台 '%s' 添加成功", newPlatform.Name), theme)
		}

		return saveConfig(config, cfgFile)
	},
}

// delete 子命令
var deleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "删除指定名称的平台",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := loadConfig(cfgFile)
		if err != nil {
			return fmt.Errorf("加载配置失败: %w", err)
		}

		name := args[0]
		newPlatforms, err := deletePlatform(config.Platforms, name)
		if err != nil {
			return err
		}

		config.Platforms = newPlatforms
		if err := saveConfig(config, cfgFile); err != nil {
			return err
		}

		theme := DefaultTheme()
		DisplaySuccess(fmt.Sprintf("✓ 平台 '%s' 删除成功", name), theme)
		return nil
	},
}

// version 子命令
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "显示 ccgate 版本信息",
	Run: func(cmd *cobra.Command, args []string) {
		theme := DefaultTheme()
		pterm.Info.Printf("%s %s\n", theme.Colors.Primary.Sprint("ccgate version"), Version)
		fmt.Printf("Commit: %s\n", Commit)
		fmt.Printf("Build Date: %s\n", BuildDate)
		fmt.Println("Claude Code 平台管理与透明代理工具")
	},
}

// Execute 执行根命令
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		theme := DefaultTheme()
		uiErr := &UIError{
			Type:      ErrorTypeSystem,
			Message:   fmt.Sprintf("错误: %v", err),
			Severity:  SeverityError,
			Timestamp: time.Now(),
		}
		uiErr.DisplayError(theme)
		os.Exit(1)
	}
}
