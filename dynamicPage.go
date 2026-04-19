package main

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type DynamicModel struct {
	width   int
	height  int
	theme   Theme
	content string
}

func initializeDynamicModel(width, height int, theme Theme) DynamicModel {
	return DynamicModel{
		width:   width,
		height:  height,
		theme:   theme,
		content: "Dynamic",
	}
}

func (m DynamicModel) Init() tea.Cmd {
	return nil
}

func (m DynamicModel) Update(msg tea.Msg) (DynamicModel, tea.Cmd) {
	return m, nil
}

func (m DynamicModel) View() string {
	body := lipgloss.Place(m.width, m.contentHeight(), lipgloss.Center, lipgloss.Center, m.content)
	return lipgloss.JoinVertical(lipgloss.Left, body, m.helpView())
}

func (m *DynamicModel) setDimensions(width, height int) {
	m.width = width
	m.height = height
}

func (m DynamicModel) contentHeight() int {
	return max(1, m.height-helpFooterHeight(m.theme, m.width, DefaultAppKeyMap))
}

func (m DynamicModel) helpView() string {
	return renderHelpFooter(m.theme, m.width, DefaultAppKeyMap)
}
