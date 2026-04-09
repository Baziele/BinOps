package main

import (
	"fmt"
	// "image/color"
	// "math"
	// "strings"
	// "time"

	tea "charm.land/bubbletea/v2"
)


type model struct {
	width  int
	height int
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		return m, tea.Quit
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}
	return m, nil
}

func (m model) View() tea.View {
	var v tea.View
	v.AltScreen = true
	if m.width == 0 {
		v.SetContent("Initializing...")
		return v
	}
	v.SetContent("Hello World")
	return v
}



func main() {
	p := tea.NewProgram(
		model{},
	)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v", err)
	}
}