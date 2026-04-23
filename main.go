package main

import (
	"fmt"
	"os"

	// "image/color"
	// "math"
	// "strings"
	// "time"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/colorprofile"
)

const BUILDERSIO_URL = "0.1.0"

type model struct {
	ready       bool
	isFileReady bool
	pages       []string
	currentPage int
	width       int
	height      int
	theme       Theme

	styles struct {
		borderStyles lipgloss.Style
		navStyles    lipgloss.Style
	}
	generalPage GeneralModel
	staticPage  StaticModel
	dynamicPage DynamicModel
	stringsPage StringsModel
	hexdumpPage HexdumpModel
	binaryName  string
	elfAnalysis ELFAnalysis
}

// contentSize returns the inner dimensions for page content (inside border, below nav).
func (m model) contentSize() (w, h int) {
	navHeight := lipgloss.Height(m.navView())
	return m.width - m.styles.borderStyles.GetHorizontalFrameSize(),
		m.height - navHeight - m.styles.borderStyles.GetVerticalFrameSize()
}

func initializeModel(binaryName string) model {
	m := model{
		pages:      []string{"General", "Static", "Dynamic", "Strings", "Hexdump"},
		binaryName: binaryName,
		theme:      GetTheme(ActiveTheme),
	}
	m.styles.borderStyles = lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(m.theme.Border).
		Margin(0, 1, 1, 1)
	m.styles.navStyles = lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(m.theme.Border).
		Padding(0, 1).
		Margin(1, 1, 0, 1)
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
		switch {
		case key.Matches(msg, DefaultAppKeyMap.Quit):
			return m, tea.Quit
		case key.Matches(msg, DefaultAppKeyMap.NextPage):
			if !m.isFileReady {
				return m, nil
			}
			m.currentPage = (m.currentPage + 1) % len(m.pages)
		case key.Matches(msg, DefaultAppKeyMap.PrevPage):
			if !m.isFileReady {
				return m, nil
			}
			m.currentPage = (m.currentPage - 1 + len(m.pages)) % len(m.pages)
		}
	case tea.WindowSizeMsg:
		navHeight := lipgloss.Height(m.navView())
		contentWidth := msg.Width - m.styles.borderStyles.GetHorizontalFrameSize()
		contentHeight := msg.Height - navHeight - m.styles.borderStyles.GetVerticalFrameSize()
		if !m.ready {
			m.width = msg.Width
			m.height = msg.Height
			m.generalPage = initializeGeneralModel(contentWidth, contentHeight, m.elfAnalysis.Stats, m.elfAnalysis.Dependencies, m.theme)
			m.staticPage = initializeStaticModel(contentWidth, contentHeight, m.elfAnalysis.File, m.elfAnalysis.Header, m.elfAnalysis.NoteSections, m.elfAnalysis.SegmentTables, m.theme)
			m.dynamicPage = initializeDynamicModel(contentWidth, contentHeight, m.theme)
			m.stringsPage = initializeStringsModel(contentWidth, contentHeight, m.elfAnalysis.Strings, m.theme)
			m.hexdumpPage = initializeHexdumpModel(contentWidth, contentHeight, m.binaryName, m.elfAnalysis.FileBytes, m.theme)
			m.ready = true
		}
		m.width = msg.Width
		m.height = msg.Height
		m.generalPage.setDimensions(contentWidth, contentHeight)
		m.staticPage.setDimensions(contentWidth, contentHeight)
		m.dynamicPage.setDimensions(contentWidth, contentHeight)
		m.stringsPage.setDimensions(contentWidth, contentHeight)
		m.hexdumpPage.setDimensions(contentWidth, contentHeight)

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
			m.hexdumpPage = initializeHexdumpModel(cw, ch, m.binaryName, msg.FileBytes, m.theme)
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
		currentView = m.dynamicPage.View()
	case 3:
		currentView = m.stringsPage.View()
	case 4:
		currentView = m.hexdumpPage.View()
	}
	body := lipgloss.JoinVertical(lipgloss.Top, currentView)
	body = m.styles.borderStyles.
		Width(m.width - m.styles.borderStyles.GetHorizontalMargins()).
		Height(m.height - lipgloss.Height(nav) - m.styles.borderStyles.GetVerticalMargins()).
		Render(body)
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
			style = style.Foreground(m.theme.NavAccent)
		}
		separator := ""
		if !isFirst {
			separator = " | "
		}
		str += style.Foreground(m.theme.NavAccent).Render(separator) + style.Render(page)
	}

	return m.styles.navStyles.Width(m.width - 2).Render(str)
}

func main() {
	opts := []tea.ProgramOption{}
	if os.Getenv("BINOPS_FORCE_COLOR") == "1" {
		opts = append(opts, tea.WithColorProfile(colorprofile.TrueColor))
	}

	p := tea.NewProgram(initializeModel(os.Args[1]), opts...)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v", err)
	}
}
