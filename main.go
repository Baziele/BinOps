package main

import (
	"fmt"
	"os"

	// "image/color"
	// "math"
	// "strings"
	// "time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type model struct {
	ready       bool
	isFileReady bool
	pages       []string
	currentPage int
	width       int
	height      int
	styles      struct {
		borderStyles lipgloss.Style
	}
	generalPage GeneralModel
	staticPage  StaticModel
	dynamicPage tea.Model
	stringsPage StringsModel
	hexdumpPage HexdumpModel
	binaryName  string
	elfAnalysis ELFAnalysis
}

// contentSize returns the inner dimensions for page content (inside border, below nav).
func (m model) contentSize() (w, h int) {
	navHeight := lipgloss.Height(m.navView())
	borderWidth := lipgloss.Width(m.styles.borderStyles.Render(""))
	borderHeight := lipgloss.Height(m.styles.borderStyles.Render(""))
	return m.width - borderWidth, m.height - navHeight - borderHeight
}

func initializeModel(binaryName string) model {
	m := model{pages: []string{"General", "Static", "Dynamic", "Strings", "Hexdump"}, binaryName: binaryName}
	m.styles.borderStyles = lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("39"))
	// m.generalPage = initializeGeneralModel()
	// m.staticPage = initializeStaticModel()
	// m.dynamicPage = initializeDynamicModel()
	// m.stringsPage = initializeStringsModel()
	// m.hexdumpPage = initializeHexdumpModel()
	return m
}

func (m model) Init() tea.Cmd {
	return analyzeBinary(m.binaryName)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch keypress := msg.String(); keypress {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "tab":
			if !m.isFileReady {
				return m, nil
			}
			m.currentPage = (m.currentPage + 1) % len(m.pages)
		case "shift+tab":
			if !m.isFileReady {
				return m, nil
			}
			m.currentPage = (m.currentPage - 1 + len(m.pages)) % len(m.pages)
		}
	case tea.WindowSizeMsg:
		navHeight := lipgloss.Height(m.navView())
		borderWidth := lipgloss.Width(m.styles.borderStyles.Render(""))
		borderHeight := lipgloss.Height(m.styles.borderStyles.Render(""))
		if !m.ready {
			m.width = msg.Width
			m.height = msg.Height
			contentWidth := m.width - borderWidth
			contentHeight := m.height - navHeight - borderHeight
			m.generalPage = initializeGeneralModel(contentWidth, contentHeight, m.elfAnalysis.Stats, m.elfAnalysis.Dependencies)
			m.staticPage = initializeStaticModel(contentWidth, contentHeight, m.elfAnalysis.File, m.elfAnalysis.Header, m.elfAnalysis.NoteSections, m.elfAnalysis.SegmentTables)
			m.dynamicPage = initializeDynamicModel()
			m.stringsPage = initializeStringsModel(contentWidth, contentHeight, m.elfAnalysis.Strings)
			m.hexdumpPage = initializeHexdumpModel(contentWidth, contentHeight, m.binaryName, m.elfAnalysis.FileBytes)
			m.ready = true
		}
		m.width = msg.Width
		m.height = msg.Height
		m.generalPage.setDimensions(m.width-borderWidth, m.height-navHeight-borderHeight)
		m.staticPage.setDimensions(m.width-borderWidth, m.height-navHeight-borderHeight)
		m.stringsPage.setDimensions(m.width-borderWidth, m.height-navHeight-borderHeight)
		m.hexdumpPage.setDimensions(m.width-borderWidth, m.height-navHeight-borderHeight)

	case ELFAnalysis:
		m.elfAnalysis = msg
		m.generalPage.fileStats = msg.Stats
		m.generalPage.dependencies = msg.Dependencies
		m.staticPage.elfFile = msg.File
		m.staticPage.elfHeader = msg.Header
		m.staticPage.elfNotes = msg.NoteSections
		m.staticPage.segmentTables = msg.SegmentTables
		m.staticPage.refreshFileSegmentTable()
		m.stringsPage.setStrings(msg.Strings)
		if m.ready && m.width > 0 {
			cw, ch := m.contentSize()
			m.hexdumpPage = initializeHexdumpModel(cw, ch, m.binaryName, msg.FileBytes)
		}
		m.isFileReady = true
	}

	switch m.currentPage {
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
	if m.width == 0 || !m.isFileReady {
		v.SetContent("Initializing...")
		return v
	}
	nav := m.navView()
	var currentView string
	switch m.currentPage {
	case 0:
		currentView = m.generalPage.View()
	case 1:
		currentView = m.staticPage.View()
	case 2:
		currentView = m.dynamicPage.View().Content
	case 3:
		currentView = m.stringsPage.View()
	case 4:
		currentView = m.hexdumpPage.View()
	}
	body := lipgloss.JoinVertical(lipgloss.Top, currentView)
	body = m.styles.borderStyles.Width(m.width).Height(m.height - lipgloss.Height(nav)).Render(body)
	v.SetContent(lipgloss.JoinVertical(lipgloss.Top, nav, body))
	return v
}

func (m model) navView() string {
	str := ""
	for i, page := range m.pages {
		isFirst, isCurrentPage := i == 0, i == m.currentPage
		style := lipgloss.NewStyle()
		if isCurrentPage {
			style = style.Bold(true)
		} else {
			style = style.Foreground(lipgloss.Color("#7D56F4"))
		}
		separator := ""
		if !isFirst {
			separator = " | "
		}
		str += style.Foreground(lipgloss.Color("#7D56F4")).Render(separator) + style.Render(page)
	}

	return m.styles.borderStyles.Width(m.width).Render(str)
}

func main() {
	p := tea.NewProgram(
		initializeModel(os.Args[1]),
	)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v", err)
	}
}
