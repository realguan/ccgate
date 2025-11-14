package main

import (
	"fmt"
	"os"
	"time"

	"github.com/pterm/pterm"
)

// Theme 定义统一的主题色彩系统
type Theme struct {
	Colors   ThemeColors
	Spacing  ThemeSpacing
	Animation AnimationConfig
}

// ThemeColors 定义主题色彩
type ThemeColors struct {
	Primary   pterm.Color // 主色：用于主要操作和强调
	Secondary pterm.Color // 辅助色：用于次要信息和标签
	Success   pterm.Color // 成功色：用于确认信息和成功状态
	Warning   pterm.Color // 警告色：用于警告和注意事项
	Error     pterm.Color // 错误色：用于错误信息和失败状态
	Info      pterm.Color // 信息色：用于提示信息和一般消息
	Muted     pterm.Color // 弱化色：用于次要文本和占位符
	Highlight pterm.Color // 高亮色：用于选中和焦点状态
}

// ThemeSpacing 定义主题间距
type ThemeSpacing struct {
	XS int // 1空格 - 紧密相关元素
	SM int // 2空格 - 相关元素
	MD int // 3空格 - 一般分隔
	LG int // 4空格 - 区块分隔
	XL int // 6空格 - 大区块分隔
}

// AnimationConfig 定义动画配置
type AnimationConfig struct {
	Duration time.Duration // 动画持续时间
	Easing   string        // 缓动函数
	Enabled  bool          // 是否启用动画
}

// DefaultTheme 返回默认主题
func DefaultTheme() *Theme {
	return &Theme{
		Colors: ThemeColors{
			Primary:   pterm.FgBlue,
			Secondary: pterm.FgCyan,
			Success:   pterm.FgGreen,
			Warning:   pterm.FgYellow,
			Error:     pterm.FgRed,
			Info:      pterm.FgMagenta,
			Muted:     pterm.FgGray,
			Highlight: pterm.FgWhite,
		},
		Spacing: ThemeSpacing{
			XS: 1,
			SM: 2,
			MD: 3,
			LG: 4,
			XL: 6,
		},
		Animation: AnimationConfig{
			Duration: 200 * time.Millisecond,
			Easing:   "ease-in-out",
			Enabled:  true,
		},
	}
}

// ResponsiveLayout 响应式布局配置
type ResponsiveLayout struct {
	MinWidth     int  // 最小宽度阈值
	MaxWidth     int  // 最大宽度阈值
	CompactMode  bool // 是否启用紧凑模式
	ShowDetails  bool // 是否显示详细信息
	SupportsColor bool // 是否支持颜色
}

// GetResponsiveLayout 获取响应式布局配置
func GetResponsiveLayout() ResponsiveLayout {
	width, _, err := pterm.GetTerminalSize()
	if err != nil {
		width = 80 // 默认值
	}

	// 检测终端颜色支持
	supportsColor := os.Getenv("NO_COLOR") == "" &&
		os.Getenv("FORCE_COLOR") != "" ||
		(os.Getenv("TERM") != "dumb" &&
			os.Getenv("NO_COLOR") == "")

	return ResponsiveLayout{
		MinWidth:     80,
		MaxWidth:     120,
		CompactMode:  width < 100,
		ShowDetails:  width >= 90,
		SupportsColor: supportsColor,
	}
}

// UIError 定义UI错误类型
type UIError struct {
	Type     ErrorType  // 错误类型
	Message  string     // 错误消息
	Recovery string     // 恢复建议
	Severity Severity   // 严重程度
	Timestamp time.Time // 错误时间
}

// ErrorType 错误类型枚举
type ErrorType int

const (
	ErrorTypeValidation ErrorType = iota // 验证错误
	ErrorTypeNetwork                     // 网络错误
	ErrorTypeConfig                      // 配置错误
	ErrorTypeSystem                      // 系统错误
	ErrorTypeUser                        // 用户操作错误
)

// Severity 错误严重程度
type Severity int

const (
	SeverityInfo     Severity = iota // 信息
	SeverityWarning                 // 警告
	SeverityError                   // 错误
	SeverityCritical                // 严重
)

// DisplayError 显示错误信息
func (e *UIError) DisplayError(theme *Theme) {
	layout := GetResponsiveLayout()

	// 根据严重程度选择颜色
	var color pterm.Color
	switch e.Severity {
	case SeverityInfo:
		color = theme.Colors.Info
	case SeverityWarning:
		color = theme.Colors.Warning
	case SeverityError:
		color = theme.Colors.Error
	case SeverityCritical:
		color = theme.Colors.Error
	default:
		color = theme.Colors.Muted
	}

	// 显示错误标题
	title := pterm.Sprintf("[%s] %s", color.Sprint("错误"), e.Message)
	fmt.Println(title)

	// 显示恢复建议（如果有）
	if e.Recovery != "" {
		pterm.Printf("%s %s\n", theme.Colors.Secondary.Sprint("建议:"), e.Recovery)
	}

	// 添加间距
	if layout.CompactMode {
		fmt.Println()
	} else {
		fmt.Println()
	}
}

// NewValidationError 创建验证错误
func NewValidationError(message string, recovery string) *UIError {
	return &UIError{
		Type:      ErrorTypeValidation,
		Message:   message,
		Recovery:  recovery,
		Severity:  SeverityError,
		Timestamp: time.Now(),
	}
}

// NewConfigError 创建配置错误
func NewConfigError(message string, recovery string) *UIError {
	return &UIError{
		Type:      ErrorTypeConfig,
		Message:   message,
		Recovery:  recovery,
		Severity:  SeverityError,
		Timestamp: time.Now(),
	}
}

// NewUserError 创建用户操作错误
func NewUserError(message string, recovery string) *UIError {
	return &UIError{
		Type:      ErrorTypeUser,
		Message:   message,
		Recovery:  recovery,
		Severity:  SeverityWarning,
		Timestamp: time.Now(),
	}
}

// DisplaySuccess 显示成功信息
func DisplaySuccess(message string, theme *Theme) {
	layout := GetResponsiveLayout()

	// 显示成功消息
	successMsg := pterm.Sprintf("[%s] %s", theme.Colors.Success.Sprint("成功"), message)
	pterm.Info.Println(successMsg)

	// 添加间距
	if !layout.CompactMode {
		fmt.Println()
	}
}

// DisplayInfo 显示信息消息
func DisplayInfo(message string, theme *Theme) {
	layout := GetResponsiveLayout()

	// 显示信息消息
	infoMsg := pterm.Sprintf("[%s] %s", theme.Colors.Info.Sprint("信息"), message)
	pterm.Info.Println(infoMsg)

	// 添加间距
	if !layout.CompactMode {
		fmt.Println()
	}
}

// DisplayWarning 显示警告信息
func DisplayWarning(message string, theme *Theme) {
	layout := GetResponsiveLayout()

	// 显示警告消息
	warningMsg := pterm.Sprintf("[%s] %s", theme.Colors.Warning.Sprint("警告"), message)
	pterm.Warning.Println(warningMsg)

	// 添加间距
	if !layout.CompactMode {
		fmt.Println()
	}
}

// Spacer 显示间距
func Spacer(size int, theme *Theme) {
	layout := GetResponsiveLayout()

	// 根据响应式布局调整间距
	actualSize := size
	if layout.CompactMode && size > 2 {
		actualSize = size / 2
	}

	for i := 0; i < actualSize; i++ {
		fmt.Println()
	}
}
