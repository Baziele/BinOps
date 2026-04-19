package main

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/mattn/go-runewidth"
)

type StringsModel struct {
	height int
	width  int
	theme  Theme
	styles struct {
		title       lipgloss.Style
		border      lipgloss.Style
		footerInfo  lipgloss.Style
		filterLabel lipgloss.Style
		filterValue lipgloss.Style
		filterHint  lipgloss.Style
		offset      lipgloss.Style
		value       lipgloss.Style
		valueAlt    lipgloss.Style
		emptyNotice lipgloss.Style
	}
	content    string
	allEntries []ELFStringEntry
	minLength  int
	shownCount int
	viewport   viewport.Model
}

func initializeStringsModel(width, height int, entries []ELFStringEntry, theme Theme) StringsModel {
	m := StringsModel{
		width:     width,
		height:    height,
		theme:     theme,
		content:   "",
		minLength: 15,
	}
	m.styles.title = lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.StringsTitleFG)
	m.styles.border = lipgloss.NewStyle().Foreground(theme.StringsTitleBG)
	m.styles.footerInfo = lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.StringsFooterFG).
		Background(theme.StringsFooterBG).
		Padding(0, 1)
	m.styles.filterLabel = lipgloss.NewStyle().Bold(true).Foreground(theme.StringsTitleFG)
	m.styles.filterValue = lipgloss.NewStyle().Bold(true).Foreground(theme.Warning)
	m.styles.filterHint = lipgloss.NewStyle().Foreground(theme.Muted)
	m.styles.offset = lipgloss.NewStyle().Foreground(theme.StringsOffset).Bold(true)
	m.styles.value = lipgloss.NewStyle().Foreground(theme.Body)
	m.styles.valueAlt = lipgloss.NewStyle().Foreground(theme.BodyAlt)
	m.styles.emptyNotice = lipgloss.NewStyle().Foreground(theme.Muted).Italic(true)
	m.viewport = viewport.New(viewport.WithWidth(m.viewportWidth()), viewport.WithHeight(m.viewportHeight(height)))
	m.setStrings(entries)

	return m
}

func (m StringsModel) Init() tea.Cmd {

	return nil
}

func (m StringsModel) Update(msg tea.Msg) (StringsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "-":
			if m.minLength > 1 {
				m.minLength--
				m.refreshContent()
			}
			return m, nil
		case "+", "=":
			m.minLength++
			m.refreshContent()
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m StringsModel) View() string {

	body := lipgloss.NewStyle().
		Width(m.width).
		Padding(0, 1).
		Render(m.viewport.View())
	return lipgloss.JoinVertical(lipgloss.Left, m.titleView(), body, m.footerView())
}

func (m StringsModel) titleView() string {
	title := m.styles.border.Render("┌|") + m.styles.title.Render("Strings") + m.styles.border.Render("|")
	filter := lipgloss.JoinHorizontal(
		lipgloss.Left,
		m.styles.filterLabel.Render("Min Length"),
		m.styles.border.Render(": "),
		m.styles.filterValue.Render(fmt.Sprintf("%d", m.minLength)),
		m.styles.filterHint.Render("  -/+"),
	)
	line := m.styles.border.Render(strings.Repeat("─", max(0, m.width-lipgloss.Width(title)-lipgloss.Width(filter)-1)))
	return lipgloss.JoinHorizontal(lipgloss.Center, title, line, filter, m.styles.border.Render("┐"))
}

func (m StringsModel) footerView() string {
	info := m.styles.footerInfo.Render(fmt.Sprintf("%d shown | %3.0f%%", m.shownCount, m.viewport.ScrollPercent()*100))
	line := m.styles.border.Render(strings.Repeat("─", max(0, m.width-lipgloss.Width(info)-3)))
	return lipgloss.JoinHorizontal(lipgloss.Center, line, m.styles.border.Render("┤"), info, m.styles.border.Render("├┘"))
}

func (m *StringsModel) setDimensions(width, height int) {
	m.width = width
	m.height = height
	m.viewport.SetWidth(m.viewportWidth())
	m.viewport.SetHeight(m.viewportHeight(height))
	m.refreshContent()
}

func (m StringsModel) viewportHeight(totalHeight int) int {
	titleHeight := lipgloss.Height(m.styles.title.Render("┌|Strings|"))
	footerHeight := lipgloss.Height(m.styles.footerInfo.Render("99999 shown | 100%"))
	return max(1, totalHeight-titleHeight-footerHeight)
}

func (m *StringsModel) setStrings(entries []ELFStringEntry) {
	m.allEntries = append(m.allEntries[:0], entries...)
	m.refreshContent()
}

func (m StringsModel) viewportWidth() int {
	return max(1, m.width-2)
}

func (m *StringsModel) refreshContent() {
	if len(m.allEntries) == 0 {
		m.shownCount = 0
		m.content = m.styles.emptyNotice.Render("No strings extracted.")
		m.viewport.SetContent(m.content)
		return
	}

	contentWidth := m.viewportWidth()
	var b strings.Builder
	m.shownCount = 0
	for _, entry := range m.allEntries {
		if len(entry.Value) < m.minLength {
			continue
		}
		if m.shownCount > 0 {
			b.WriteByte('\n')
		}
		offset := fmt.Sprintf("0x%x", entry.Offset)
		valueStyle := m.styles.value
		if m.shownCount%2 == 1 {
			valueStyle = m.styles.valueAlt
		}
		b.WriteString(renderWrappedStringEntry(offset+" ", entry.Value, contentWidth, m.styles.offset, valueStyle))
		m.shownCount++
	}

	if m.shownCount == 0 {
		m.content = m.styles.emptyNotice.Render(fmt.Sprintf("No strings found with min length >= %d.", m.minLength))
		m.viewport.SetContent(m.content)
		return
	}

	m.content = b.String()
	m.viewport.SetContent(m.content)
}

func renderWrappedStringEntry(prefix, value string, width int, prefixStyle, valueStyle lipgloss.Style) string {
	usableWidth := max(1, width)
	prefixWidth := runewidth.StringWidth(prefix)
	if prefixWidth >= usableWidth {
		prefixWidth = 0
		prefix = ""
	}

	lines := wrapExactVisualText(value, max(1, usableWidth-prefixWidth))
	if len(lines) == 0 {
		lines = []string{""}
	}

	var out strings.Builder
	continuationPrefix := strings.Repeat(" ", prefixWidth)
	for i, line := range lines {
		if i > 0 {
			out.WriteByte('\n')
		}
		if i == 0 {
			out.WriteString(prefixStyle.Render(prefix))
		} else if prefixWidth > 0 {
			out.WriteString(continuationPrefix)
		}
		out.WriteString(valueStyle.Render(line))
	}

	return out.String()
}

func wrapExactVisualText(s string, width int) []string {
	if s == "" {
		return []string{""}
	}
	if width <= 0 {
		return []string{s}
	}

	var lines []string
	var current strings.Builder
	currentWidth := 0

	for _, r := range s {
		if r == '\n' {
			lines = append(lines, current.String())
			current.Reset()
			currentWidth = 0
			continue
		}

		rw := runewidth.RuneWidth(r)
		if rw <= 0 {
			rw = 1
		}
		if currentWidth+rw > width && current.Len() > 0 {
			lines = append(lines, current.String())
			current.Reset()
			currentWidth = 0
		}

		current.WriteRune(r)
		currentWidth += rw
	}

	lines = append(lines, current.String())
	return lines
}
