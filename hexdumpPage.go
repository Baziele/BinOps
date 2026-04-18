package main

import (
	"encoding/binary"
	"fmt"
	"image/color"
	"math"
	"os"
	"strconv"
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type HexDocument struct {
	base  []byte
	dirty map[int]byte
}

func newHexDocument(fileBytes []byte) *HexDocument {
	return &HexDocument{
		base:  fileBytes,
		dirty: make(map[int]byte),
	}
}

func (d *HexDocument) Len() int {
	if d == nil {
		return 0
	}
	return len(d.base)
}

func (d *HexDocument) ByteAt(i int) (byte, bool) {
	if d == nil || i < 0 || i >= len(d.base) {
		return 0, false
	}
	if b, ok := d.dirty[i]; ok {
		return b, true
	}
	return d.base[i], true
}

func (d *HexDocument) HasDirtyAt(i int) bool {
	if d == nil {
		return false
	}
	_, ok := d.dirty[i]
	return ok
}

func (d *HexDocument) SetByteAt(i int, b byte) bool {
	if d == nil || i < 0 || i >= len(d.base) {
		return false
	}
	if d.base[i] == b {
		delete(d.dirty, i)
		return true
	}
	d.dirty[i] = b
	return true
}

func (d *HexDocument) ReadAt(offset, n int) ([]byte, bool) {
	if d == nil || n <= 0 || offset < 0 || offset+n > len(d.base) {
		return nil, false
	}
	buf := make([]byte, n)
	for i := 0; i < n; i++ {
		buf[i], _ = d.ByteAt(offset + i)
	}
	return buf, true
}

func (d *HexDocument) DirtyCount() int {
	if d == nil {
		return 0
	}
	return len(d.dirty)
}

func (d *HexDocument) Materialize() []byte {
	if d == nil {
		return nil
	}
	out := make([]byte, len(d.base))
	copy(out, d.base)
	for offset, value := range d.dirty {
		out[offset] = value
	}
	return out
}

func (d *HexDocument) Save(path string) error {
	if d == nil {
		return nil
	}
	mode := os.FileMode(0o644)
	if info, err := os.Stat(path); err == nil {
		mode = info.Mode().Perm()
	}
	data := d.Materialize()
	if err := os.WriteFile(path, data, mode); err != nil {
		return err
	}
	d.base = data
	clear(d.dirty)
	return nil
}

type HexViewState struct {
	cursorOffset int
	topRow       int
	bytesPerRow  int
	visibleRows  int
	byteOrder    binary.ByteOrder
	editMode     bool
	highNibble   bool
}

type renderedHexRow struct {
	address string
	hex     string
	ascii   string
}

type hexCellStyles struct {
	normal         lipgloss.Style
	selected       lipgloss.Style
	edited         lipgloss.Style
	selectedEdited lipgloss.Style
}

type hexdumpPalette struct {
	null      hexCellStyles
	space     hexCellStyles
	high      hexCellStyles
	control   hexCellStyles
	printable hexCellStyles
	standard  hexCellStyles
}

type HexdumpModel struct {
	width      int
	height     int
	theme      Theme
	binaryPath string
	document   *HexDocument
	view       HexViewState
	rowCache   map[int]renderedHexRow
	layout     struct {
		addressWidth int
		hexWidth     int
		asciiWidth   int
	}
	styles struct {
		title           lipgloss.Style
		address         lipgloss.Style
		hex             lipgloss.Style
		ascii           lipgloss.Style
		converter       lipgloss.Style
		selectedAddress lipgloss.Style
	}
	palette       hexdumpPalette
	notifications string
}

type HexdumpKeyMap struct {
	Left         key.Binding
	Right        key.Binding
	Up           key.Binding
	Down         key.Binding
	PageUp       key.Binding
	PageDown     key.Binding
	RowStart     key.Binding
	RowEnd       key.Binding
	ToggleEndian key.Binding
	ToggleEdit   key.Binding
	Save         key.Binding
}

var HexdumpDefaultKeyMap = HexdumpKeyMap{
	Left:         key.NewBinding(key.WithKeys("h", "left"), key.WithHelp("←/h", "move left")),
	Right:        key.NewBinding(key.WithKeys("l", "right"), key.WithHelp("→/l", "move right")),
	Up:           key.NewBinding(key.WithKeys("k", "up"), key.WithHelp("↑/k", "move up")),
	Down:         key.NewBinding(key.WithKeys("j", "down"), key.WithHelp("↓/j", "move down")),
	PageUp:       key.NewBinding(key.WithKeys("pgup"), key.WithHelp("pgup", "page up")),
	PageDown:     key.NewBinding(key.WithKeys("pgdown"), key.WithHelp("pgdn", "page down")),
	RowStart:     key.NewBinding(key.WithKeys("home", "^"), key.WithHelp("home", "row start")),
	RowEnd:       key.NewBinding(key.WithKeys("end", "$"), key.WithHelp("end", "row end")),
	ToggleEndian: key.NewBinding(key.WithKeys("n"), key.WithHelp("n", "toggle endian")),
	ToggleEdit:   key.NewBinding(key.WithKeys("i"), key.WithHelp("i", "edit mode")),
	Save:         key.NewBinding(key.WithKeys("ctrl+s"), key.WithHelp("ctrl+s", "save")),
}

const converterShortRead = "EOF"

func hexdumpByteForeground(theme Theme, b byte) color.Color {
	switch {
	case b == 0x00:
		return theme.Hexdump.Null
	case b == 0x09 || b == 0x20:
		return theme.Hexdump.Space
	case b >= 0x80:
		return theme.Hexdump.High
	case (b >= 0x01 && b <= 0x1F) || b == 0x7F:
		return theme.Hexdump.Control
	case b >= 0x21 && b <= 0x7E:
		return theme.Hexdump.Printable
	default:
		return theme.Hexdump.Standard
	}
}

func newHexCellStyles(theme Theme, fg color.Color) hexCellStyles {
	base := lipgloss.NewStyle().Foreground(fg)
	return hexCellStyles{
		normal:         base,
		selected:       base.Background(theme.Hexdump.SelectedBG).Foreground(theme.Hexdump.SelectedFG),
		edited:         base.Underline(true).Bold(true),
		selectedEdited: base.Background(theme.Hexdump.SelectedBG).Foreground(theme.Hexdump.SelectedFG).Underline(true).Bold(true),
	}
}

func newHexdumpPalette(theme Theme) hexdumpPalette {
	return hexdumpPalette{
		null:      newHexCellStyles(theme, hexdumpByteForeground(theme, 0x00)),
		space:     newHexCellStyles(theme, hexdumpByteForeground(theme, 0x20)),
		high:      newHexCellStyles(theme, hexdumpByteForeground(theme, 0x80)),
		control:   newHexCellStyles(theme, hexdumpByteForeground(theme, 0x01)),
		printable: newHexCellStyles(theme, hexdumpByteForeground(theme, 0x21)),
		standard:  newHexCellStyles(theme, theme.Hexdump.Standard),
	}
}

func initializeHexdumpModel(width, height int, binaryPath string, fileBytes []byte, theme Theme) HexdumpModel {
	m := HexdumpModel{
		width:         width,
		height:        height,
		theme:         theme,
		binaryPath:    binaryPath,
		document:      newHexDocument(fileBytes),
		rowCache:      make(map[int]renderedHexRow),
		palette:       newHexdumpPalette(theme),
		notifications: "Read-only mode. Press i to edit, ctrl+s to save.",
	}
	m.view.byteOrder = binary.LittleEndian
	m.view.highNibble = true
	m.styles.title = lipgloss.NewStyle().Bold(true).Foreground(theme.PanelTitle)
	m.styles.address = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(theme.Border).BorderTop(false)
	m.styles.hex = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(theme.Border).BorderTop(false)
	m.styles.ascii = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(theme.Border).BorderTop(false)
	m.styles.selectedAddress = lipgloss.NewStyle().Background(theme.Hexdump.SelectedBG).Foreground(theme.Hexdump.SelectedFG).Bold(true)
	m.setDimensions(width, height)
	return m
}

func (m HexdumpModel) Init() tea.Cmd {
	return nil
}

func (m HexdumpModel) Update(msg tea.Msg) (HexdumpModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, HexdumpDefaultKeyMap.Up):
			m.moveCursor(-m.view.bytesPerRow)
		case key.Matches(msg, HexdumpDefaultKeyMap.Down):
			m.moveCursor(m.view.bytesPerRow)
		case key.Matches(msg, HexdumpDefaultKeyMap.Left):
			m.moveCursor(-1)
		case key.Matches(msg, HexdumpDefaultKeyMap.Right):
			m.moveCursor(1)
		case key.Matches(msg, HexdumpDefaultKeyMap.PageUp):
			m.moveCursor(-(m.view.visibleRows * m.view.bytesPerRow))
		case key.Matches(msg, HexdumpDefaultKeyMap.PageDown):
			m.moveCursor(m.view.visibleRows * m.view.bytesPerRow)
		case key.Matches(msg, HexdumpDefaultKeyMap.RowStart):
			m.moveToRowBoundary(false)
		case key.Matches(msg, HexdumpDefaultKeyMap.RowEnd):
			m.moveToRowBoundary(true)
		case key.Matches(msg, HexdumpDefaultKeyMap.ToggleEndian):
			if m.view.byteOrder == binary.LittleEndian {
				m.view.byteOrder = binary.BigEndian
				m.notifications = "Big-endian converter view."
			} else {
				m.view.byteOrder = binary.LittleEndian
				m.notifications = "Little-endian converter view."
			}
		case key.Matches(msg, HexdumpDefaultKeyMap.ToggleEdit):
			m.view.editMode = !m.view.editMode
			m.view.highNibble = true
			if m.view.editMode {
				m.notifications = "Edit mode enabled. Type hex digits to modify bytes."
			} else {
				m.notifications = "Read-only mode enabled."
			}
		case key.Matches(msg, HexdumpDefaultKeyMap.Save):
			if m.document.DirtyCount() == 0 {
				m.notifications = "No edits to save."
				return m, nil
			}
			if m.binaryPath == "" {
				m.notifications = "Save failed: missing target path."
				return m, nil
			}
			if err := m.document.Save(m.binaryPath); err != nil {
				m.notifications = fmt.Sprintf("Save failed: %v", err)
			} else {
				m.notifications = "Saved changes to disk."
				m.invalidateAllRows()
				return m, analyzeBinary(m.binaryPath)
			}
		case msg.String() == "esc":
			if m.view.editMode {
				m.view.editMode = false
				m.view.highNibble = true
				m.notifications = "Edit mode cancelled."
			}
		default:
			if m.view.editMode {
				if value, ok := parseHexKey(msg.String()); ok {
					m.applyHexDigit(value)
				}
			}
		}
	}
	return m, nil
}

func (m *HexdumpModel) setDimensions(width, height int) {
	m.width = width
	m.height = height
	m.layout.addressWidth = 8
	m.view.bytesPerRow = max(1, (width-13)/4)
	m.layout.hexWidth = max(2, m.view.bytesPerRow*3-1)
	m.layout.asciiWidth = max(1, m.view.bytesPerRow)

	converterWidth := max(18, width/4)
	m.styles.converter = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(m.theme.Border).
		BorderTop(false).
		Foreground(m.theme.Body).
		Width(converterWidth).
		Height(2)

	converterRowsHeight := lipgloss.Height(m.lowerConverterView())
	m.view.visibleRows = max(1, height-converterRowsHeight-1)
	m.ensureCursorInBounds()
	m.ensureVisible()
	m.invalidateAllRows()
}

func (m *HexdumpModel) selectionStart() int {
	if m.document.Len() == 0 {
		return 0
	}
	return m.view.cursorOffset
}

func (m *HexdumpModel) readAt(n int) ([]byte, bool) {
	return m.document.ReadAt(m.selectionStart(), n)
}

func (m *HexdumpModel) totalRows() int {
	if m.document.Len() == 0 || m.view.bytesPerRow == 0 {
		return 0
	}
	return (m.document.Len()-1)/m.view.bytesPerRow + 1
}

func (m *HexdumpModel) ensureCursorInBounds() {
	if m.document.Len() == 0 {
		m.view.cursorOffset = 0
		m.view.topRow = 0
		return
	}
	m.view.cursorOffset = clamp(m.view.cursorOffset, 0, m.document.Len()-1)
}

func (m *HexdumpModel) ensureVisible() {
	row := 0
	if m.view.bytesPerRow > 0 {
		row = m.view.cursorOffset / m.view.bytesPerRow
	}
	maxTop := max(0, m.totalRows()-m.view.visibleRows)
	switch {
	case row < m.view.topRow:
		m.view.topRow = row
	case row >= m.view.topRow+m.view.visibleRows:
		m.view.topRow = row - m.view.visibleRows + 1
	}
	m.view.topRow = clamp(m.view.topRow, 0, maxTop)
}

func (m *HexdumpModel) moveCursor(delta int) {
	if m.document.Len() == 0 {
		return
	}
	oldOffset := m.view.cursorOffset
	m.view.cursorOffset = clamp(m.view.cursorOffset+delta, 0, m.document.Len()-1)
	m.view.highNibble = true
	m.invalidateOffset(oldOffset)
	m.invalidateOffset(m.view.cursorOffset)
	m.ensureVisible()
}

func (m *HexdumpModel) moveToRowBoundary(end bool) {
	if m.document.Len() == 0 {
		return
	}
	oldOffset := m.view.cursorOffset
	rowStart := (m.view.cursorOffset / m.view.bytesPerRow) * m.view.bytesPerRow
	if end {
		rowEnd := min(rowStart+m.view.bytesPerRow-1, m.document.Len()-1)
		m.view.cursorOffset = rowEnd
	} else {
		m.view.cursorOffset = rowStart
	}
	m.view.highNibble = true
	m.invalidateOffset(oldOffset)
	m.invalidateOffset(m.view.cursorOffset)
	m.ensureVisible()
}

func (m *HexdumpModel) applyHexDigit(value byte) {
	if m.document.Len() == 0 {
		return
	}
	current, ok := m.document.ByteAt(m.view.cursorOffset)
	if !ok {
		return
	}
	if m.view.highNibble {
		current = (value << 4) | (current & 0x0F)
		m.view.highNibble = false
	} else {
		current = (current & 0xF0) | value
		m.view.highNibble = true
	}
	m.document.SetByteAt(m.view.cursorOffset, current)
	m.invalidateOffset(m.view.cursorOffset)
	m.notifications = fmt.Sprintf("Edited offset 0x%X. Unsaved edits: %d", m.view.cursorOffset, m.document.DirtyCount())
	if m.view.highNibble && m.view.cursorOffset < m.document.Len()-1 {
		oldOffset := m.view.cursorOffset
		m.view.cursorOffset++
		m.invalidateOffset(oldOffset)
		m.invalidateOffset(m.view.cursorOffset)
		m.ensureVisible()
	}
}

func (m *HexdumpModel) invalidateOffset(offset int) {
	if m.view.bytesPerRow <= 0 {
		return
	}
	delete(m.rowCache, offset/m.view.bytesPerRow)
}

func (m *HexdumpModel) invalidateAllRows() {
	clear(m.rowCache)
}

func (m HexdumpModel) renderVisibleRows() (string, string, string) {
	var addrBuilder, hexBuilder, asciiBuilder strings.Builder
	for i := 0; i < m.view.visibleRows; i++ {
		row := m.view.topRow + i
		var rendered renderedHexRow
		if row < m.totalRows() {
			rendered = m.cachedRow(row)
		} else {
			rendered = renderedHexRow{
				address: strings.Repeat(" ", m.layout.addressWidth),
				hex:     strings.Repeat(" ", m.layout.hexWidth),
				ascii:   strings.Repeat(" ", m.layout.asciiWidth),
			}
		}
		addrBuilder.WriteString(rendered.address)
		hexBuilder.WriteString(rendered.hex)
		asciiBuilder.WriteString(rendered.ascii)
		if i < m.view.visibleRows-1 {
			addrBuilder.WriteByte('\n')
			hexBuilder.WriteByte('\n')
			asciiBuilder.WriteByte('\n')
		}
	}
	return addrBuilder.String(), hexBuilder.String(), asciiBuilder.String()
}

func (m HexdumpModel) cachedRow(row int) renderedHexRow {
	if cached, ok := m.rowCache[row]; ok {
		return cached
	}
	rendered := m.renderRow(row)
	m.rowCache[row] = rendered
	return rendered
}

func (m HexdumpModel) renderRow(row int) renderedHexRow {
	rowStart := row * m.view.bytesPerRow
	cursorRow := 0
	if m.view.bytesPerRow > 0 {
		cursorRow = m.view.cursorOffset / m.view.bytesPerRow
	}

	address := fmt.Sprintf("%08X", rowStart)
	if row == cursorRow && m.document.Len() > 0 {
		address = m.styles.selectedAddress.Render(address)
	}

	var hexBuilder, asciiBuilder strings.Builder
	for i := 0; i < m.view.bytesPerRow; i++ {
		offset := rowStart + i
		if i > 0 {
			hexBuilder.WriteByte(' ')
		}
		if offset >= m.document.Len() {
			hexBuilder.WriteString("  ")
			asciiBuilder.WriteByte(' ')
			continue
		}

		value, _ := m.document.ByteAt(offset)
		selected := offset == m.view.cursorOffset
		edited := m.document.HasDirtyAt(offset)
		style := m.styleForByte(value, selected, edited)
		hexBuilder.WriteString(style.Render(fmt.Sprintf("%02X", value)))
		asciiBuilder.WriteString(style.Render(renderASCIIByte(value)))
	}

	return renderedHexRow{
		address: address,
		hex:     padLine(hexBuilder.String(), m.layout.hexWidth),
		ascii:   padLine(asciiBuilder.String(), m.layout.asciiWidth),
	}
}

func (m HexdumpModel) styleForByte(b byte, selected bool, edited bool) lipgloss.Style {
	var variants hexCellStyles
	switch {
	case b == 0x00:
		variants = m.palette.null
	case b == 0x09 || b == 0x20:
		variants = m.palette.space
	case b >= 0x80:
		variants = m.palette.high
	case (b >= 0x01 && b <= 0x1F) || b == 0x7F:
		variants = m.palette.control
	case b >= 0x21 && b <= 0x7E:
		variants = m.palette.printable
	default:
		variants = m.palette.standard
	}

	switch {
	case selected && edited:
		return variants.selectedEdited
	case selected:
		return variants.selected
	case edited:
		return variants.edited
	default:
		return variants.normal
	}
}

func renderASCIIByte(b byte) string {
	if b >= 0x20 && b <= 0x7E {
		return string(b)
	}
	return "."
}

func padLine(value string, width int) string {
	lineWidth := lipgloss.Width(value)
	if lineWidth >= width {
		return value
	}
	return value + strings.Repeat(" ", width-lineWidth)
}

func (m HexdumpModel) View() string {
	addressText, hexText, asciiText := m.renderVisibleRows()
	upperView := lipgloss.JoinHorizontal(
		lipgloss.Left,
		m.renderPanel("Address", m.styles.address.Height(m.view.visibleRows).Render(addressText)),
		m.renderPanel("Hex", m.styles.hex.Height(m.view.visibleRows).Render(hexText)),
		m.renderPanel("ASCII", m.styles.ascii.Height(m.view.visibleRows).Render(asciiText)),
	)
	lowerView := m.lowerConverterView()
	return lipgloss.JoinVertical(lipgloss.Left, upperView, lowerView)
}

func (m HexdumpModel) renderPanel(titleText string, body string) string {
	title := m.styles.title.Render("┌" + titleText)
	lineWidth := max(0, lipgloss.Width(body)-lipgloss.Width(title)-1)
	line := strings.Repeat("─", lineWidth)
	header := lipgloss.JoinHorizontal(lipgloss.Center, title, line, "┐")
	return lipgloss.JoinVertical(lipgloss.Left, header, body)
}

func (m HexdumpModel) lowerConverterView() string {
	row1 := lipgloss.JoinHorizontal(lipgloss.Left, m.convertSigned8Bit(), m.convertSigned32Bit(), m.convertHexadecimal(), m.convertFloat32Bit())
	row2 := lipgloss.JoinHorizontal(lipgloss.Left, m.convertUnsigned8Bit(), m.convertUnsigned32Bit(), m.convertOctal(), m.convertFloat64Bit())
	row3 := lipgloss.JoinHorizontal(lipgloss.Left, m.convertSigned16Bit(), m.convertSigned64Bit(), m.convertBinary(), m.convertOffset())
	row4 := lipgloss.JoinHorizontal(lipgloss.Left, m.convertUnsigned16Bit(), m.convertUnsigned64Bit(), m.convertStreamLength(), m.convertNotification())
	return lipgloss.JoinVertical(lipgloss.Left, row1, row2, row3, row4)
}

func (m HexdumpModel) renderConverter(titleText, value string) string {
	title := m.styles.title.Render("┌" + titleText)
	line := strings.Repeat("─", max(0, m.styles.converter.GetWidth()-lipgloss.Width(title)-1))
	header := lipgloss.JoinHorizontal(lipgloss.Center, title, line, "┐")
	return lipgloss.JoinVertical(lipgloss.Left, header, m.styles.converter.Render(value))
}

func (m HexdumpModel) convertSigned8Bit() string {
	buf, ok := m.readAt(1)
	if !ok {
		return m.renderConverter("Signed 8 Bit", converterShortRead)
	}
	return m.renderConverter("Signed 8 Bit", strconv.FormatInt(int64(int8(buf[0])), 10))
}

func (m HexdumpModel) convertSigned32Bit() string {
	buf, ok := m.readAt(4)
	if !ok {
		return m.renderConverter("Signed 32 Bit", converterShortRead)
	}
	v := int32(m.view.byteOrder.Uint32(buf))
	return m.renderConverter("Signed 32 Bit", strconv.FormatInt(int64(v), 10))
}

func (m HexdumpModel) convertHexadecimal() string {
	buf, ok := m.readAt(4)
	if !ok {
		return m.renderConverter("Hexadecimal", converterShortRead)
	}
	return m.renderConverter("Hexadecimal", fmt.Sprintf("%08X", m.view.byteOrder.Uint32(buf)))
}

func (m HexdumpModel) convertFloat32Bit() string {
	buf, ok := m.readAt(4)
	if !ok {
		return m.renderConverter("Float 32 bit", converterShortRead)
	}
	u := m.view.byteOrder.Uint32(buf)
	return m.renderConverter("Float 32 bit", strconv.FormatFloat(float64(math.Float32frombits(u)), 'e', -1, 32))
}

func (m HexdumpModel) convertUnsigned8Bit() string {
	buf, ok := m.readAt(1)
	if !ok {
		return m.renderConverter("Unsigned 8 Bit", converterShortRead)
	}
	return m.renderConverter("Unsigned 8 Bit", strconv.FormatUint(uint64(buf[0]), 10))
}

func (m HexdumpModel) convertUnsigned32Bit() string {
	buf, ok := m.readAt(4)
	if !ok {
		return m.renderConverter("Unsigned 32 Bit", converterShortRead)
	}
	u := m.view.byteOrder.Uint32(buf)
	return m.renderConverter("Unsigned 32 Bit", strconv.FormatUint(uint64(u), 10))
}

func (m HexdumpModel) convertOctal() string {
	buf, ok := m.readAt(1)
	if !ok {
		return m.renderConverter("Octal", converterShortRead)
	}
	return m.renderConverter("Octal", strconv.FormatUint(uint64(buf[0]), 8))
}

func (m HexdumpModel) convertFloat64Bit() string {
	buf, ok := m.readAt(8)
	if !ok {
		return m.renderConverter("Float 64 bit", converterShortRead)
	}
	u := m.view.byteOrder.Uint64(buf)
	return m.renderConverter("Float 64 bit", strconv.FormatFloat(math.Float64frombits(u), 'e', -1, 64))
}

func (m HexdumpModel) convertSigned16Bit() string {
	buf, ok := m.readAt(2)
	if !ok {
		return m.renderConverter("Signed 16 Bit", converterShortRead)
	}
	v := int16(m.view.byteOrder.Uint16(buf))
	return m.renderConverter("Signed 16 Bit", strconv.FormatInt(int64(v), 10))
}

func (m HexdumpModel) convertSigned64Bit() string {
	buf, ok := m.readAt(8)
	if !ok {
		return m.renderConverter("Signed 64 Bit", converterShortRead)
	}
	v := int64(m.view.byteOrder.Uint64(buf))
	return m.renderConverter("Signed 64 Bit", strconv.FormatInt(v, 10))
}

func (m HexdumpModel) convertBinary() string {
	buf, ok := m.readAt(1)
	if !ok {
		return m.renderConverter("Binary", converterShortRead)
	}
	return m.renderConverter("Binary", fmt.Sprintf("%08b", buf[0]))
}

func (m HexdumpModel) convertOffset() string {
	return m.renderConverter("Offset", strconv.FormatInt(int64(m.selectionStart()), 16))
}

func (m HexdumpModel) convertUnsigned16Bit() string {
	buf, ok := m.readAt(2)
	if !ok {
		return m.renderConverter("Unsigned 16 Bit", converterShortRead)
	}
	u := m.view.byteOrder.Uint16(buf)
	return m.renderConverter("Unsigned 16 Bit", strconv.FormatUint(uint64(u), 10))
}

func (m HexdumpModel) convertUnsigned64Bit() string {
	buf, ok := m.readAt(8)
	if !ok {
		return m.renderConverter("Unsigned 64 Bit", converterShortRead)
	}
	u := m.view.byteOrder.Uint64(buf)
	return m.renderConverter("Unsigned 64 Bit", strconv.FormatUint(u, 10))
}

func (m HexdumpModel) convertStreamLength() string {
	return m.renderConverter("Stream Length", strconv.FormatInt(int64(m.document.Len()), 10))
}

func (m HexdumpModel) convertNotification() string {
	mode := "view"
	if m.view.editMode {
		nibble := "high"
		if !m.view.highNibble {
			nibble = "low"
		}
		mode = "edit/" + nibble
	}
	value := fmt.Sprintf("%s | dirty:%d | %s", mode, m.document.DirtyCount(), m.notifications)
	return m.renderConverter("Notifications", value)
}

func parseHexKey(input string) (byte, bool) {
	if len(input) != 1 {
		return 0, false
	}
	switch {
	case input[0] >= '0' && input[0] <= '9':
		return input[0] - '0', true
	case input[0] >= 'a' && input[0] <= 'f':
		return input[0] - 'a' + 10, true
	case input[0] >= 'A' && input[0] <= 'F':
		return input[0] - 'A' + 10, true
	default:
		return 0, false
	}
}

func clamp(v, low, high int) int {
	if high < low {
		return low
	}
	if v < low {
		return low
	}
	if v > high {
		return high
	}
	return v
}
