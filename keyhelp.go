package main

import (
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/lipgloss/v2"
)

type shortHelpProvider interface {
	ShortHelp() []key.Binding
}

type AppKeyMap struct {
	NextPage key.Binding
	PrevPage key.Binding
	Quit     key.Binding
}

var DefaultAppKeyMap = AppKeyMap{
	NextPage: key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "next page")),
	PrevPage: key.NewBinding(key.WithKeys("shift+tab"), key.WithHelp("shift+tab", "prev page")),
	Quit:     key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
}

func (km AppKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		shortHelpBinding("tab/s-tab", "pages"),
		shortHelpBinding("q", "quit"),
	}
}

func (km AppKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{km.ShortHelp()}
}

func shortHelpBinding(label, desc string) key.Binding {
	return key.NewBinding(
		key.WithKeys(label),
		key.WithHelp(label, desc),
	)
}

func footerContentText(info string, maps ...shortHelpProvider) string {
	var parts []string
	for _, km := range maps {
		if km == nil {
			continue
		}
		for _, binding := range km.ShortHelp() {
			if !binding.Enabled() {
				continue
			}
			item := strings.TrimSpace(binding.Help().Key + " " + binding.Help().Desc)
			if item != "" {
				parts = append(parts, item)
			}
		}
	}
	if info != "" {
		parts = append(parts, info)
	}
	return strings.Join(parts, " • ")
}

func renderHelpFooter(theme Theme, width int, maps ...shortHelpProvider) string {
	if width <= 0 {
		return ""
	}

	text := footerContentText("", maps...)
	if text == "" {
		return ""
	}

	maxContentWidth := max(1, width)
	text = clipVisual(text, maxContentWidth)
	return lipgloss.NewStyle().
		Width(width).
		Foreground(theme.MutedStrong).
		Render(text)
}

func helpFooterHeight(theme Theme, width int, maps ...shortHelpProvider) int {
	return lipgloss.Height(renderHelpFooter(theme, width, maps...))
}
