package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/fatih/color"
)

// proxyToClaude 透明代理到 claude，设置环境变量并执行
func proxyToClaude(platform *Platform, claudeArgs []string) error {
	// 设置环境变量
	setEnvironmentVariables(platform)

	// 查找 claude 可执行文件
	claudePath, err := exec.LookPath("claude")
	if err != nil {
		return fmt.Errorf(
			"错误：找不到 claude 可执行文件\n" +
				"请确保 claude 已安装并在 PATH 中\n\n" +
				"安装说明: https://claude.ai/download",
		)
	}

	// 打印执行信息
	printExecutionInfo(platform, claudeArgs)

	// 使用 syscall.Exec 进行进程替换（完全透明）
	args := append([]string{"claude"}, claudeArgs...)
	env := os.Environ()

	// 进程替换 - ccgate 进程被 claude 替换
	return syscall.Exec(claudePath, args, env)
}

// setEnvironmentVariables 设置平台相关的环境变量
func setEnvironmentVariables(platform *Platform) {
	os.Setenv("ANTHROPIC_BASE_URL", platform.AnthropicBaseURL)
	os.Setenv("ANTHROPIC_AUTH_TOKEN", platform.AnthropicAuthToken)
	os.Setenv("ANTHROPIC_MODEL", platform.AnthropicModel)
	if platform.AnthropicSmallModel != "" {
		os.Setenv("ANTHROPIC_SMALL_FAST_MODEL", platform.AnthropicSmallModel)
	}
}

// printDryRun 打印 dry-run 模式的输出
func printDryRun(platform *Platform, claudeArgs []string) {
	color.Yellow("\n=== DRY RUN MODE ===")
	color.Cyan("\n→ 将使用平台: %s", platform.Name)
	if platform.Vendor != "" {
		fmt.Printf("  厂商: %s\n", platform.Vendor)
	}

	color.Magenta("\n→ 将设置以下环境变量:")
	fmt.Printf("  ANTHROPIC_BASE_URL=%s\n", platform.AnthropicBaseURL)
	fmt.Printf("  ANTHROPIC_AUTH_TOKEN=%s\n", maskToken(platform.AnthropicAuthToken))
	fmt.Printf("  ANTHROPIC_MODEL=%s\n", platform.AnthropicModel)
	if platform.AnthropicSmallModel != "" {
		fmt.Printf("  ANTHROPIC_SMALL_FAST_MODEL=%s\n", platform.AnthropicSmallModel)
	}

	color.Green("\n→ 将执行命令:")
	if len(claudeArgs) > 0 {
		fmt.Printf("  claude %s\n", strings.Join(claudeArgs, " "))
	} else {
		fmt.Println("  claude (交互式)")
	}

	color.Yellow("\n=== DRY RUN MODE ===\n")
}

// printExecutionInfo 打印即将执行的信息
func printExecutionInfo(platform *Platform, claudeArgs []string) {
	fmt.Println()
	color.Green("✓ 环境变量设置完成")
	color.Cyan("→ 使用平台: %s", platform.Name)

	if len(claudeArgs) > 0 {
		color.Magenta("→ 执行: claude %s\n", strings.Join(claudeArgs, " "))
	} else {
		color.Magenta("→ 启动交互式 claude\n")
	}
}
