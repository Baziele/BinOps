package main

import (
	"fmt"
	"strings"

	"debug/elf"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type StaticModel struct {
	width int
	height int
	styles struct{
		title lipgloss.Style
		header lipgloss.Style
		fileSegment lipgloss.Style
	}
	header string
	fileSegments []string
	content string
	currentSegment int
	elfFile    *elf.File
	elfHeader  ELFStaticHeader
	elfNotes   []ELFNoteSection
}

type KeyMap struct {
    Left key.Binding
    Right key.Binding
}

var DefaultKeyMap = KeyMap{
    Left: key.NewBinding(
        key.WithKeys("h", "left"),        // actual keybindings
        key.WithHelp("←/h", "move left"), // corresponding help text
    ),
    Right: key.NewBinding(
        key.WithKeys("l", "right"),
        key.WithHelp("→/l", "move right"),
    ),
}



func initializeStaticModel(width, height int, elfFile *elf.File, header ELFStaticHeader, notes []ELFNoteSection) StaticModel {
	m := StaticModel{
		width: width,
		height: height,
		fileSegments: []string{"Program Header", "Section Header", "Symbols", "Dynamic Symbol", "Dynamic", "Relocations"},
		content: "Static",
		elfFile: elfFile,
		elfHeader: header,
		elfNotes: notes,
	}
	b := lipgloss.NormalBorder()
	b.Right = "├"
	b.Left = "┤"
	b.Bottom = ""
	b.Top = ""
	m.styles.title = lipgloss.NewStyle().Bold(true) //.BorderStyle(b).Padding(0, 1).Margin(0) //.BorderBottom(false)
	m.styles.header = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderTop(false).Width(m.width/2).Height(m.height/2).Margin(0)
	m.styles.fileSegment = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderTop(false).Width(m.width).Height(m.height/2).Margin(0)
	return m
}


func (m StaticModel) Init() tea.Cmd{
	return nil
}

func (m StaticModel) Update(msg tea.Msg) (StaticModel, tea.Cmd){
	switch msg := msg.(type) {
    case tea.KeyPressMsg:
        switch {
        case key.Matches(msg, DefaultKeyMap.Left):
            m.currentSegment = (m.currentSegment - 1 + len(m.fileSegments)) % len(m.fileSegments)
        case key.Matches(msg, DefaultKeyMap.Right):
            m.currentSegment = (m.currentSegment + 1) % len(m.fileSegments)
        }
    }
	return m, nil
}


func (m StaticModel) View() string{
	return lipgloss.JoinVertical(lipgloss.Left, lipgloss.JoinHorizontal(lipgloss.Center, m.headerView(), m.notesView()), m.fileSegmentsView())
}

func (m StaticModel) headerView() string{
	title := m.styles.title.Render("┌┤Headers├")
	line := strings.Repeat("─", max(0, (m.width/2)-lipgloss.Width(title) - 1))
	line = lipgloss.JoinHorizontal(lipgloss.Center, title, line, "┐")
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Cyan)
	if m.elfFile == nil {
		return m.styles.header.Height(m.height/2 - lipgloss.Height(line)).Render(lipgloss.JoinVertical(lipgloss.Left, line, "No ELF file loaded."))
	}

	h := m.elfHeader
	var b strings.Builder
	write := func(label, value string) {
		b.WriteString(labelStyle.Render(label))
		b.WriteString(value)
		b.WriteByte('\n')
	}

	write("Class: ", h.Class)
	write("Endianness: ", h.Endianness)
	write("Version: ", h.Version)
	write("OS/ABI: ", h.OSABI)
	write("ABI Version: ", h.ABIVersion)
	write("Type: ", h.Type)
	write("Arch: ", h.Arch)
	write("Entry point address: ", h.EntryAddress)
	if h.ProgHeaderStart != "" {
		write("Start of program headers: ", h.ProgHeaderStart)
		write("Start of section headers: ", h.SectionHeaderStart)
		write("Flags: ", h.Flags)
		write("Size of this header: ", h.EhsizeBytes)
		write("Size of program header: ", h.PhEntSizeBytes)
	}
	write("Number of program headers: ", h.ProgramHeaderCount)
	header := m.styles.header.Height(m.height/2 - lipgloss.Height(line)).Render(strings.TrimSuffix(b.String(), "\n"))
	return lipgloss.JoinVertical(lipgloss.Left, line, header)
}

func (m StaticModel) notesView() string {
	title := m.styles.title.Render("┌┤Notes├")
	line := strings.Repeat("─", max(0, (m.width/2)-lipgloss.Width(title)-1))
	line = lipgloss.JoinHorizontal(lipgloss.Center, title, line, "┐")
	innerH := m.height/2 - lipgloss.Height(line)

	displayStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("39"))
	sectionStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4")).Bold(true)
	tableHeaderStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	ownerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	sizeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	detailKeyStyle := lipgloss.NewStyle().Foreground(lipgloss.Cyan)
	detailMutedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245"))

	if m.elfFile == nil {
		body := m.styles.header.Height(innerH).Render("No ELF file loaded.")
		return lipgloss.JoinVertical(lipgloss.Left, line, body)
	}
	if len(m.elfNotes) == 0 {
		body := m.styles.header.Height(innerH).Render("No note sections in this ELF file.")
		return lipgloss.JoinVertical(lipgloss.Left, line, body)
	}

	var b strings.Builder
	for _, sec := range m.elfNotes {
		b.WriteString(displayStyle.Render("Displaying notes found in: "))
		b.WriteString(sectionStyle.Render(sec.SectionName))
		b.WriteByte('\n')
		b.WriteString(tableHeaderStyle.Render("  Owner                Data size        Description"))
		b.WriteByte('\n')
		for _, e := range sec.Entries {
			sizeStr := fmt.Sprintf("0x%08x", e.DataSize)
			b.WriteString("  ")
			b.WriteString(ownerStyle.Render(fmt.Sprintf("%-20s", e.Owner)))
			b.WriteString(" ")
			b.WriteString(sizeStyle.Render(fmt.Sprintf("%-16s", sizeStr)))
			b.WriteString(" ")
			b.WriteString(descStyle.Render(e.Description))
			b.WriteByte('\n')
			for _, d := range e.Details {
				writeNoteDetailLine(&b, d, detailKeyStyle, detailMutedStyle)
			}
		}
		b.WriteByte('\n')
	}
	notes := m.styles.header.Height(innerH).Render(strings.TrimSuffix(b.String(), "\n"))
	return lipgloss.JoinVertical(lipgloss.Left, line, notes)
}

// writeNoteDetailLine colors leading indentation and a "Key:" prefix like readelf output.
func writeNoteDetailLine(b *strings.Builder, line string, keyStyle, mutedStyle lipgloss.Style) {
	if line == "" {
		return
	}
	trimStart := 0
	for trimStart < len(line) && line[trimStart] == ' ' {
		trimStart++
	}
	pad := line[:trimStart]
	rest := line[trimStart:]
	b.WriteString(pad)
	if strings.HasPrefix(rest, "Properties:") {
		b.WriteString(mutedStyle.Render(rest))
		b.WriteByte('\n')
		return
	}
	if idx := strings.Index(rest, ": "); idx >= 0 {
		key := rest[:idx+1]
		val := rest[idx+2:]
		b.WriteString(keyStyle.Render(key))
		b.WriteString(val)
		b.WriteByte('\n')
		return
	}
	b.WriteString(rest)
	b.WriteByte('\n')
}

func (m StaticModel) fileSegmentsView() string{
	str := "┌"
	for i, page := range(m.fileSegments){
		isFirst, isCurrentPage := i == 0, i == m.currentSegment
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
	line := strings.Repeat("─", max(0, m.width - lipgloss.Width(str) - 1))
	line = lipgloss.JoinHorizontal(lipgloss.Center, str, line, "┐")
	fileSegmentsContent := m.fileSegments[m.currentSegment] 
	contents := m.styles.fileSegment.Height(m.height/2 - lipgloss.Height(line) -2).Render(fileSegmentsContent)

	return lipgloss.JoinVertical(lipgloss.Left, line, contents)
}


func (m *StaticModel) setDimensions(width, height int){
	m.width = width
	m.height = height
	m.styles.header = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderTop(false).Width(m.width/2).Height(m.height/2)
	m.styles.fileSegment = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderTop(false).Width(m.width).Height(m.height/2).Margin(0)

}