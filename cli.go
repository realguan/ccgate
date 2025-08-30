package main

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	cfgFile      string
	platformName string
	skipConfirm  bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "cctool",
	Short: "Claude Code 平台选择器",
	Long:  `交互式选择并启动不同的 Claude Code 平台配置`,
	Run: func(cmd *cobra.Command, args []string) {
		// default behavior: interactive start.
		platforms, err := loadPlatforms(cfgFile)
		if err != nil {
			color.Red("错误: %v\n", err)
			os.Exit(1)
		}
		if err := launchClaudeWithPlatform(platforms, platformName, cfgFile); err != nil {
			color.Red("错误: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	cobra.OnInitialize()

	// global flags
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "f", "", "指定配置文件路径")
	rootCmd.Flags().StringVarP(&platformName, "platform", "p", "", "指定平台名称")
	rootCmd.Flags().BoolVarP(&skipConfirm, "yes", "y", false, "跳过确认提示")

	// add subcommands
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(deleteCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(completionCmd)
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "列出所有可用平台",
	Run: func(cmd *cobra.Command, args []string) {
		platforms, err := loadPlatforms(cfgFile)
		if err != nil {
			color.Red("错误: %v\n", err)
			os.Exit(1)
		}
		listPlatforms(platforms)
	},
}

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "添加新平台",
	Run: func(cmd *cobra.Command, args []string) {
		platforms, err := loadPlatforms(cfgFile)
		if err != nil {
			color.Red("错误: %v\n", err)
			os.Exit(1)
		}

		newPlatform, err := addNewPlatform()
		if err != nil {
			color.Red("错误: %v\n", err)
			os.Exit(1)
		}

		existingIndex := -1
		for i, p := range platforms {
			if p.Name == newPlatform.Name {
				existingIndex = i
				break
			}
		}

		if existingIndex != -1 {
			platforms[existingIndex] = newPlatform
			color.Yellow("平台 '%s' 已存在，正在更新配置...\n", newPlatform.Name)
		} else {
			platforms = append(platforms, newPlatform)
		}

		if err := savePlatforms(platforms, cfgFile); err != nil {
			color.Red("保存失败: %v\n", err)
			os.Exit(1)
		}
		color.Green("平台 '%s' 保存成功!\n", newPlatform.Name)
	},
}

var deleteCmd = &cobra.Command{
	Use:   "delete [name]",
	Short: "删除指定名称的平台",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		platforms, err := loadPlatforms(cfgFile)
		if err != nil {
			color.Red("错误: %v\n", err)
			os.Exit(1)
		}

		name := args[0]
		platforms, err = deletePlatform(platforms, name)
		if err != nil {
			color.Red("错误: %v\n", err)
			os.Exit(1)
		}
		if err := savePlatforms(platforms, cfgFile); err != nil {
			color.Red("保存失败: %v\n", err)
			os.Exit(1)
		}
		color.Green("平台 '%s' 删除成功!\n", name)
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "显示版本信息",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Claude Code 平台选择器 v1.1.0")
	},
}

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh]",
	Short: "生成 shell 自动补全脚本",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		shell := args[0]
		switch shell {
		case "bash":
			if err := rootCmd.GenBashCompletion(os.Stdout); err != nil {
				fmt.Println("生成 bash 完成脚本失败:", err)
				os.Exit(1)
			}
		case "zsh":
			if err := rootCmd.GenZshCompletion(os.Stdout); err != nil {
				fmt.Println("生成 zsh 完成脚本失败:", err)
				os.Exit(1)
			}
		default:
			fmt.Println("不支持的 shell 类型。请选择 bash 或 zsh")
			os.Exit(1)
		}
	},
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
