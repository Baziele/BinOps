package main

import (
	"fmt"
	"strings"

	"debug/elf"

	goansi "github.com/charmbracelet/x/ansi"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/mattn/go-runewidth"
)

type StaticModel struct {
	width int
	height int
	theme Theme
	styles struct{
		title lipgloss.Style
		header lipgloss.Style
		fileSegment lipgloss.Style
		fileSegmentTable table.Styles
	}
	header string
	fileSegments []string
	content string
	currentSegment int
	elfFile         *elf.File
	elfHeader       ELFStaticHeader
	elfNotes        []ELFNoteSection
	segmentTables   []ELFStaticTable
	fileSegmentTable table.Model
	detailOpen       bool
	detailContent    string
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



func initializeStaticModel(width, height int, elfFile *elf.File, header ELFStaticHeader, notes []ELFNoteSection, segmentTables []ELFStaticTable, theme Theme) StaticModel {
	if segmentTables == nil {
		segmentTables = ParseELFSegmentTables(elfFile)
	}
	m := StaticModel{
		width: width,
		height: height,
		theme: theme,
		fileSegments: []string{"Program Header", "Section Header", "Symbols", "Dynamic Symbol", "Dynamic", "Relocations"},
		content: "Static",
		elfFile: elfFile,
		elfHeader: header,
		elfNotes: notes,
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
		Width(m.width/2).
		Height(max(1, m.topPanelHeight()-1)).
		Margin(0)
	m.styles.fileSegment = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(theme.Border).
		BorderTop(false).
		Width(m.width).
		Height(max(1, m.bottomPanelHeight()-1)).
		Margin(0)
	m.styles.fileSegmentTable = themedStaticTableStyles(theme)
	m.fileSegmentTable.SetStyles(m.styles.fileSegmentTable)
	(&m).refreshFileSegmentTable()
	return m
}

func themedStaticTableStyles(theme Theme) table.Styles {
	return table.Styles{
		Header: lipgloss.NewStyle().
			Bold(true).
			Foreground(theme.Label).
			Padding(0, 1),
		Cell: lipgloss.NewStyle().
			Foreground(theme.Body).
			Padding(0, 1),
		Selected: lipgloss.NewStyle().
			Bold(true).
			Foreground(theme.SelectionFG).
			Background(theme.SelectionBG),
	}
}


func (m StaticModel) Init() tea.Cmd{
	return nil
}

func (m StaticModel) Update(msg tea.Msg) (StaticModel, tea.Cmd) {
	if m.detailOpen {
		if kp, ok := msg.(tea.KeyPressMsg); ok {
			switch kp.String() {
			case "esc", "q", "enter":
				m.detailOpen = false
				return m, nil
			}
		}
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, DefaultKeyMap.Left):
			if len(m.fileSegments) == 0 {
				return m, nil
			}
			n := len(m.fileSegments)
			m.currentSegment = (m.currentSegment - 1 + n) % n
			m.refreshFileSegmentTable()
			return m, nil
		case key.Matches(msg, DefaultKeyMap.Right):
			if len(m.fileSegments) == 0 {
				return m, nil
			}
			n := len(m.fileSegments)
			m.currentSegment = (m.currentSegment + 1) % n
			m.refreshFileSegmentTable()
			return m, nil
		case msg.String() == "enter":
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
		return base
	}
	panel := m.renderDetailPanel()
	return overlayCenter(base, panel, m.width, m.height)
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
	if m.height <= 1 {
		return 1
	}
	return max(1, m.height/2)
}

func (m StaticModel) bottomPanelHeight() int {
	return max(1, m.height-m.topPanelHeight())
}

func (m StaticModel) headerView() string{
	title := m.styles.title.Render("┌┤Headers├")
	line := strings.Repeat("─", max(0, (m.width/2)-lipgloss.Width(title) - 1))
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
	title := m.styles.title.Render("┌┤Notes├")
	line := strings.Repeat("─", max(0, (m.width/2)-lipgloss.Width(title)-1))
	line = lipgloss.JoinHorizontal(lipgloss.Center, title, line, "┐")
	innerH := max(1, m.topPanelHeight()-lipgloss.Height(line))

	displayStyle := lipgloss.NewStyle().Foreground(m.theme.PanelTitle)
	sectionStyle := lipgloss.NewStyle().Foreground(m.theme.NavAccent).Bold(true)
	tableHeaderStyle := lipgloss.NewStyle().Foreground(m.theme.Muted)
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
	tableBox := lipgloss.NewStyle().
		Width(innerW).
		MaxWidth(innerW).
		Height(max(1, innerH-lipgloss.Height(segTitle))).
		MaxHeight(max(1, innerH-lipgloss.Height(segTitle))).
		Render(m.fileSegmentTable.View())
	contents := m.styles.fileSegment.Height(panelContentH).Render(lipgloss.JoinVertical(lipgloss.Left, segTitle, tableBox))

	return lipgloss.JoinVertical(lipgloss.Left, line, contents)
}


func (m *StaticModel) setDimensions(width, height int){
	m.width = width
	m.height = height
	m.styles.header = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(m.theme.Border).
		BorderTop(false).
		Width(m.width/2).
		Height(max(1, m.topPanelHeight()-1))
	m.styles.fileSegment = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(m.theme.Border).
		BorderTop(false).
		Width(m.width).
		Height(max(1, m.bottomPanelHeight()-1)).
		Margin(0)
	m.styles.fileSegmentTable = themedStaticTableStyles(m.theme)
	m.fileSegmentTable.SetStyles(m.styles.fileSegmentTable)
	m.refreshFileSegmentTable()
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

	setFallback := func(msg string) {
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
	cols := tableColumnsFromHeaders(seg.Headers, rows, tableWidth, m.currentSegment)
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

func tableColumnsFromHeaders(headers []string, rows []table.Row, totalWidth int, segmentIdx int) []table.Column {
	tw := totalWidth - 2
	if tw < 8 {
		tw = 8
	}
	n := len(headers)
	if n == 0 {
		return nil
	}

	// Use actual content widths — no proportional inflation.
	widths := make([]int, n)
	for j, h := range headers {
		title := shortStaticTableHeader(h)
		w := max(runewidth.StringWidth(title), 3)
		for _, r := range rows {
			if j < len(r) {
				cw := runewidth.StringWidth(r[j])
				if cw > w {
					w = cw
				}
			}
		}
		if cap := staticColumnMaxWidth(segmentIdx, h); cap > 0 && w > cap {
			w = cap
		}
		widths[j] = w
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
		if widths[best] <= 3 {
			break
		}
		widths[best]--
	}

	// Give any leftover space only to the last column so no gaps appear mid-table.
	total := 0
	for _, w := range widths {
		total += w
	}
	if total < tw && n > 0 {
		widths[n-1] += tw - total
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