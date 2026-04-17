package main

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type StringsModel struct {
	height int
	width  int
	styles struct {
		title       lipgloss.Style
		border      lipgloss.Style
		footerInfo  lipgloss.Style
		offset      lipgloss.Style
		value       lipgloss.Style
		valueAlt    lipgloss.Style
		emptyNotice lipgloss.Style
	}
	content  string
	viewport viewport.Model
}

func initializeStringsModel(width, height int, entries []ELFStringEntry) StringsModel {
	m := StringsModel{
		width:   width,
		height:  height,
		content: "",
	}
	m.styles.title = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("230")).Background(lipgloss.Color("63")).Padding(0, 1)
	m.styles.border = lipgloss.NewStyle().Foreground(lipgloss.Color("63"))
	m.styles.footerInfo = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("229")).Background(lipgloss.Color("63")).Padding(0, 1)
	m.styles.offset = lipgloss.NewStyle().Foreground(lipgloss.Color("117")).Bold(true)
	m.styles.value = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	m.styles.valueAlt = lipgloss.NewStyle().Foreground(lipgloss.Color("250"))
	m.styles.emptyNotice = lipgloss.NewStyle().Foreground(lipgloss.Color("244")).Italic(true)
	m.viewport = viewport.New(viewport.WithWidth(width), viewport.WithHeight(m.viewportHeight(height)))
	m.setStrings(entries)

	return m
}

func (m StringsModel) Init() tea.Cmd {

	return nil
}

func (m StringsModel) Update(msg tea.Msg) (StringsModel, tea.Cmd) {
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m StringsModel) View() string {

	// m.viewport.sethe
	return lipgloss.JoinVertical(lipgloss.Left, m.titleView(), m.viewport.View(), m.footerView())
}

func (m StringsModel) titleView() string {
	title := m.styles.title.Render("┌┤Strings├")
	line := m.styles.border.Render(strings.Repeat("─", max(0, m.width-lipgloss.Width(title)-1)))
	return lipgloss.JoinHorizontal(lipgloss.Center, title, line, m.styles.border.Render("┐"))
}

func (m StringsModel) footerView() string {
	info := m.styles.footerInfo.Render(lipgloss.Sprintf("%3.f%%:%3.f%%", m.viewport.ScrollPercent()*100, m.viewport.HorizontalScrollPercent()*100))
	line := m.styles.border.Render(strings.Repeat("─", max(0, m.viewport.Width()-lipgloss.Width(info)-3)))
	return lipgloss.JoinHorizontal(lipgloss.Center, line, m.styles.border.Render("┤"), info, m.styles.border.Render("├┘"))
}

func (m *StringsModel) setDimensions(width, height int) {
	m.width = width
	m.height = height
	m.viewport.SetWidth(width)
	m.viewport.SetHeight(m.viewportHeight(height))
}

func (m StringsModel) viewportHeight(totalHeight int) int {
	titleHeight := lipgloss.Height(m.styles.title.Render("┌┤Strings├"))
	footerHeight := lipgloss.Height(m.styles.footerInfo.Render("100%:100%"))
	return max(1, totalHeight-titleHeight-footerHeight)
}

func (m *StringsModel) setStrings(entries []ELFStringEntry) {
	if len(entries) == 0 {
		m.content = m.styles.emptyNotice.Render("No strings extracted.")
		m.viewport.SetContent(m.content)
		return
	}

	var b strings.Builder
	for i, entry := range entries {
		if i > 0 {
			b.WriteByte('\n')
		}
		offset := m.styles.offset.Render(fmt.Sprintf("0x%x", entry.Offset))
		valueStyle := m.styles.value
		if i%2 == 1 {
			valueStyle = m.styles.valueAlt
		}
		value := valueStyle.Render(entry.Value)
		b.WriteString(offset + " " + value)
	}
	m.content = b.String()
	m.viewport.SetContent(m.content)
}
