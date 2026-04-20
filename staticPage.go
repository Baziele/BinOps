package main

import (
	"fmt"
	"strings"

	"debug/elf"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	goansi "github.com/charmbracelet/x/ansi"
	"github.com/mattn/go-runewidth"
)

type StaticModel struct {
	width  int
	height int
	theme  Theme
	styles struct {
		title            lipgloss.Style
		header           lipgloss.Style
		fileSegment      lipgloss.Style
		fileSegmentTable table.Styles
	}
	header           string
	fileSegments     []string
	content          string
	currentSegment   int
	elfFile          *elf.File
	elfHeader        ELFStaticHeader
	elfNotes         []ELFNoteSection
	segmentTables    []ELFStaticTable
	fileSegmentTable table.Model
	detailOpen       bool
	detailContent    string
}

type StaticKeyMap struct {
	Left        key.Binding
	Right       key.Binding
	OpenDetail  key.Binding
	CloseDetail key.Binding
}

var StaticDefaultKeyMap = StaticKeyMap{
	Left: key.NewBinding(
		key.WithKeys("h", "left"),        // actual keybindings
		key.WithHelp("←/h", "move left"), // corresponding help text
	),
	Right: key.NewBinding(
		key.WithKeys("l", "right"),
		key.WithHelp("→/l", "move right"),
	),
	OpenDetail: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "open row")),
	CloseDetail: key.NewBinding(
		key.WithKeys("esc", "enter"),
		key.WithHelp("esc", "close detail")),
}

func (km StaticKeyMap) ShortHelp() []key.Binding {
	if km.CloseDetail.Enabled() {
		return []key.Binding{shortHelpBinding("esc/enter", "close detail")}
	}
	var items []key.Binding
	if km.Left.Enabled() || km.Right.Enabled() {
		items = append(items, shortHelpBinding("h/l", "section"))
	}
	items = append(items, shortHelpBinding("j/k", "scroll"))
	if km.OpenDetail.Enabled() {
		items = append(items, shortHelpBinding("enter", "row"))
	}
	return items
}

func (km StaticKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{km.ShortHelp()}
}

func initializeStaticModel(width, height int, elfFile *elf.File, header ELFStaticHeader, notes []ELFNoteSection, segmentTables []ELFStaticTable, theme Theme) StaticModel {
	if segmentTables == nil {
		segmentTables = ParseELFSegmentTables(elfFile)
	}
	m := StaticModel{
		width:         width,
		height:        height,
		theme:         theme,
		fileSegments:  []string{"Program Header", "Section Header", "Symbols", "Dynamic Symbol", "Dynamic", "Relocations"},
		content:       "Static",
		elfFile:       elfFile,
		elfHeader:     header,
		elfNotes:      notes,
		segmentTables: segmentTables,
		fileSegmentTable: table.New(
			table.WithColumns([]table.Column{{Title: " ", Width: max(6, width-4)}}),
			table.WithRows(nil),
			table.WithFocused(true),
			table.WithWidth(width),
			table.WithHeight(max(8, height/2-lipgloss.Height(" ")-4)),
		),
	}
	m.styles.title = lipgloss.NewStyle().Bold(true).Foreground(theme.PanelTitle)
	m.styles.header = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(theme.Border).
		BorderTop(false).
		Width(m.width / 2).
		Height(max(1, m.topPanelHeight()-1)).
		Margin(0)
	m.styles.fileSegment = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(theme.Border).
		BorderTop(false).
		Width(m.width).
		Height(max(1, m.bottomPanelHeight()-1)).
		Margin(0)
	m.styles.fileSegmentTable = themedStaticTableStyles(theme, 1)
	m.fileSegmentTable.SetStyles(m.styles.fileSegmentTable)
	(&m).refreshFileSegmentTable()
	return m
}

func themedStaticTableStyles(theme Theme, cellPadding int) table.Styles {
	return table.Styles{
		Header: lipgloss.NewStyle().
			Bold(true).
			Foreground(theme.Label).
			Padding(0, cellPadding),
		Cell: lipgloss.NewStyle().
			Padding(0, cellPadding),
		Selected: lipgloss.NewStyle().
			Bold(true).
			Foreground(theme.SelectionFG).
			Background(theme.SelectionBG),
	}
}

func (m StaticModel) Init() tea.Cmd {
	return nil
}

func (m StaticModel) Update(msg tea.Msg) (StaticModel, tea.Cmd) {
	if m.detailOpen {
		if kp, ok := msg.(tea.KeyPressMsg); ok {
			switch {
			case key.Matches(kp, StaticDefaultKeyMap.CloseDetail):
				m.detailOpen = false
				return m, nil
			}
		}
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, StaticDefaultKeyMap.Left):
			if len(m.fileSegments) == 0 {
				return m, nil
			}
			n := len(m.fileSegments)
			m.currentSegment = (m.currentSegment - 1 + n) % n
			m.refreshFileSegmentTable()
			return m, nil
		case key.Matches(msg, StaticDefaultKeyMap.Right):
			if len(m.fileSegments) == 0 {
				return m, nil
			}
			n := len(m.fileSegments)
			m.currentSegment = (m.currentSegment + 1) % n
			m.refreshFileSegmentTable()
			return m, nil
		case key.Matches(msg, StaticDefaultKeyMap.OpenDetail):
			if len(m.segmentTables) == 0 || m.currentSegment < 0 || m.currentSegment >= len(m.segmentTables) {
				break
			}
			seg := m.segmentTables[m.currentSegment]
			if len(seg.Rows) == 0 {
				break
			}
			cursor := m.fileSegmentTable.Cursor()
			if cursor < 0 || cursor >= len(seg.Rows) {
				break
			}
			m.detailContent = buildStaticRowDetail(seg, cursor, m.currentSegment, m.theme)
			m.detailOpen = true
			return m, nil
		}
	}
	var tcmd tea.Cmd
	m.fileSegmentTable, tcmd = m.fileSegmentTable.Update(msg)
	return m, tcmd
}

func (m StaticModel) View() string {
	base := lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.JoinHorizontal(lipgloss.Center, m.headerView(), m.notesView()),
		m.fileSegmentsView())
	if !m.detailOpen {
		return lipgloss.JoinVertical(lipgloss.Left, base, m.helpView())
	}
	panel := m.renderDetailPanel()
	return lipgloss.JoinVertical(lipgloss.Left, overlayCenter(base, panel, m.width, m.contentHeight()), m.helpView())
}

func (m StaticModel) renderDetailPanel() string {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(m.theme.PanelTitle)
	hintStyle := lipgloss.NewStyle().Foreground(m.theme.Muted)
	sepStyle := lipgloss.NewStyle().Foreground(m.theme.Subtle)

	innerW := max(24, min(52, m.width-12))
	sepLine := sepStyle.Render(strings.Repeat("─", innerW))

	hint := hintStyle.Render("enter/esc  close")
	title := titleStyle.Render("Row Detail")

	contentLines := strings.Split(m.detailContent, "\n")
	maxBody := max(4, m.height/3)
	if len(contentLines) > maxBody {
		contentLines = contentLines[:maxBody]
	}
	body := strings.Join(contentLines, "\n")

	inner := lipgloss.JoinVertical(lipgloss.Left, title, sepLine, body, sepLine, hint)
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.theme.Border).
		Padding(0, 1).
		Width(innerW + 2).
		Render(inner)
}

func (m StaticModel) topPanelHeight() int {
	contentHeight := m.contentHeight()
	if contentHeight <= 1 {
		return 1
	}
	return max(1, contentHeight/2)
}

func (m StaticModel) bottomPanelHeight() int {
	return max(1, m.contentHeight()-m.topPanelHeight())
}

func (m StaticModel) headerView() string {
	title := m.styles.title.Render("┌|Headers|")
	line := strings.Repeat("─", max(0, (m.width/2)-lipgloss.Width(title)-1))
	line = lipgloss.JoinHorizontal(lipgloss.Center, title, line, "┐")
	labelStyle := lipgloss.NewStyle().Foreground(m.theme.Label)
	panelContentH := max(1, m.topPanelHeight()-lipgloss.Height(line))
	if m.elfFile == nil {
		return m.styles.header.Height(panelContentH).Render(lipgloss.JoinVertical(lipgloss.Left, line, "No ELF file loaded."))
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
	header := m.styles.header.Height(panelContentH).Render(strings.TrimSuffix(b.String(), "\n"))
	return lipgloss.JoinVertical(lipgloss.Left, line, header)
}

func (m StaticModel) notesView() string {
	title := m.styles.title.Render("┌|Notes|")
	line := strings.Repeat("─", max(0, (m.width/2)-lipgloss.Width(title)-1))
	line = lipgloss.JoinHorizontal(lipgloss.Center, title, line, "┐")
	innerH := max(1, m.topPanelHeight()-lipgloss.Height(line))
	innerW := max(16, m.styles.header.GetWidth()-m.styles.header.GetHorizontalFrameSize())

	displayStyle := lipgloss.NewStyle().Foreground(m.theme.PanelTitle).Bold(true)
	sectionStyle := lipgloss.NewStyle().Foreground(m.theme.NavAccent).Bold(true)
	labelStyle := lipgloss.NewStyle().Foreground(m.theme.Label)
	ownerStyle := lipgloss.NewStyle().Foreground(m.theme.Body)
	sizeStyle := lipgloss.NewStyle().Foreground(m.theme.Warning)
	descStyle := lipgloss.NewStyle().Foreground(m.theme.Body)
	detailKeyStyle := lipgloss.NewStyle().Foreground(m.theme.Label)
	detailMutedStyle := lipgloss.NewStyle().Foreground(m.theme.MutedStrong)
	if m.elfFile == nil {
		body := m.styles.header.Height(innerH).Render("No ELF file loaded.")
		return lipgloss.JoinVertical(lipgloss.Left, line, body)
	}
	if len(m.elfNotes) == 0 {
		body := m.styles.header.Height(innerH).Render("No note sections in this ELF file.")
		return lipgloss.JoinVertical(lipgloss.Left, line, body)
	}

	var b strings.Builder
	for secIdx, sec := range m.elfNotes {
		if secIdx > 0 {
			b.WriteByte('\n')
			b.WriteByte('\n')
		}
		b.WriteString(displayStyle.Render("Section"))
		b.WriteString(" ")
		b.WriteString(sectionStyle.Render(sec.SectionName))
		b.WriteByte('\n')
		for entryIdx, e := range sec.Entries {
			if entryIdx > 0 {
				b.WriteByte('\n')
			}
			sizeStr := fmt.Sprintf("0x%08x", e.DataSize)
			b.WriteString(renderWrappedNoteField("Owner", e.Owner, innerW, labelStyle, ownerStyle))
			b.WriteByte('\n')
			b.WriteString(renderWrappedNoteField("Size", sizeStr, innerW, labelStyle, sizeStyle))
			b.WriteByte('\n')
			b.WriteString(renderWrappedNoteField("Description", e.Description, innerW, labelStyle, descStyle))
			for _, d := range e.Details {
				b.WriteByte('\n')
				writeNoteDetailLine(&b, d, innerW, detailKeyStyle, detailMutedStyle, ownerStyle)
			}
		}
	}
	notes := m.styles.header.Height(innerH).Render(strings.TrimSuffix(b.String(), "\n"))
	return lipgloss.JoinVertical(lipgloss.Left, line, notes)
}

// writeNoteDetailLine colors leading indentation and a "Key:" prefix like readelf output.
func writeNoteDetailLine(b *strings.Builder, line string, width int, keyStyle, mutedStyle, bodyStyle lipgloss.Style) {
	if line == "" {
		return
	}
	trimStart := 0
	for trimStart < len(line) && line[trimStart] == ' ' {
		trimStart++
	}
	rest := line[trimStart:]
	indent := min(trimStart, max(0, width-6))
	prefix := strings.Repeat(" ", indent)
	if strings.HasPrefix(rest, "Properties:") {
		b.WriteString(renderWrappedStyledBlock(prefix, rest, width, mutedStyle))
		return
	}
	if idx := strings.Index(rest, ": "); idx >= 0 {
		key := rest[:idx+1]
		val := rest[idx+2:]
		labelPrefix := prefix + key
		b.WriteString(renderWrappedStyledValue(labelPrefix, val, width, keyStyle, bodyStyle))
		return
	}
	b.WriteString(renderWrappedStyledBlock(prefix, rest, width, bodyStyle))
}

func renderWrappedNoteField(label, value string, width int, labelStyle, valueStyle lipgloss.Style) string {
	return renderWrappedStyledValue(label+": ", value, width, labelStyle, valueStyle)
}

func renderWrappedStyledValue(prefix, value string, width int, prefixStyle, valueStyle lipgloss.Style) string {
	usableWidth := max(8, width)
	prefixWidth := runewidth.StringWidth(prefix)
	if prefixWidth >= usableWidth-2 {
		prefixWidth = 0
		prefix = ""
	}
	valueWidth := max(8, usableWidth-prefixWidth)
	wrapped := wrapVisualText(strings.TrimSpace(value), valueWidth)
	lines := strings.Split(wrapped, "\n")
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

func renderWrappedStyledBlock(prefix, value string, width int, style lipgloss.Style) string {
	usableWidth := max(8, width)
	prefixWidth := runewidth.StringWidth(prefix)
	textWidth := max(8, usableWidth-prefixWidth)
	wrapped := wrapVisualText(strings.TrimSpace(value), textWidth)
	lines := strings.Split(wrapped, "\n")

	var out strings.Builder
	for i, line := range lines {
		if i > 0 {
			out.WriteByte('\n')
		}
		out.WriteString(prefix)
		out.WriteString(style.Render(line))
	}
	return out.String()
}

func wrapVisualText(s string, width int) string {
	s = strings.TrimSpace(strings.ReplaceAll(s, "\t", " "))
	if s == "" {
		return ""
	}
	width = max(8, width)

	words := strings.Fields(s)
	if len(words) == 0 {
		return ""
	}

	var lines []string
	current := ""
	for _, piece := range splitVisualWord(words[0], width) {
		if current == "" {
			current = piece
			continue
		}
		lines = append(lines, current)
		current = piece
	}
	for _, word := range words[1:] {
		wordParts := splitVisualWord(word, width)
		candidate := current + " " + wordParts[0]
		if runewidth.StringWidth(candidate) <= width {
			current = candidate
		} else {
			if current != "" {
				lines = append(lines, current)
			}
			current = wordParts[0]
		}
		for _, part := range wordParts[1:] {
			if current != "" {
				lines = append(lines, current)
			}
			current = part
		}
	}
	if current != "" {
		lines = append(lines, current)
	}
	return strings.Join(lines, "\n")
}

func splitVisualWord(word string, width int) []string {
	if word == "" {
		return []string{""}
	}
	if runewidth.StringWidth(word) <= width {
		return []string{word}
	}

	var parts []string
	var b strings.Builder
	currentWidth := 0
	for _, r := range word {
		rw := runewidth.RuneWidth(r)
		if rw <= 0 {
			rw = 1
		}
		if currentWidth+rw > width && b.Len() > 0 {
			parts = append(parts, b.String())
			b.Reset()
			currentWidth = 0
		}
		b.WriteRune(r)
		currentWidth += rw
	}
	if b.Len() > 0 {
		parts = append(parts, b.String())
	}
	return parts
}

func (m StaticModel) fileSegmentTabLine() string {
	str := "┌"
	for i, page := range m.fileSegments {
		isFirst, isCurrentPage := i == 0, i == m.currentSegment
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
	line := strings.Repeat("─", max(0, m.width-lipgloss.Width(str)-1))
	return lipgloss.JoinHorizontal(lipgloss.Center, str, line, "┐")
}

func (m StaticModel) fileSegmentPanelInnerHeight() int {
	line := m.fileSegmentTabLine()
	panelContentH := max(1, m.bottomPanelHeight()-lipgloss.Height(line))
	return max(1, panelContentH-m.styles.fileSegment.GetVerticalFrameSize())
}

func (m StaticModel) fileSegmentPanelInnerWidth() int {
	return max(1, m.width-m.styles.fileSegment.GetHorizontalFrameSize())
}

func (m StaticModel) fileSegmentsView() string {
	line := m.fileSegmentTabLine()
	panelContentH := max(1, m.bottomPanelHeight()-lipgloss.Height(line))
	innerH := m.fileSegmentPanelInnerHeight()
	innerW := m.fileSegmentPanelInnerWidth()
	segIdx := 0
	if len(m.fileSegments) > 0 {
		segIdx = (m.currentSegment%len(m.fileSegments) + len(m.fileSegments)) % len(m.fileSegments)
	}
	segTitle := m.styles.title.Render(clipVisual(m.fileSegments[segIdx], max(8, innerW-2)))
	tableView := lipgloss.NewStyle().
		Foreground(m.theme.Body).
		Render(m.fileSegmentTable.View())
	tableBox := lipgloss.NewStyle().
		Width(innerW).
		MaxWidth(innerW).
		Height(max(1, innerH-lipgloss.Height(segTitle))).
		MaxHeight(max(1, innerH-lipgloss.Height(segTitle))).
		Render(tableView)
	contents := m.styles.fileSegment.Height(panelContentH).Render(lipgloss.JoinVertical(lipgloss.Left, segTitle, tableBox))

	return lipgloss.JoinVertical(lipgloss.Left, line, contents)
}

func (m *StaticModel) setDimensions(width, height int) {
	m.width = width
	m.height = height
	m.styles.title = lipgloss.NewStyle().Bold(true).Foreground(m.theme.PanelTitle)
	m.styles.header = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(m.theme.Border).
		BorderTop(false).
		Width(m.width / 2).
		Height(max(1, m.topPanelHeight()-1))
	m.styles.fileSegment = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(m.theme.Border).
		BorderTop(false).
		Width(m.width).
		Height(max(1, m.bottomPanelHeight()-1)).
		Margin(0)
	m.styles.fileSegmentTable = themedStaticTableStyles(m.theme, 1)
	m.fileSegmentTable.SetStyles(m.styles.fileSegmentTable)
	m.refreshFileSegmentTable()
}

func (m StaticModel) helpKeyMap() StaticKeyMap {
	km := StaticDefaultKeyMap
	km.CloseDetail.SetEnabled(m.detailOpen)
	km.Left.SetEnabled(!m.detailOpen && len(m.fileSegments) > 0)
	km.Right.SetEnabled(!m.detailOpen && len(m.fileSegments) > 0)
	km.OpenDetail.SetEnabled(!m.detailOpen)
	return km
}

func (m StaticModel) helpView() string {
	return renderHelpFooter(m.theme, m.width, m.helpKeyMap(), DefaultAppKeyMap)
}

func (m StaticModel) contentHeight() int {
	return max(1, m.height-helpFooterHeight(m.theme, m.width, m.helpKeyMap(), DefaultAppKeyMap))
}

func (m *StaticModel) refreshFileSegmentTable() {
	if m == nil {
		return
	}

	nSeg := len(m.fileSegments)
	if nSeg == 0 {
		m.currentSegment = 0
	} else {
		m.currentSegment = ((m.currentSegment % nSeg) + nSeg) % nSeg
	}

	innerH := m.fileSegmentPanelInnerHeight()
	innerW := m.fileSegmentPanelInnerWidth()
	segTitleH := 0
	if nSeg > 0 && m.currentSegment >= 0 && m.currentSegment < nSeg {
		title := clipVisual(m.fileSegments[m.currentSegment], max(8, innerW-2))
		segTitleH = lipgloss.Height(m.styles.title.Render(title))
	}

	// Bubbles table subtracts header height from this value; if it is too small,
	// the viewport height goes negative and row rendering can panic. Keep a floor.
	avail := max(1, innerH-segTitleH)
	tableOuter := max(3, avail)
	tableWidth := m.fileSegmentPanelInnerWidth()
	tablePadding := staticTableCellPadding(tableWidth, 1)

	setFallback := func(msg string) {
		m.styles.fileSegmentTable = themedStaticTableStyles(m.theme, tablePadding)
		m.fileSegmentTable.SetStyles(m.styles.fileSegmentTable)
		m.fileSegmentTable.SetWidth(tableWidth)
		// Reconfigure safely: columns updates render previous rows immediately.
		// Put a zero-cell row first so render cannot index past columns.
		m.fileSegmentTable.SetRows([]table.Row{{}})
		m.fileSegmentTable.SetColumns([]table.Column{{Title: " ", Width: max(6, tableWidth-2)}})
		m.fileSegmentTable.SetRows([]table.Row{{msg}})
		m.fileSegmentTable.SetCursor(0)
		m.fileSegmentTable.SetHeight(tableOuter)
	}

	if len(m.segmentTables) == 0 || m.currentSegment < 0 || m.currentSegment >= len(m.segmentTables) {
		setFallback("(no table data)")
		return
	}
	seg := m.segmentTables[m.currentSegment]
	if len(seg.Headers) == 0 {
		setFallback("(no data)")
		return
	}

	rows := tableRowsNormalized(seg.Rows, len(seg.Headers))
	tablePadding = staticTableCellPadding(tableWidth, len(seg.Headers))
	m.styles.fileSegmentTable = themedStaticTableStyles(m.theme, tablePadding)
	m.fileSegmentTable.SetStyles(m.styles.fileSegmentTable)
	cols := tableColumnsFromHeaders(seg.Headers, rows, tableWidth, m.currentSegment, tablePadding)
	if len(rows) == 0 {
		// Empty body: bubbles leaves cursor at -1 for 0 rows; SetCursor(0) does not fix it.
		placeholder := make(table.Row, len(seg.Headers))
		for i := range placeholder {
			placeholder[i] = "—"
		}
		rows = []table.Row{placeholder}
	}

	m.fileSegmentTable.SetWidth(tableWidth)
	// Reconfigure safely when switching to a table with fewer columns.
	m.fileSegmentTable.SetRows([]table.Row{{}})
	m.fileSegmentTable.SetColumns(cols)
	m.fileSegmentTable.SetRows(rows)
	m.fileSegmentTable.SetCursor(0)
	m.fileSegmentTable.SetHeight(tableOuter)
}

func shortStaticTableHeader(h string) string {
	switch h {
	case "Flags":
		return "Flg"
	case "EntSize":
		return "EntSz"
	case "SymIdx":
		return "Sym#"
	case "FileSz":
		return "FSz"
	case "MemSz":
		return "MSz"
	default:
		return h
	}
}

func tableColumnsFromHeaders(headers []string, rows []table.Row, totalWidth int, segmentIdx int, cellPadding int) []table.Column {
	n := len(headers)
	if n == 0 {
		return nil
	}

	tw := totalWidth - (n * cellPadding * 2)
	if tw < n {
		tw = n
	}

	minColWidth := 3
	if tw < n*minColWidth {
		minColWidth = max(1, tw/n)
	}

	// Keep both the natural width (true content size) and a softer preferred
	// width that keeps dense tables compact until there is room to breathe.
	widths := make([]int, n)
	naturalWidths := make([]int, n)
	for j, h := range headers {
		title := shortStaticTableHeader(h)
		natural := max(runewidth.StringWidth(title), minColWidth)
		for _, r := range rows {
			if j < len(r) {
				cw := runewidth.StringWidth(r[j])
				if cw > natural {
					natural = cw
				}
			}
		}
		naturalWidths[j] = natural
		widths[j] = natural
		if cap := staticColumnMaxWidth(segmentIdx, h); cap > 0 && widths[j] > cap {
			widths[j] = cap
		}
	}

	// If total natural width exceeds tw, trim the widest columns first.
	for {
		total := 0
		for _, w := range widths {
			total += w
		}
		if total <= tw {
			break
		}
		best := 0
		for k := 1; k < n; k++ {
			if widths[k] > widths[best] {
				best = k
			}
		}
		if widths[best] <= minColWidth {
			break
		}
		widths[best]--
	}

	// First, expand capped columns back toward their natural size when space
	// allows, so wide layouts don't leave most of the table cramped.
	total := 0
	for _, w := range widths {
		total += w
	}
	if total < tw && n > 0 {
		extra := tw - total
		for extra > 0 {
			progressed := false
			for j := range widths {
				if widths[j] >= naturalWidths[j] {
					continue
				}
				widths[j]++
				extra--
				progressed = true
				if extra == 0 {
					break
				}
			}
			if !progressed {
				break
			}
		}
		if extra > 0 {
			for j := 0; extra > 0; j = (j + 1) % n {
				widths[j]++
				extra--
			}
		}
	}

	cols := make([]table.Column, n)
	for j := range headers {
		title := shortStaticTableHeader(headers[j])
		if title == "" {
			title = " "
		}
		cols[j] = table.Column{Title: title, Width: widths[j]}
	}
	return cols
}

func staticTableCellPadding(totalWidth, columnCount int) int {
	if columnCount <= 0 {
		return 1
	}

	// Prefer some spacing, but switch to a compact layout before padding causes
	// dense tables to overflow narrow terminals.
	if totalWidth-(columnCount*2) < columnCount*3 {
		return 0
	}
	return 1
}

func staticColumnMaxWidth(segmentIdx int, header string) int {
	switch segmentIdx {
	case 2, 3: // Symbols / Dynamic Symbols
		switch header {
		case "Name":
			return 18
		case "Value":
			return 14
		case "Size":
			return 8
		case "Type":
			return 7
		case "Bind":
			return 7
		case "Vis":
			return 10
		case "Section":
			return 10
		case "Version":
			return 12
		case "Library":
			return 14
		default:
			return 10
		}
	case 5: // Relocations
		switch header {
		case "Section":
			return 18
		case "Kind":
			return 4
		case "Offset":
			return 10
		case "SymIdx":
			return 4
		case "Symbol":
			return 14
		case "Type":
			return 18
		case "Addend":
			return 10
		default:
			return 8
		}
	default:
		switch header {
		case "Name":
			return 18
		case "Type":
			return 16
		case "Flags":
			return 8
		case "Addr", "Off", "Size", "Offset", "VAddr", "PAddr", "FileSz", "MemSz", "Align", "EntSize":
			return 10
		default:
			return 14
		}
	}
}

// overlayCenter draws panel as a transparent overlay centered over bg.
// It uses ANSI-aware cuts so existing terminal colours in bg are preserved
// in the regions around the panel.
func overlayCenter(bg, panel string, bgW, bgH int) string {
	panelW := lipgloss.Width(panel)
	panelH := lipgloss.Height(panel)
	startX := max(0, (bgW-panelW)/2)
	startY := max(0, (bgH-panelH)/2)

	bgLines := strings.Split(bg, "\n")
	for len(bgLines) < bgH {
		bgLines = append(bgLines, strings.Repeat(" ", bgW))
	}

	panelLines := strings.Split(panel, "\n")
	for i, pLine := range panelLines {
		row := startY + i
		if row >= len(bgLines) {
			break
		}
		bgLine := bgLines[row]
		left := goansi.Cut(bgLine, 0, startX)
		right := goansi.TruncateLeft(bgLine, startX+panelW, "")
		bgLines[row] = left + pLine + right
	}
	return strings.Join(bgLines, "\n")
}

func staticSegmentLongTitle(tab int) string {
	names := []string{
		"Program headers",
		"Section headers",
		"Symbol table (.symtab)",
		"Dynamic symbol table (.dynsym)",
		"Dynamic section (.dynamic)",
		"Relocations (SHT_REL / SHT_RELA)",
	}
	if tab >= 0 && tab < len(names) {
		return names[tab]
	}
	return ""
}

func buildStaticRowDetail(seg ELFStaticTable, row, segmentTab int, theme Theme) string {
	if row < 0 || row >= len(seg.Rows) {
		return ""
	}
	headers := seg.Headers
	cells := seg.Rows[row]
	var details []string
	if seg.CellDetails != nil && row < len(seg.CellDetails) {
		details = seg.CellDetails[row]
	}

	keyStyle := lipgloss.NewStyle().Foreground(theme.Label).Bold(true)
	metaStyle := lipgloss.NewStyle().Foreground(theme.MutedStrong)
	detailStyle := lipgloss.NewStyle().Foreground(theme.Subtle).Italic(true)

	maxKeyW := 0
	for _, h := range headers {
		if len(h) > maxKeyW {
			maxKeyW = len(h)
		}
	}

	var b strings.Builder
	b.WriteString(metaStyle.Render(fmt.Sprintf("%s  row %d", staticSegmentLongTitle(segmentTab), row+1)))
	b.WriteByte('\n')

	for j := range headers {
		if j >= len(cells) {
			break
		}
		key := fmt.Sprintf("%-*s", maxKeyW, headers[j])
		b.WriteString(keyStyle.Render(key))
		b.WriteString("  ")
		b.WriteString(cells[j])
		if details != nil && j < len(details) && details[j] != "" {
			b.WriteString("  ")
			b.WriteString(detailStyle.Render(details[j]))
		}
		b.WriteByte('\n')
	}
	return strings.TrimSuffix(b.String(), "\n")
}

func clipVisual(s string, maxW int) string {
	if maxW <= 1 {
		if maxW == 1 {
			return "…"
		}
		return ""
	}
	if runewidth.StringWidth(s) <= maxW {
		return s
	}
	var b strings.Builder
	w := 0
	for _, r := range s {
		rw := runewidth.RuneWidth(r)
		if w+rw > maxW-1 {
			break
		}
		b.WriteRune(r)
		w += rw
	}
	b.WriteString("…")
	return b.String()
}

func tableRowsNormalized(rows [][]string, nCol int) []table.Row {
	if nCol <= 0 {
		return nil
	}
	out := make([]table.Row, len(rows))
	for i, r := range rows {
		if len(r) > nCol {
			r = r[:nCol]
		}
		row := make(table.Row, nCol)
		copy(row, table.Row(r))
		for j := len(r); j < nCol; j++ {
			row[j] = ""
		}
		out[i] = row
	}
	return out
}
