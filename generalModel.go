package main

import (
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"
)

type GeneralModel struct {
	width                int
	height               int
	theme                Theme
	applicationName      string
	description          string
	repository           string
	creator              string
	fileStats            ELFStats
	dependencies         []ELFDependency
	dependenciesSelected int
	dependenciesOffset   int
	// propertiesViewport viewport.Model
	// dependenciesViewport viewport.Model
	content string
	styles  struct {
		title        lipgloss.Style
		stats        lipgloss.Style
		dependencies lipgloss.Style
		label        lipgloss.Style
		description  lipgloss.Style
		creator      lipgloss.Style
		depHeader    lipgloss.Style
		depSelected  lipgloss.Style
	}
}

func initializeGeneralModel(width, height int, fileStats ELFStats, dependencies []ELFDependency, theme Theme) GeneralModel {
	m := GeneralModel{
		width:           width,
		height:          height,
		theme:           theme,
		content:         "General",
		// applicationName: "BinOps",
		description:     "The greatest static binary analysis too of all time",
		repository:      "https://github.com/baziele/binops",
		creator:         "@baziele",
		fileStats:       fileStats,
		dependencies:    dependencies,
		// propertiesViewport: viewport.New(viewport.WithWidth(40), viewport.WithHeight(15)),
		// dependenciesViewport: viewport.New(viewport.WithWidth(60), viewport.WithHeight(7)),
	}

// 	m.applicationName = `
// █████      ███                                       
// ▒▒███      ▒▒▒                                        
//  ▒███████  ████  ████████    ██████  ████████   █████ 
//  ▒███▒▒███▒▒███ ▒▒███▒▒███  ███▒▒███▒▒███▒▒███ ███▒▒  
//  ▒███ ▒███ ▒███  ▒███ ▒███ ▒███ ▒███ ▒███ ▒███▒▒█████ 
//  ▒███ ▒███ ▒███  ▒███ ▒███ ▒███ ▒███ ▒███ ▒███ ▒▒▒▒███
//  ████████  █████ ████ █████▒▒██████  ▒███████  ██████ 
// ▒▒▒▒▒▒▒▒  ▒▒▒▒▒ ▒▒▒▒ ▒▒▒▒▒  ▒▒▒▒▒▒   ▒███▒▒▒  ▒▒▒▒▒▒  
//                                      ▒███             
//                                      █████            
//                                     ▒▒▒▒▒             
// 	`

// 	 m.applicationName = `
//  ███████████  █████ ██████   █████    ███████    ███████████   █████████ 
// ░░███░░░░░███░░███ ░░██████ ░░███   ███░░░░░███ ░░███░░░░░███ ███░░░░░███
//  ░███    ░███ ░███  ░███░███ ░███  ███     ░░███ ░███    ░███░███    ░░░ 
//  ░██████████  ░███  ░███░░███░███ ░███      ░███ ░██████████ ░░█████████ 
//  ░███░░░░░███ ░███  ░███ ░░██████ ░███      ░███ ░███░░░░░░   ░░░░░░░░███
//  ░███    ░███ ░███  ░███  ░░█████ ░░███     ███  ░███         ███    ░███
//  ███████████  █████ █████  ░░█████ ░░░███████░   █████       ░░█████████ 
// ░░░░░░░░░░░  ░░░░░ ░░░░░    ░░░░░    ░░░░░░░    ░░░░░         ░░░░░░░░░  
// 	`
	
// 	m.applicationName = `
// ░▒▓███████▓▒░░▒▓█▓▒░▒▓███████▓▒░ ░▒▓██████▓▒░░▒▓███████▓▒░ ░▒▓███████▓▒░ 
// ░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░        
// ░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░        
// ░▒▓███████▓▒░░▒▓█▓▒░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░░▒▓█▓▒░▒▓███████▓▒░ ░▒▓██████▓▒░  
// ░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░             ░▒▓█▓▒░ 
// ░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░             ░▒▓█▓▒░ 
// ░▒▓███████▓▒░░▒▓█▓▒░▒▓█▓▒░░▒▓█▓▒░░▒▓██████▓▒░░▒▓█▓▒░      ░▒▓███████▓▒░  
	// `

// 	m.applicationName = `
// ██████╗ ██╗███╗   ██╗ ██████╗ ██████╗ ███████╗
// ██╔══██╗██║████╗  ██║██╔═══██╗██╔══██╗██╔════╝
// ██████╔╝██║██╔██╗ ██║██║   ██║██████╔╝███████╗
// ██╔══██╗██║██║╚██╗██║██║   ██║██╔═══╝ ╚════██║
// ██████╔╝██║██║ ╚████║╚██████╔╝██║     ███████║
// ╚═════╝ ╚═╝╚═╝  ╚═══╝ ╚═════╝ ╚═╝     ╚══════╝
// `

 	m.applicationName = `
▀█████████▄   ▄█  ███▄▄▄▄    ▄██████▄     ▄███████▄    ▄████████ 
  ███    ███ ███  ███▀▀▀██▄ ███    ███   ███    ███   ███    ███ 
  ███    ███ ███▌ ███   ███ ███    ███   ███    ███   ███    █▀  
 ▄███▄▄▄██▀  ███▌ ███   ███ ███    ███   ███    ███   ███        
▀▀███▀▀▀██▄  ███▌ ███   ███ ███    ███ ▀█████████▀  ▀███████████ 
  ███    ██▄ ███  ███   ███ ███    ███   ███                 ███ 
  ███    ███ ███  ███   ███ ███    ███   ███           ▄█    ███ 
▄█████████▀  █▀    ▀█   █▀   ▀██████▀   ▄████▀       ▄████████▀  
                                                                 	
`

// 	m.applicationName = `
// `

	m.styles.title = lipgloss.NewStyle().Bold(true).Foreground(theme.PanelTitle)
	m.styles.stats = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(theme.Border).
		BorderTop(false)
	m.styles.dependencies = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(theme.Border).
		BorderTop(false)
	m.styles.label = lipgloss.NewStyle().Foreground(theme.Label)
	m.styles.description = lipgloss.NewStyle().Foreground(theme.Description)
	m.styles.creator = lipgloss.NewStyle().Italic(true).Foreground(theme.Creator)
	m.styles.depHeader = lipgloss.NewStyle().Bold(true).Foreground(theme.Label)
	m.styles.depSelected = lipgloss.NewStyle().Foreground(theme.SelectionFG).Background(theme.SelectionBG)
	// m.propertiesViewport.SetContent(m.propertiesView())
	// m.dependenciesViewport.SetContent(m.dependenciesView())

	return m
}

func (m GeneralModel) Init() tea.Cmd {
	return nil
}

func (m GeneralModel) Update(msg tea.Msg) (GeneralModel, tea.Cmd) {
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

func (m GeneralModel) View() string {

	applicationName := lipgloss.NewStyle().Foreground(m.theme.Body).Render(m.applicationName)
	description := m.styles.description.Render(m.description)
	repository := lipgloss.NewStyle().Hyperlink(m.repository).Render(m.repository)
	creator := m.styles.creator.Render(m.creator)
	properties := m.statsView()
	dependencies := m.dependenciesView()

	v := lipgloss.JoinVertical(lipgloss.Center, applicationName, description, repository, creator, properties, dependencies)
	v = lipgloss.PlaceHorizontal(m.width, lipgloss.Center, v)
	return v
}

func (m *GeneralModel) setDimensions(width, height int) {
	m.width = width
	m.height = height
	// m.styles.stats.Width(min(40, width-2)).Height(min(15, height-2))
	// m.styles.dependencies.Width(min(60, width-2)).Height(min(7, height-2))
}

func (m GeneralModel) statsView() string {
	var contents strings.Builder
	valStyle := lipgloss.NewStyle().Foreground(m.theme.Body)
	write := func(label, value string) {
		contents.WriteString(m.styles.label.Render(label) + valStyle.Render(value))
		contents.WriteByte('\n')
	}
	optionalInt64 := func(value int64, ok bool, format string) string {
		if !ok {
			return "N/A"
		}
		return lipgloss.Sprintf(format, value)
	}
	optionalUint64 := func(value uint64, ok bool, format string) string {
		if !ok {
			return "N/A"
		}
		return lipgloss.Sprintf(format, value)
	}
	idWithName := func(value uint32, ok bool, name string) string {
		if !ok {
			return "N/A"
		}
		if name != "" {
			return lipgloss.Sprintf("%d/%s", value, name)
		}
		return lipgloss.Sprintf("%d", value)
	}
	optionalTime := func(t time.Time, ok bool) string {
		if !ok {
			return "N/A"
		}
		// Match common stat-style layout: date, time with nanoseconds, numeric zone (e.g. +0000).
		return t.Format("2006-01-02 15:04:05.000000000 -0700")
	}
	write("File Size: ", humanFileSize(m.fileStats.FileSize))
	write("Blocks: ", optionalInt64(m.fileStats.Blocks, m.fileStats.HasBlocks, "%d"))
	write("Block Size: ", optionalInt64(m.fileStats.BlockSize, m.fileStats.HasBlockSize, "%d bytes"))
	write("Links: ", optionalUint64(m.fileStats.Links, m.fileStats.HasLinks, "%d"))
	write("UID: ", idWithName(m.fileStats.UID, m.fileStats.HasUID, m.fileStats.UIDName))
	write("GID: ", idWithName(m.fileStats.GID, m.fileStats.HasGID, m.fileStats.GIDName))
	write("Access: ", optionalTime(m.fileStats.LastAccessAt, m.fileStats.HasLastAccessTime))
	write("Modify: ", optionalTime(m.fileStats.LastModAt, m.fileStats.HasLastModTime))

	fullContent := m.styles.stats.Render(contents.String())
	title := m.styles.title.Render("┌|" + m.fileStats.Filename + "|")
	line := strings.Repeat("─", max(0, lipgloss.Width(fullContent)-lipgloss.Width(title)-1))
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
				return m.styles.depHeader
			}
			if row == m.dependenciesSelected {
				return m.styles.depSelected
			}
			return lipgloss.NewStyle()
		})

	if len(m.dependencies) == 0 {
		t.Row("No dependencies reported", "not found")
	} else {
		for _, dependency := range m.dependencies {
			path := dependency.LibraryPath
			if !dependency.Found && path == "" {
				path = "not found"
			}
			t.Row(dependency.LibraryName, path)
		}
	}

	fullContent := m.styles.dependencies.Render(t.Render())
	title := m.styles.title.Render("┌|Dependencies|")
	line := strings.Repeat("─", max(0, lipgloss.Width(fullContent)-lipgloss.Width(title)-1))
	line = lipgloss.JoinHorizontal(lipgloss.Center, title, line, "┐")
	return lipgloss.JoinVertical(lipgloss.Left, line, fullContent)
}

func humanFileSize(bytes int64) string {
	if bytes < 0 {
		return fmt.Sprintf("%d B", bytes)
	}
	const (
		KiB = 1024
		MiB = KiB * 1024
		GiB = MiB * 1024
		TiB = GiB * 1024
	)
	b := float64(bytes)
	switch {
	case bytes < KiB:
		return fmt.Sprintf("%d B", bytes)
	case bytes < MiB:
		return fmt.Sprintf("%.1f KB", b/KiB)
	case bytes < GiB:
		return fmt.Sprintf("%.1f MB", b/MiB)
	case bytes < TiB:
		return fmt.Sprintf("%.2f GB", b/GiB)
	default:
		return fmt.Sprintf("%.2f TB", b/TiB)
	}
}

func (m GeneralModel) dependenciesVisibleRows() int {
	// Reserve a reasonable slice of the page for dependencies while allowing scrolling.
	return max(4, min(12, m.height/3))
}

func (m GeneralModel) dependenciesDataRowsVisible() int {
	// The table header consumes one visible line.
	return max(1, m.dependenciesVisibleRows()-1)
}
