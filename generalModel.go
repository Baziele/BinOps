package main

import (
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"
)

type GeneralModel struct {
	width int
	height int
	applicationName string
	description string 
	repository string
	creator string 
	filename string
	fileStats ELFStats
	dependencies []ELFDependency
	dependenciesSelected int
	dependenciesOffset int
	// propertiesViewport viewport.Model
	// dependenciesViewport viewport.Model
	content string
	styles struct {
		title lipgloss.Style
		stats lipgloss.Style
		dependencies lipgloss.Style
		label lipgloss.Style
	}
}

func initializeGeneralModel(width, height int, fileStats ELFStats, dependencies []ELFDependency) GeneralModel {
	m := GeneralModel{
		width: width,
		height: height,
		content: "General",
		applicationName: "BinOps",
		description: "The greatest static binary analysis too of all time",
		repository: "https://github.com/baziele/binops",
		creator: "@baziele",
		filename: "maliciousFile.exe",
		fileStats: fileStats,
		dependencies: dependencies,
		// propertiesViewport: viewport.New(viewport.WithWidth(40), viewport.WithHeight(15)),
		// dependenciesViewport: viewport.New(viewport.WithWidth(60), viewport.WithHeight(7)),
	}
	m.styles.title = lipgloss.NewStyle().Bold(true)	
	m.styles.stats = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderTop(false)//.Width(min(40, width-2)).Height(min(15, height-2))
	m.styles.dependencies = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderTop(false)//.Width(min(60, width-2)).Height(min(7, height-2))
	m.styles.label = lipgloss.NewStyle().Foreground(lipgloss.Cyan)
	// m.propertiesViewport.SetContent(m.propertiesView())
	// m.dependenciesViewport.SetContent(m.dependenciesView())

	return m
}


func (m GeneralModel) Init() tea.Cmd{
	return nil
}

func (m GeneralModel) Update(msg tea.Msg) (GeneralModel, tea.Cmd){
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if len(m.dependencies) == 0 {
			return m, nil
		}
		visibleDataRows := m.dependenciesDataRowsVisible()
		switch msg.String() {
		case "down", "j":
			if m.dependenciesSelected < len(m.dependencies)-1 {
				m.dependenciesSelected++
			}
		case "up", "k":
			if m.dependenciesSelected > 0 {
				m.dependenciesSelected--
			}
		}
		maxOffset := max(0, len(m.dependencies)-visibleDataRows)
		if m.dependenciesSelected < m.dependenciesOffset {
			m.dependenciesOffset = m.dependenciesSelected
		}
		if m.dependenciesSelected >= m.dependenciesOffset+visibleDataRows {
			m.dependenciesOffset = m.dependenciesSelected - visibleDataRows + 1
		}
		m.dependenciesOffset = min(max(0, m.dependenciesOffset), maxOffset)
	}
	return m, nil
}


func (m GeneralModel) View() string{
	
	applicationName := lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4")).Render(m.applicationName)
	description := lipgloss.NewStyle().Underline(true).Render(m.description) //BorderBottom(true)
	repository := lipgloss.NewStyle().Hyperlink(m.repository).Render(m.repository)
	creator := lipgloss.NewStyle().Italic(true).Foreground(lipgloss.Color("#7D56F4")).Render(m.creator)
	filename := lipgloss.NewStyle().Render(m.filename)
	properties := m.statsView()
	dependencies := m.dependenciesView()	

	v := lipgloss.JoinVertical(lipgloss.Center, applicationName, description, repository, creator, filename, properties, dependencies)
	v = lipgloss.PlaceHorizontal(m.width , lipgloss.Center, v)
	return v
}

func (m *GeneralModel) setDimensions(width, height int){
	m.width = width
	m.height = height
	// m.styles.stats.Width(min(40, width-2)).Height(min(15, height-2))
	// m.styles.dependencies.Width(min(60, width-2)).Height(min(7, height-2))
}

func (m GeneralModel) statsView() string {
	var contents strings.Builder
	write := func(label, value string) {
		contents.WriteString(m.styles.label.Render(label) + value)
		contents.WriteByte('\n')
	}
	write("File Size: ", lipgloss.Sprintf("%d bytes", m.fileStats.FileSize))
	write("Blocks: ", lipgloss.Sprintf("%d", m.fileStats.Blocks))
	write("Block Size: ", lipgloss.Sprintf("%d bytes", m.fileStats.BlockSize))
	write("Links: ", lipgloss.Sprintf("%d", m.fileStats.Links))
	write("UID: ", lipgloss.Sprintf("%d", m.fileStats.UID))
	write("GID: ", lipgloss.Sprintf("%d", m.fileStats.GID))
	write("Access: ", lipgloss.Sprintf("%s", time.Unix(m.fileStats.LastAccessTime, 0).Format(time.RFC3339)))
	write("Modified: ", lipgloss.Sprintf("%s", time.Unix(m.fileStats.LastModificationTime, 0).Format(time.RFC3339)))
	
	fullContent := m.styles.stats.Render(contents.String())
	title := m.styles.title.Render("┌┤Stats├")
	line := strings.Repeat("─", max(0, lipgloss.Width(fullContent) -lipgloss.Width(title) - 1))
	line = lipgloss.JoinHorizontal(lipgloss.Center, title, line, "┐")
	return lipgloss.JoinVertical(lipgloss.Left, line, fullContent)

}

func (m GeneralModel) dependenciesView() string {
	visibleRows := m.dependenciesVisibleRows()
	t := table.New().
		Headers("Library", "Path").
		BorderTop(false).
		BorderBottom(false).
		BorderLeft(false).
		BorderRight(false).
		BorderHeader(false).
		BorderColumn(false).
		BorderRow(false).
		Height(visibleRows).
		YOffset(m.dependenciesOffset).
		StyleFunc(func(row, _ int) lipgloss.Style {
			if row == table.HeaderRow {
				return lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Cyan)
			}
			if row == m.dependenciesSelected {
				return lipgloss.NewStyle().
					Foreground(lipgloss.Color("230")).
					Background(lipgloss.Color("62"))
			}
			return lipgloss.NewStyle()
		})

	if len(m.dependencies) == 0 {
		t.Row("No dependencies reported", "not found")
	} else {
		for _, dependency := range m.dependencies {
			path := dependency.LibraryPath
			if !dependency.Found {
				path = "not found"
			}
			t.Row(dependency.LibraryName, path)
		}
	}

	fullContent := m.styles.dependencies.Render(t.Render())
	title := m.styles.title.Render("┌┤Dependencies├")
	line := strings.Repeat("─", max(0, lipgloss.Width(fullContent) -lipgloss.Width(title) - 1))
	line = lipgloss.JoinHorizontal(lipgloss.Center, title, line, "┐")
	return lipgloss.JoinVertical(lipgloss.Left, line, fullContent)
}

func (m GeneralModel) dependenciesVisibleRows() int {
	// Reserve a reasonable slice of the page for dependencies while allowing scrolling.
	return max(4, min(12, m.height/3))
}

func (m GeneralModel) dependenciesDataRowsVisible() int {
	// The table header consumes one visible line.
	return max(1, m.dependenciesVisibleRows()-1)
}