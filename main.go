package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/spf13/cobra"
)

var (
	// Version 在编译时通过 ldflags 注入
	Version = "dev"

	platformName string
	dryRun       bool
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "ccgate [claude-args...]",
		Short: "Claude Code 的透明代理与配置管理工具",
		Long: `ccgate 允许你管理多个 Claude Code 平台配置（包括官方和第三方兼容接口），
并能以透明代理的方式启动 claude 客户端，自动注入相应的环境变量。`,
		Version:             Version,
		DisableFlagParsing:  false,
		Run:                 runProxy,
	}

	rootCmd.Flags().StringVarP(&platformName, "platform", "p", "", "直接指定要使用的平台名称")
	rootCmd.Flags().BoolVar(&dryRun, "dry-run", false, "仅打印环境变量和命令，不实际执行")

	rootCmd.FParseErrWhitelist.UnknownFlags = true

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func runProxy(cmd *cobra.Command, args []string) {
	config, configPath, err := LoadConfig()
	if err != nil {
		fmt.Printf("加载配置失败: %v\n", err)
		os.Exit(1)
	}

	if len(config.Platforms) == 0 {
		fmt.Printf("⚠️  未找到配置文件: %s\n", configPath)
		fmt.Println("正在生成示例配置...")
		
		example := GenerateExampleConfig()
		if err := SaveConfig(example, configPath); err != nil {
			fmt.Printf("保存示例配置失败: %v\n", err)
			os.Exit(1)
		}
		
		fmt.Printf("✅ 已生成示例配置到: %s\n请编辑该文件填入真实的 API Key 后重新运行。\n", configPath)
		os.Exit(0)
	}

	var selectedPlatform *Platform

	if platformName != "" {
		for _, p := range config.Platforms {
			if p.Name == platformName {
				selectedPlatform = &p
				break
			}
		}
		if selectedPlatform == nil {
			fmt.Printf("❌ 找不到名为 '%s' 的平台配置\n", platformName)
			os.Exit(1)
		}
	} else {
		p, err := SelectPlatform(config.Platforms)
		if err != nil {
			if err.Error() != "未选择任何平台" {
				fmt.Println(err)
			}
			os.Exit(0)
		}
		selectedPlatform = p
	}

	claudePath, err := exec.LookPath("claude")
	if err != nil {
		fmt.Println("❌ 未找到 'claude' 命令。请先安装 Claude Code CLI。")
		fmt.Println("安装指南: https://docs.anthropic.com/en/docs/claude-code")
		os.Exit(1)
	}

	env := os.Environ()
	env = append(env, fmt.Sprintf("ANTHROPIC_BASE_URL=%s", selectedPlatform.AnthropicBaseURL))
	env = append(env, fmt.Sprintf("ANTHROPIC_AUTH_TOKEN=%s", selectedPlatform.AnthropicAuthToken))
	env = append(env, fmt.Sprintf("ANTHROPIC_MODEL=%s", selectedPlatform.AnthropicModel))
	if selectedPlatform.AnthropicSmallModel != "" {
		env = append(env, fmt.Sprintf("ANTHROPIC_SMALL_FAST_MODEL=%s", selectedPlatform.AnthropicSmallModel))
	}
	
	for k, v := range selectedPlatform.ExtraEnv {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}

	passArgs := filterArgs(os.Args[1:])

	if dryRun {
		fmt.Println("🔍 Dry Run 模式")
		fmt.Printf("执行命令: %s %v\n", claudePath, passArgs)
		fmt.Println("环境变量:")
		fmt.Printf("  ANTHROPIC_BASE_URL=%s\n", selectedPlatform.AnthropicBaseURL)
		fmt.Printf("  ANTHROPIC_AUTH_TOKEN=%s\n", selectedPlatform.MaskedToken())
		fmt.Printf("  ANTHROPIC_MODEL=%s\n", selectedPlatform.AnthropicModel)
		return
	}

	execArgs := append([]string{"claude"}, passArgs...)
	
	if err := syscall.Exec(claudePath, execArgs, env); err != nil {
		fmt.Printf("❌ 启动 claude 失败: %v\n", err)
		os.Exit(1)
	}
}

func filterArgs(args []string) []string {
	var result []string
	skipNext := false
	for _, arg := range args {
		if skipNext {
			skipNext = false
			continue
		}
		if arg == "-p" || arg == "--platform" {
			skipNext = true
			continue
		}
		if arg == "--dry-run" {
			continue
		}
		result = append(result, arg)
	}
	return result
}