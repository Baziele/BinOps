package main

import (
	"fmt"
	// "image/color"
	// "math"
	// "strings"
	// "time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)


type model struct {
	pages []string
	currentPage int
	width  int
	height int
	styles struct{
		boarderStyles lipgloss.Style
	}
	generalPage tea.Model
	staticPage tea.Model
	dynamicPage tea.Model
	stringsPage tea.Model
	hexdumpPage tea.Model
}

func initializeModel() model{
	m := model{pages: []string{"General", "Static", "Dynamic", "Strings", "Hexdump"}}
	m.styles.boarderStyles = lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("39"))
	m.generalPage = initializeGeneralModel()
	m.staticPage = initializeStaticModel()
	m.dynamicPage = initializeDynamicModel()
	m.stringsPage = initializeStringsModel()
	m.hexdumpPage = initializeHexdumpModel()
	return m
}


func (m model) Init() tea.Cmd {
	return m.generalPage.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch keypress := msg.String(); keypress {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "tab":
			m.currentPage = (m.currentPage + 1) % len(m.pages)
		case "shift+tab":
			m.currentPage = (m.currentPage - 1 + len(m.pages)) % len(m.pages)
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	switch m.currentPage{
	case 0:
		newGeneralPage, newCmd := m.generalPage.Update(msg)
		m.generalPage = newGeneralPage
		cmd = newCmd
	case 1:
		newStaticPage, newCmd := m.staticPage.Update(msg)
		m.staticPage = newStaticPage
		cmd = newCmd
	case 2:
		newDynamicPage, newCmd := m.dynamicPage.Update(msg)
		m.dynamicPage = newDynamicPage
		cmd = newCmd
	case 3:
		newStringsPage, newCmd := m.stringsPage.Update(msg)
		m.stringsPage = newStringsPage
		cmd = newCmd
	case 4:
		newHexdumpPage, newCmd := m.hexdumpPage.Update(msg)
		m.hexdumpPage = newHexdumpPage
		cmd = newCmd
	}

	return m, cmd
}

func (m model) View() tea.View {
	var v tea.View
	v.AltScreen = true
	if m.width == 0 {
		v.SetContent("Initializing...")
		return v
	}
	nav := m.navView()
	var currentView tea.View
	switch m.currentPage {
	case 0:
		currentView = m.generalPage.View()
	case 1:
		currentView = m.staticPage.View()
	case 2:
		currentView = m.dynamicPage.View()
	case 3:
		currentView = m.stringsPage.View()
	case 4:
		currentView = m.hexdumpPage.View()
	}
	body := lipgloss.JoinVertical(lipgloss.Top, currentView.Content)
	body = m.styles.boarderStyles.Width(m.width).Height(m.height - lipgloss.Height(nav)).Render(body)
	v.SetContent(lipgloss.JoinVertical(lipgloss.Top, nav, body))
	return v
}

func (m model) navView() string{
	str := ""
	for i, page := range(m.pages){
		isFirst, isCurrentPage := i == 0, i == m.currentPage
		style := lipgloss.NewStyle()
		if isCurrentPage {
			style = style.Bold(true)
		}else{
			style = style.Foreground(lipgloss.Color("#7D56F4"))
		}
		separator := ""
		if !isFirst{
			separator = " | "
		}
		str += style.Foreground(lipgloss.Color("#7D56F4")).Render(separator) + style.Render(  page)
	}

	return m.styles.boarderStyles.Width(m.width).Render(str)
}



func main() {
	p := tea.NewProgram(
		initializeModel(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v", err)
	}
}