package main

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// --- 图标定义 ---
const (
	IconRocket   = "🚀"
	IconServer   = "🖥️ "
	IconLock     = "🔒"
	IconLink     = "🔗"
	IconModel    = "🧠"
	IconFast     = "⚡"
	IconCheck    = "✓"
	IconStar     = "⭐"
	IconPin      = "📌"
	IconAdd      = "➕"
	IconWarn     = "⚠️ "
)

// --- 预设模版结构 ---
type VendorTemplate struct {
	Name        string
	Vendor      string
	BaseURL     string
	DefaultModel string
}

// 内置模版库
var vendorTemplates = []VendorTemplate{
	{Name: "Anthropic (官方接口)", Vendor: "Anthropic", BaseURL: "https://api.anthropic.com", DefaultModel: "claude-3-5-sonnet-20240620"},
	{Name: "MiniMax (海螺大模型)", Vendor: "MiniMax", BaseURL: "https://api.minimax.chat/v1", DefaultModel: "abab6.5-chat"},
	{Name: "DeepSeek (深度求索)", Vendor: "DeepSeek", BaseURL: "https://api.deepseek.com", DefaultModel: "deepseek-coder"},
	{Name: "Moonshot (Kimi)", Vendor: "Moonshot", BaseURL: "https://api.moonshot.cn/v1", DefaultModel: "moonshot-v1-8k"},
	{Name: "ZhipuAI (智谱GLM)", Vendor: "ZhipuAI", BaseURL: "https://open.bigmodel.cn/api/paas/v4", DefaultModel: "glm-4"},
	{Name: "Custom (自定义配置)", Vendor: "", BaseURL: "", DefaultModel: ""},
}

// --- 样式定义 ---
var (
	primaryColor   = lipgloss.Color("63")
	secondaryColor = lipgloss.Color("86")
	accentColor    = lipgloss.Color("205")
	mutedColor     = lipgloss.Color("240")
	warnColor      = lipgloss.Color("196")
	
docStyle = lipgloss.NewStyle().Margin(1, 2)
	
	listTitleStyle = lipgloss.NewStyle().
		Background(primaryColor).
		Foreground(lipgloss.Color("255")).
		Padding(0, 1).
		Bold(true)
	
	selectedItemStyle = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(accentColor).
		Foreground(accentColor).
		PaddingLeft(1)
	
	detailStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(mutedColor).
		Padding(1, 2)
	
	detailHeaderStyle = lipgloss.NewStyle().
		Foreground(secondaryColor).
		Bold(true).
		Border(lipgloss.NormalBorder(), false, false, true, false).
		BorderForeground(mutedColor).
		MarginBottom(1).
		PaddingBottom(0)
			
	detailLabelStyle = lipgloss.NewStyle().Foreground(mutedColor).Width(16)
	detailValueStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))

	badgeStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("230")).
		Background(lipgloss.Color("62")).
		Padding(0, 1).
		MarginRight(1).
		Bold(true)
	
	inputStyle = lipgloss.NewStyle().Foreground(accentColor)
	inputLabelStyle = lipgloss.NewStyle().Foreground(mutedColor).Width(18)
)

type item struct {
	platform Platform
}
func (i item) Title() string       { return i.platform.Title() }
func (i item) Description() string { return i.platform.DescriptionText() }
func (i item) FilterValue() string { return i.platform.FilterValue() }

type templateItem struct {
	template VendorTemplate
}
func (t templateItem) Title() string { return t.template.Name }
func (t templateItem) Description() string { return t.template.BaseURL }
func (t templateItem) FilterValue() string { return t.template.Name }


type model struct {
	list          list.Model
	viewport      viewport.Model
	spinner       spinner.Model
	platforms     []Platform
	selected      *Platform
	width, height int
	ready         bool
	
	state      appState
	loadingMsg string
	
	inputs     []textinput.Model
	focusIndex int
	isEditing  bool
	editIndex  int
	
	deleteTarget string
	templateList list.Model
}

type appState int

const (
	stateBrowsing appState = iota
	stateLoading
	stateQuitting
	stateTemplateSelect
	stateInput
	stateConfirmDelete
)

type tickMsg time.Time

func (m model) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		
		availableWidth := m.width - 4
		availableHeight := m.height - 2
		
		listWidth := int(float64(availableWidth) * 0.35)
		if listWidth < 25 { listWidth = 25 }
		
		detailWidth := availableWidth - listWidth - 2
		
m.list.SetSize(listWidth, availableHeight)
		m.viewport = viewport.New(detailWidth-4, availableHeight-2)
		if m.state == stateTemplateSelect {
			m.templateList.SetSize(m.width-4, m.height-4)
		}
		
m.updateViewport()

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" { return m, tea.Quit }

		if m.state == stateConfirmDelete {
			switch strings.ToLower(msg.String()) {
			case "y", "enter":
				m.deletePlatform(m.deleteTarget)
				m.state = stateBrowsing
				m.deleteTarget = ""
				return m, nil
			case "n", "esc":
				m.state = stateBrowsing
				m.deleteTarget = ""
				return m, nil
			}
			return m, nil
		}

		if m.state == stateTemplateSelect {
			switch msg.String() {
			case "esc":
				m.state = stateBrowsing
				return m, nil
			case "enter":
				if i, ok := m.templateList.SelectedItem().(templateItem); ok {
					m.state = stateInput
					m.initInputsFromTemplate(i.template)
				}
				return m, nil
			}
			m.templateList, cmd = m.templateList.Update(msg)
			return m, cmd
		}

		if m.state == stateInput {
			switch msg.String() {
			case "esc":
				if m.isEditing { m.state = stateBrowsing } else { m.state = stateTemplateSelect }
				return m, nil
			case "tab", "shift+tab", "enter", "up", "down":
				s := msg.String()
				if s == "enter" && m.focusIndex == len(m.inputs) {
					return m, m.submitInput()
				}
				if s == "up" || s == "shift+tab" { m.focusIndex-- } else { m.focusIndex++ }
				if m.focusIndex > len(m.inputs) { m.focusIndex = 0 } else if m.focusIndex < 0 { m.focusIndex = len(m.inputs) }
				
				cmds := make([]tea.Cmd, len(m.inputs))
				for i := 0; i <= len(m.inputs)-1; i++ {
					if i == m.focusIndex {
						cmds[i] = m.inputs[i].Focus()
						m.inputs[i].PromptStyle = inputStyle
						m.inputs[i].TextStyle = inputStyle
					} else {
						m.inputs[i].Blur()
						m.inputs[i].PromptStyle = lipgloss.NewStyle()
						m.inputs[i].TextStyle = lipgloss.NewStyle()
					}
				}
				return m, tea.Batch(cmds...)
			}
			cmd := m.updateInputs(msg)
			return m, cmd
		}
		
		if m.state == stateLoading { return m, nil }

		switch msg.String() {
		case "q":
			return m, tea.Quit
		case "enter":
			if i, ok := m.list.SelectedItem().(item); ok {
				m.selected = &i.platform
				m.state = stateLoading
				m.loadingMsg = fmt.Sprintf("正在连接到 %s...", i.platform.Name)
				return m, tea.Batch(m.spinner.Tick, tea.Tick(800*time.Millisecond, func(t time.Time) tea.Msg { return tickMsg(t) }))
			}
		case "p":
			if i, ok := m.list.SelectedItem().(item); ok {
				m.togglePin(i.platform.Name)
			}
			return m, nil
		case "x", "delete":
			if i, ok := m.list.SelectedItem().(item); ok {
				m.state = stateConfirmDelete
				m.deleteTarget = i.platform.Name
			}
			return m, nil
		case "a":
			m.state = stateTemplateSelect
			m.isEditing = false
			m.initTemplateList()
			return m, nil
		case "e":
			if i, ok := m.list.SelectedItem().(item); ok {
				m.state = stateInput
				m.isEditing = true
				m.editIndex = m.findPlatformIndex(i.platform.Name)
				m.initInputs(i.platform)
			}
			return m, nil
		}

	case tickMsg:
		if m.state == stateLoading {
			m.state = stateQuitting
			return m, tea.Quit
		}
		
	case spinner.TickMsg:
		if m.state == stateLoading {
			var spinnerCmd tea.Cmd
			m.spinner, spinnerCmd = m.spinner.Update(msg)
			return m, spinnerCmd
		}
	}

	if m.state == stateBrowsing {
		m.list, cmd = m.list.Update(msg)
		cmds = append(cmds, cmd)
		if m.list.SelectedItem() != nil {
			m.updateViewport()
		}
	}

	return m, tea.Batch(cmds...)
}

func (m *model) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))
	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}
	return tea.Batch(cmds...)
}

func (m *model) initTemplateList() {
	items := make([]list.Item, len(vendorTemplates))
	for i, t := range vendorTemplates {
		items[i] = templateItem{template: t}
	}
	
l := list.New(items, list.NewDefaultDelegate(), m.width-10, 14)
	l.Title = "🛠️  Step 1: 选择 Provider 供应商"
	l.Styles.Title = listTitleStyle.Copy().Background(accentColor)
	l.SetShowHelp(false)
	l.SetFilteringEnabled(true)
	
m.templateList = l
}

func (m *model) initInputsFromTemplate(t VendorTemplate) {
	p := Platform{
		Name:             strings.ToLower(strings.Split(t.Name, " ")[0]),
		Vendor:           t.Vendor,
		AnthropicBaseURL: t.BaseURL,
		AnthropicModel:   t.DefaultModel,
	}
	if t.Vendor == "" { p.Name = "" }
	
m.initInputs(p)
	
	if t.Vendor != "" {
		m.inputs[0].Blur()
		m.inputs[3].Focus() 
		m.inputs[3].PromptStyle = inputStyle
		m.inputs[3].TextStyle = inputStyle
		m.focusIndex = 3
	} else {
		m.inputs[0].Focus()
		m.focusIndex = 0
	}
}

func (m *model) initInputs(p Platform) {
	m.inputs = make([]textinput.Model, 5)
	var t textinput.Model
	
t = textinput.New(); t.Placeholder = "名称 (例如: dev-minimax)"; t.SetValue(p.Name); m.inputs[0] = t
t = textinput.New(); t.Placeholder = "厂商 (例如: MiniMax)"; t.SetValue(p.Vendor); m.inputs[1] = t
t = textinput.New(); t.Placeholder = "https://..."; t.SetValue(p.AnthropicBaseURL); m.inputs[2] = t
t = textinput.New(); t.Placeholder = "API Token"; t.EchoMode = textinput.EchoPassword; t.SetValue(p.AnthropicAuthToken); m.inputs[3] = t
t = textinput.New(); t.Placeholder = "Model ID"; t.SetValue(p.AnthropicModel); m.inputs[4] = t
	
	for i := range m.inputs {
		m.inputs[i].PromptStyle = lipgloss.NewStyle()
		m.inputs[i].TextStyle = lipgloss.NewStyle()
	}
}

func (m *model) submitInput() tea.Cmd {
	p := Platform{
		Name:               m.inputs[0].Value(),
		Vendor:             m.inputs[1].Value(),
		AnthropicBaseURL:   m.inputs[2].Value(),
		AnthropicAuthToken: m.inputs[3].Value(),
		AnthropicModel:     m.inputs[4].Value(),
	}
	if p.Name == "" { return nil }
	if m.isEditing {
		p.Pinned = m.platforms[m.editIndex].Pinned
		p.Description = m.platforms[m.editIndex].Description
		p.ExtraEnv = m.platforms[m.editIndex].ExtraEnv
		m.platforms[m.editIndex] = p
	} else {
		m.platforms = append(m.platforms, p)
	}
	SaveConfig(&Config{Platforms: m.platforms}, "")
	m.refreshList()
	m.state = stateBrowsing
	return nil
}

func (m *model) togglePin(name string) {
	idx := m.findPlatformIndex(name)
	if idx != -1 {
		m.platforms[idx].Pinned = !m.platforms[idx].Pinned
		SaveConfig(&Config{Platforms: m.platforms}, "")
		m.refreshList()
	}
}

func (m *model) deletePlatform(name string) {
	idx := m.findPlatformIndex(name)
	if idx != -1 {
		m.platforms = append(m.platforms[:idx], m.platforms[idx+1:]...)
		SaveConfig(&Config{Platforms: m.platforms}, "")
		m.refreshList()
	}
}

func (m *model) findPlatformIndex(name string) int {
	for i, p := range m.platforms { if p.Name == name { return i } }
	return -1
}

func (m *model) refreshList() {
	sort.Slice(m.platforms, func(i, j int) bool {
		if m.platforms[i].Pinned != m.platforms[j].Pinned { return m.platforms[i].Pinned }
		return m.platforms[i].Name < m.platforms[j].Name
	})
	items := make([]list.Item, len(m.platforms))
	for i, p := range m.platforms { items[i] = item{platform: p} }
	m.list.SetItems(items)
}

func (m *model) updateViewport() {
	sel := m.list.SelectedItem()
	if sel == nil {
		m.viewport.SetContent("未选中任何配置")
		return
	}
	p := sel.(item).platform
	
	var b strings.Builder
	b.WriteString(detailHeaderStyle.Render(fmt.Sprintf("%s %s", IconRocket, p.Name)))
	if p.Pinned { b.WriteString(" " + IconPin) }
	b.WriteString("\n\n")
	
	if p.Vendor != "" {
		b.WriteString(badgeStyle.Copy().Background(primaryColor).Render(p.Vendor))
	} else {
		b.WriteString(badgeStyle.Copy().Background(primaryColor).Render("Anthropic"))
	}
	if strings.Contains(strings.ToLower(p.AnthropicModel), "claude-3") {
		b.WriteString(badgeStyle.Copy().Background(lipgloss.Color("170")).Render("Smart"))
	}
	b.WriteString("\n\n")
	
	renderRow(&b, IconServer+" Vendor", p.Vendor)
	renderRow(&b, "📝 Description", p.Description)
	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().Foreground(accentColor).Render("🔧 Environment Variables"))
	b.WriteString("\n\n")
	renderRow(&b, IconLink+" Base URL", p.AnthropicBaseURL)
	renderRow(&b, IconModel+" Model", p.AnthropicModel)
	if p.AnthropicSmallModel != "" { renderRow(&b, IconFast+" Fast Model", p.AnthropicSmallModel) }
	renderRow(&b, IconLock+" Token", p.MaskedToken())
	
m.viewport.SetContent(b.String())
}

func renderRow(b *strings.Builder, label, value string) {
	if value == "" { return }
	b.WriteString(detailLabelStyle.Render(label))
	b.WriteString(detailValueStyle.Render(value))
	b.WriteString("\n")
}

func (m model) View() string {
	if !m.ready { return "初始化中..." }
	
	if m.state == stateLoading {
		str := fmt.Sprintf("\n\n   %s %s\n\n", m.spinner.View(), m.loadingMsg)
		box := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(accentColor).Padding(1, 4).Align(lipgloss.Center).Width(50).Height(8)
		content := box.Render(lipgloss.Place(50-2, 8-2, lipgloss.Center, lipgloss.Center, str))
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
	}
	
	if m.state == stateConfirmDelete {
		str := fmt.Sprintf("\n%s 警告\n\n确定要删除配置 [%s] 吗？\n\n[Y] 确认删除    [N] 取消操作", IconWarn, m.deleteTarget)
		box := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(warnColor).Padding(1, 4).Align(lipgloss.Center).Width(50).Height(10)
		content := box.Render(lipgloss.Place(50-2, 10-2, lipgloss.Center, lipgloss.Center, str))
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
	}
	
	if m.state == stateTemplateSelect {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, docStyle.Render(m.templateList.View()))
	}
	
	if m.state == stateInput {
		var b strings.Builder
		title := "🆕 Step 2: 详细配置"
		if m.isEditing { title = "✏️ 编辑配置" }
		
b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(primaryColor).Render(title))
		b.WriteString("\n\n")
		
		labels := []string{"Name (名称)", "Vendor (厂商)", "Base URL (地址)", "Token (令牌)", "Model (模型)"}
		for i := 0; i < len(m.inputs); i++ {
			b.WriteString(inputLabelStyle.Render(labels[i]))
			b.WriteString(m.inputs[i].View())
			b.WriteString("\n")
		}
		
b.WriteString("\n")
		btn := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("[ 保存 ]")
		if m.focusIndex == len(m.inputs) {
			btn = lipgloss.NewStyle().Foreground(accentColor).Bold(true).Render("[ 保存 ]")
		}
		b.WriteString(btn)
		
		box := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1, 2).Width(75)
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box.Render(b.String()))
	}
	
	if m.state == stateQuitting { return "" }

	listView := m.list.View()
	detailView := detailStyle.Width(m.viewport.Width + 2).Height(m.list.Height()).Render(m.viewport.View())
	
	help := lipgloss.NewStyle().Foreground(mutedColor).MarginTop(1).Render(
		"↑/↓: 导航 • Enter: 选择 • a: 新增 • e: 编辑 • p: 置顶 • x: 删除 • q: 退出",
	)

	return docStyle.Render(lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.JoinHorizontal(lipgloss.Top, listView, "  ", detailView),
		help,
	))
}

func SelectPlatform(platforms []Platform) (*Platform, error) {
	sort.Slice(platforms, func(i, j int) bool {
		if platforms[i].Pinned != platforms[j].Pinned { return platforms[i].Pinned }
		return platforms[i].Name < platforms[j].Name
	})

	items := make([]list.Item, len(platforms))
	for i, p := range platforms { items[i] = item{platform: p} }

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "🚀 CC Gate (Claude Code 代理管理器)"
	l.SetShowHelp(false)
	l.Styles.Title = listTitleStyle
	
d := list.NewDefaultDelegate()
	d.Styles.SelectedTitle = selectedItemStyle
	d.Styles.SelectedDesc = selectedItemStyle.Copy().Foreground(secondaryColor)
	l.SetDelegate(d)
	
s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(accentColor)

	m := model{list: l, platforms: platforms, spinner: s, state: stateBrowsing}

	p := tea.NewProgram(m, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil { return nil, err }

	if m, ok := finalModel.(model); ok && m.selected != nil { return m.selected, nil }
	return nil, fmt.Errorf("未选择任何平台")
}