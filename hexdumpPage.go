package main

import (
	"encoding/binary"
	"fmt"
	"image/color"
	"math"
	"strconv"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type HexdumpModel struct {
	width int
	height int
	content string
	address string
	hex string 
	ascii string
	addressViewport viewport.Model
	hexViewport viewport.Model
	asciiViewPort viewport.Model
	styles struct {
		title lipgloss.Style
		address lipgloss.Style
		hex lipgloss.Style
		ascii lipgloss.Style
		converter lipgloss.Style
	}
	currentHexValue string
	cursorY int
	offset int
	streamLength int
	notifications string
	fileBytes []byte
	bytesPerLine int
	byteOrder   binary.ByteOrder // LittleEndian or BigEndian for multi-byte converters
}

type HexdumpKeyMap struct {
	Left         key.Binding
	Right        key.Binding
	Up           key.Binding
	Down         key.Binding
	ToggleEndian key.Binding
}

var HexdumpDefaultKeyMap = HexdumpKeyMap{
	Left:         key.NewBinding(key.WithKeys("h", "left"), key.WithHelp("←/h", "move left")),
	Right:        key.NewBinding(key.WithKeys("l", "right"), key.WithHelp("→/l", "move right")),
	Up:           key.NewBinding(key.WithKeys("k", "up"), key.WithHelp("↑/k", "move up")),
	Down:         key.NewBinding(key.WithKeys("j", "down"), key.WithHelp("↓/j", "move down")),
	ToggleEndian: key.NewBinding(key.WithKeys("n"), key.WithHelp("n", "toggle endian")),
}

const converterShortRead = "EOF"

// hexdumpByteForeground maps each byte to a semantic color (Dracula-inspired palette).
func hexdumpByteForeground(b byte) color.Color {
	switch {
	case b == 0x00:
		return lipgloss.Color("#444444") // null
	case b == 0x09 || b == 0x20:
		return lipgloss.Color("#8BE9FD") // tab, space
	case b >= 0x80:
		return lipgloss.Color("#BD93F9") // high bytes
	case (b >= 0x01 && b <= 0x1f) || b == 0x7f:
		return lipgloss.Color("#FFB86C") // control (incl. CR, LF, DEL)
	case b >= 0x21 && b <= 0x7e:
		return lipgloss.Color("#50FA7B") // printable ASCII (excl. space)
	default:
		return lipgloss.Color("#F8F8F2") // standard / off-white
	}
}

func hexdumpByteStyle(b byte, selected bool) lipgloss.Style {
	s := lipgloss.NewStyle().Foreground(hexdumpByteForeground(b))
	if selected {
		s = s.Background(lipgloss.Color("#FFFFFF"))
	}
	return s
}

// selectionStart is the file offset of the highlighted byte (matches cursorY + offset in buildHexdumpColumns).
func (m HexdumpModel) selectionStart() int {
	return m.cursorY + m.offset
}

// readAt returns n bytes from fileBytes at the current selection; ok is false if the read would pass EOF.
func (m HexdumpModel) readAt(n int) ([]byte, bool) {
	if n <= 0 {
		return nil, false
	}
	start := m.selectionStart()
	if start < 0 || start+n > len(m.fileBytes) {
		return nil, false
	}
	return m.fileBytes[start : start+n], true
}

// buildHexdumpColumns formats raw bytes into parallel address / hex / ASCII column text for the viewports.
func buildHexdumpColumns(data []byte, bytesPerLine int, offset int, cursorY int) (addressText, hexText, asciiText string) {
	if len(data) == 0 {
		return "", "", ""
	}
	addressStyle := lipgloss.NewStyle().Background(lipgloss.White)
	var ab, hb, sb strings.Builder
	for off := 0; off < len(data); off += bytesPerLine {
		end := off + bytesPerLine
		if end > len(data) {
			end = len(data)
		}
		chunk := data[off:end]
		var str string
		if off == cursorY {
			str = addressStyle.Render(fmt.Sprintf("%08x", off))
			} else {
				str = fmt.Sprintf("%08x", off)
			}
			ab.WriteString(str + "\n")
		for i, b := range chunk {
			if i > 0 {
				hb.WriteByte(' ')
			}
			sel := off+i == cursorY+offset
			str = hexdumpByteStyle(b, sel).Render(fmt.Sprintf("%02X", b))
			hb.WriteString(str)
		}
		hb.WriteByte('\n')
		for i, b := range chunk {
			var ch string
			if b >= 0x20 && b <= 0x7E {
				ch = string(b)
			} else {
				ch = "."
			}
			sel := off+i == cursorY+offset
			str = hexdumpByteStyle(b, sel).Render(ch)
			sb.WriteString(str)
		}
		sb.WriteByte('\n')
	}
	return ab.String(), hb.String(), sb.String()
}

func initializeHexdumpModel(width, height int, fileBytes []byte) HexdumpModel {
	
	m := HexdumpModel{
		width:           width,
		height:          height,
		content:         "Hexdump",
		offset:          0,
		cursorY:         0,
		streamLength:    len(fileBytes),
		notifications:   "",
		fileBytes:       fileBytes,
		currentHexValue: "32",
		byteOrder:       binary.LittleEndian,
	}

	m.styles.title = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Cyan)
	m.styles.converter = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderTop(false).Width(m.width /4).Height(2)
	m.styles.address = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderTop(false)//.Width(8 + 2).Height(m.height/2)
	// m.styles.hex = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderTop(false).Width(((m.width - m.styles.address.GetWidth()) /4) * 3).Height(m.height - (lipgloss.Height(m.convertBinary()) *4)-1)
	m.styles.hex = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderTop(false)//.Width(((m.width - m.styles.address.GetWidth() - 4) /4) * 3 + 2).Height(m.height - (lipgloss.Height(m.convertBinary()) *4)-1)
	m.styles.ascii = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderTop(false)//.Width(((m.width - m.styles.address.GetWidth() - 4) /4) * 1 + 2).Height(m.height - (lipgloss.Height(m.convertBinary()) *4)-1)

	m.addressViewport = viewport.New(viewport.WithWidth(8), viewport.WithHeight(m.height - (lipgloss.Height(m.convertBinary()) *4)-1))
	m.addressViewport.SetContent(m.address)
	m.hexViewport = viewport.New(viewport.WithWidth(((m.width - m.addressViewport.Width() - 6) /4) * 3 ), viewport.WithHeight(m.height - (lipgloss.Height(m.convertBinary()) *4)-1))
	m.hexViewport.SetContent(m.hex)
	m.asciiViewPort = viewport.New(viewport.WithWidth(((m.width - m.addressViewport.Width() - 6) /4) * 1 ), viewport.WithHeight(m.height - (lipgloss.Height(m.convertBinary()) *4)-1))
	m.asciiViewPort.SetContent(m.ascii)
	m.bytesPerLine = m.hexViewport.Width()/3
	m.refreshHexdumpColumns()
	return m 
}


func (m HexdumpModel) Init() tea.Cmd{
	return nil
}

func (m HexdumpModel) Update(msg tea.Msg) (HexdumpModel, tea.Cmd){
	var cmd tea.Cmd
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, HexdumpDefaultKeyMap.Up):
			m.cursorY = max(0, m.cursorY - m.bytesPerLine) 
			m.currentHexValue = fmt.Sprintf("%02x", m.fileBytes[m.cursorY + m.offset])
			m.repositionView()
			m.refreshHexdumpColumns()
			return m, nil
		case key.Matches(msg, HexdumpDefaultKeyMap.Down):
			m.cursorY = min(m.cursorY + m.bytesPerLine, m.streamLength - m.bytesPerLine) 
			m.currentHexValue = fmt.Sprintf("%02x", m.fileBytes[m.cursorY + m.offset])
			m.repositionView()
			m.refreshHexdumpColumns()
			return m, nil
		case key.Matches(msg, HexdumpDefaultKeyMap.Left):
			m.offset = (m.offset - 1) % m.streamLength
			m.currentHexValue = fmt.Sprintf("%02x", m.fileBytes[m.cursorY + m.offset])
			m.repositionView()
			m.refreshHexdumpColumns()
			return m, nil
		case key.Matches(msg, HexdumpDefaultKeyMap.Right):
			m.offset = (m.offset + 1) % m.streamLength
			m.currentHexValue = fmt.Sprintf("%02x", m.fileBytes[m.cursorY + m.offset])
			m.repositionView()
			m.refreshHexdumpColumns()
			return m, nil
		case key.Matches(msg, HexdumpDefaultKeyMap.ToggleEndian):
			if m.byteOrder == binary.LittleEndian {
				m.byteOrder = binary.BigEndian
				m.notifications = "Big-endian"
			} else {
				m.byteOrder = binary.LittleEndian
				m.notifications = "Little-endian"
			}
			return m, nil
		case msg.String() == "enter":
			m.currentHexValue = fmt.Sprintf("%02x", m.fileBytes[m.cursorY + m.offset])
			// m.refreshHexdumpColumns()
			return m, nil
		}
	}
	m.addressViewport, cmd = m.addressViewport.Update(msg)
	cmds = append(cmds, cmd)
	m.hexViewport, cmd = m.hexViewport.Update(msg)
	cmds = append(cmds, cmd)
	m.asciiViewPort, cmd = m.asciiViewPort.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m *HexdumpModel) repositionView() {  
    minimum := m.addressViewport.YOffset()  
    maximum := minimum + m.addressViewport.Height() - 1  
  
    if row := m.cursorY / m.bytesPerLine; row < minimum {  
        m.addressViewport.ScrollUp(minimum - row)  
		m.hexViewport.ScrollUp(minimum - row)
		m.asciiViewPort.ScrollUp(minimum - row)
    } else if row > maximum {  
        m.addressViewport.ScrollDown(row - maximum)  
		m.hexViewport.ScrollDown(row - maximum)
		m.asciiViewPort.ScrollDown(row - maximum)
    }  
}


func (m HexdumpModel) View() string{
	upperView := lipgloss.JoinHorizontal(lipgloss.Left, m.addressView(), m.hexView(), m.asciiView())
	lowerView := m.lowerConverterView()
	return lipgloss.JoinVertical(lipgloss.Left, upperView, lowerView)
}


func (m HexdumpModel) addressView() string {
	title := m.styles.title.Render("┌Address")
	line := strings.Repeat("─",  m.addressViewport.Width() + lipgloss.Width(m.styles.address.Render("")) -lipgloss.Width(title)-1)
	line = lipgloss.JoinHorizontal(lipgloss.Center, title, line, "┐")
	panelContent := m.styles.address.Render(m.addressViewport.View())

	return lipgloss.JoinVertical(lipgloss.Left, line, panelContent)
}

func (m HexdumpModel) hexView() string {
	title := m.styles.title.Render("┌Hex")
	line := strings.Repeat("─",  m.hexViewport.Width() + lipgloss.Width(m.styles.hex.Render("")) -lipgloss.Width(title)-1)
	line = lipgloss.JoinHorizontal(lipgloss.Center, title, line, "┐")
	panelContent := m.styles.hex.Render(m.hexViewport.View())

	return lipgloss.JoinVertical(lipgloss.Left, line, panelContent)
}

func (m HexdumpModel) asciiView() string {
	title := m.styles.title.Render("┌ASCII")
	line := strings.Repeat("─",  m.asciiViewPort.Width() + lipgloss.Width(m.styles.ascii.Render("")) -lipgloss.Width(title)-1)
	line = lipgloss.JoinHorizontal(lipgloss.Center, title, line, "┐")
	panelContent := m.styles.ascii.Render(m.asciiViewPort.View())

	return lipgloss.JoinVertical(lipgloss.Left, line, panelContent)
}
	
func (m *HexdumpModel) setDimensions(width, height int) {
	m.width = width
	m.height = height
	m.styles.converter = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderTop(false).Width(m.width /4).Height(2)
	// m.styles.title = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Cyan)
	// m.styles.address = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderTop(false)//.Width(8 + 2).Height(m.height - (lipgloss.Height(m.convertBinary()) *4)-1)
	// m.styles.hex = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderTop(false)//.Width(((m.width - m.addressViewport.Width() - 8) /4) * 3 + 2).Height(m.height - (lipgloss.Height(m.convertBinary()) *4)-1)
	// m.styles.ascii = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderTop(false)//.Width(((m.width - m.addressViewport.Width() - 8) /4) * 1 + 2).Height(m.height - (lipgloss.Height(m.convertBinary()) *4)-1)
	
	m.addressViewport.SetWidth(8)
	m.hexViewport.SetWidth(((m.width - m.addressViewport.Width() - 6) /4) * 3 )
	m.asciiViewPort.SetWidth(((m.width - m.addressViewport.Width() - 6) /4) * 1 )
	m.addressViewport.SetHeight(m.height - (lipgloss.Height(m.convertBinary()) *4)-1)
	m.hexViewport.SetHeight(m.height - (lipgloss.Height(m.convertBinary()) *4)-1)
	m.asciiViewPort.SetHeight(m.height - (lipgloss.Height(m.convertBinary()) *4)-1)
	m.bytesPerLine = m.hexViewport.Width()/3
	m.refreshHexdumpColumns()
}

func (m* HexdumpModel) refreshHexdumpColumns() {
	addr, hexCol, asciiCol := buildHexdumpColumns(m.fileBytes, m.bytesPerLine, m.offset, m.cursorY)
	m.addressViewport.SetContent(addr)
	m.hexViewport.SetContent(hexCol)
	m.asciiViewPort.SetContent(asciiCol)
}

//Converter = sigend 8 bit, signed 32 bit, hexadecimal, float 32 bit, 
//unsigned 8 bit, unsigned 32 bit, octal, float 62 bit,
//signed 16 bit, signed 64 bit, binary, offset,
//unsigned 16 bit, unsigned 64 bit, stream length, notification

func (m HexdumpModel) lowerConverterView() string {
	row1 := lipgloss.JoinHorizontal(lipgloss.Left, m.convertSigned8Bit(), m.convertSigned32Bit(), m.convertHexadecimal(), m.convertFloat32Bit())
	row2 := lipgloss.JoinHorizontal(lipgloss.Left, m.convertUnsigned8Bit(), m.convertUnsigned32Bit(), m.convertOctal(), m.convertFloat64Bit())
	row3 := lipgloss.JoinHorizontal(lipgloss.Left, m.convertSigned16Bit(), m.convertSigned64Bit(), m.convertBinary(), m.convertOffset())
	row4 := lipgloss.JoinHorizontal(lipgloss.Left, m.convertUnsigned16Bit(), m.convertUnsigned64Bit(), m.convertStreamLength(), m.convertNotification())

	return lipgloss.JoinVertical(lipgloss.Left, row1, row2, row3, row4)
}



func (m HexdumpModel) convertSigned8Bit() string {
	title := m.styles.title.Render("┌Signed 8 Bit")
	line := strings.Repeat("─",  m.styles.converter.GetWidth()-lipgloss.Width(title)-1)
	line = lipgloss.JoinHorizontal(lipgloss.Center, title, line, "┐")
	buf, ok := m.readAt(1)
	if !ok {
		return lipgloss.JoinVertical(lipgloss.Left, line, m.styles.converter.Render(converterShortRead))
	}
	panelContent := m.styles.converter.Render(strconv.FormatInt(int64(int8(buf[0])), 10))

	return lipgloss.JoinVertical(lipgloss.Left, line, panelContent)
}

func (m HexdumpModel) convertSigned32Bit() string {
	title := m.styles.title.Render("┌Signed 32 Bit")
	line := strings.Repeat("─",  m.styles.converter.GetWidth()-lipgloss.Width(title)-1)
	line = lipgloss.JoinHorizontal(lipgloss.Center, title, line, "┐")
	buf, ok := m.readAt(4)
	if !ok {
		return lipgloss.JoinVertical(lipgloss.Left, line, m.styles.converter.Render(converterShortRead))
	}
	v := int32(m.byteOrder.Uint32(buf))
	panelContent := m.styles.converter.Render(strconv.FormatInt(int64(v), 10))

	return lipgloss.JoinVertical(lipgloss.Left, line, panelContent)
}


func (m HexdumpModel) convertHexadecimal() string {
	title := m.styles.title.Render("┌Hexadecimal")
	line := strings.Repeat("─",  m.styles.converter.GetWidth()-lipgloss.Width(title)-1)
	line = lipgloss.JoinHorizontal(lipgloss.Center, title, line, "┐")
	buf, ok := m.readAt(4)
	if !ok {
		return lipgloss.JoinVertical(lipgloss.Left, line, m.styles.converter.Render(converterShortRead))
	}
	panelContent := m.styles.converter.Render(fmt.Sprintf("%08X", m.byteOrder.Uint32(buf)))

	return lipgloss.JoinVertical(lipgloss.Left, line, panelContent)
}

func (m HexdumpModel) convertFloat32Bit() string {
	title := m.styles.title.Render("┌Float 32 bit")
	line := strings.Repeat("─",  m.styles.converter.GetWidth()-lipgloss.Width(title)-1)
	line = lipgloss.JoinHorizontal(lipgloss.Center, title, line, "┐")
	buf, ok := m.readAt(4)
	if !ok {
		return lipgloss.JoinVertical(lipgloss.Left, line, m.styles.converter.Render(converterShortRead))
	}
	u := m.byteOrder.Uint32(buf)
	panelContent := m.styles.converter.Render(strconv.FormatFloat(float64(math.Float32frombits(u)), 'e', -1, 32))

	return lipgloss.JoinVertical(lipgloss.Left, line, panelContent)
}

func (m HexdumpModel) convertUnsigned8Bit() string {
	title := m.styles.title.Render("┌Unsigned 8 Bit")
	line := strings.Repeat("─",  m.styles.converter.GetWidth()-lipgloss.Width(title)-1)
	line = lipgloss.JoinHorizontal(lipgloss.Center, title, line, "┐")
	buf, ok := m.readAt(1)
	if !ok {
		return lipgloss.JoinVertical(lipgloss.Left, line, m.styles.converter.Render(converterShortRead))
	}
	panelContent := m.styles.converter.Render(strconv.FormatUint(uint64(buf[0]), 10))

	return lipgloss.JoinVertical(lipgloss.Left, line, panelContent)
}

func (m HexdumpModel) convertUnsigned32Bit() string {
	title := m.styles.title.Render("┌Unsigned 32 Bit")
	line := strings.Repeat("─",  m.styles.converter.GetWidth()-lipgloss.Width(title)-1)
	line = lipgloss.JoinHorizontal(lipgloss.Center, title, line, "┐")
	buf, ok := m.readAt(4)
	if !ok {
		return lipgloss.JoinVertical(lipgloss.Left, line, m.styles.converter.Render(converterShortRead))
	}
	u := m.byteOrder.Uint32(buf)
	panelContent := m.styles.converter.Render(strconv.FormatUint(uint64(u), 10))

	return lipgloss.JoinVertical(lipgloss.Left, line, panelContent)
}

func (m HexdumpModel) convertOctal() string {
	title := m.styles.title.Render("┌Octal")
	line := strings.Repeat("─",  m.styles.converter.GetWidth()-lipgloss.Width(title)-1)
	line = lipgloss.JoinHorizontal(lipgloss.Center, title, line, "┐")
	buf, ok := m.readAt(1)
	if !ok {
		return lipgloss.JoinVertical(lipgloss.Left, line, m.styles.converter.Render(converterShortRead))
	}
	panelContent := m.styles.converter.Render(strconv.FormatUint(uint64(buf[0]), 8))

	return lipgloss.JoinVertical(lipgloss.Left, line, panelContent)
}

func (m HexdumpModel) convertFloat64Bit() string {
	title := m.styles.title.Render("┌Float 64 bit")
	line := strings.Repeat("─",  m.styles.converter.GetWidth()-lipgloss.Width(title)-1)
	line = lipgloss.JoinHorizontal(lipgloss.Center, title, line, "┐")
	buf, ok := m.readAt(8)
	if !ok {
		return lipgloss.JoinVertical(lipgloss.Left, line, m.styles.converter.Render(converterShortRead))
	}
	u := m.byteOrder.Uint64(buf)
	panelContent := m.styles.converter.Render(strconv.FormatFloat(math.Float64frombits(u), 'e', -1, 64))

	return lipgloss.JoinVertical(lipgloss.Left, line, panelContent)
}

func (m HexdumpModel) convertSigned16Bit() string {
	title := m.styles.title.Render("┌Signed 16 Bit")
	line := strings.Repeat("─",  m.styles.converter.GetWidth()-lipgloss.Width(title)-1)
	line = lipgloss.JoinHorizontal(lipgloss.Center, title, line, "┐")
	buf, ok := m.readAt(2)
	if !ok {
		return lipgloss.JoinVertical(lipgloss.Left, line, m.styles.converter.Render(converterShortRead))
	}
	v := int16(m.byteOrder.Uint16(buf))
	panelContent := m.styles.converter.Render(strconv.FormatInt(int64(v), 10))

	return lipgloss.JoinVertical(lipgloss.Left, line, panelContent)
}

func (m HexdumpModel) convertSigned64Bit() string {
	title := m.styles.title.Render("┌Signed 64 Bit")
	line := strings.Repeat("─",  m.styles.converter.GetWidth()-lipgloss.Width(title)-1)
	line = lipgloss.JoinHorizontal(lipgloss.Center, title, line, "┐")
	buf, ok := m.readAt(8)
	if !ok {
		return lipgloss.JoinVertical(lipgloss.Left, line, m.styles.converter.Render(converterShortRead))
	}
	v := int64(m.byteOrder.Uint64(buf))
	panelContent := m.styles.converter.Render(strconv.FormatInt(v, 10))

	return lipgloss.JoinVertical(lipgloss.Left, line, panelContent)
}

func (m HexdumpModel) convertBinary() string {
	title := m.styles.title.Render("┌Binary")
	line := strings.Repeat("─",  m.styles.converter.GetWidth()-lipgloss.Width(title)-1)
	line = lipgloss.JoinHorizontal(lipgloss.Center, title, line, "┐")
	buf, ok := m.readAt(1)
	if !ok {
		return lipgloss.JoinVertical(lipgloss.Left, line, m.styles.converter.Render(converterShortRead))
	}
	panelContent := m.styles.converter.Render(fmt.Sprintf("%08b", buf[0]))

	return lipgloss.JoinVertical(lipgloss.Left, line, panelContent)
}

func (m HexdumpModel) convertOffset() string {
	title := m.styles.title.Render("┌Offset")
	line := strings.Repeat("─",  m.styles.converter.GetWidth()-lipgloss.Width(title)-1)
	line = lipgloss.JoinHorizontal(lipgloss.Center, title, line, "┐")

	panelContent := m.styles.converter.Render(strconv.FormatInt(int64(m.offset), 16))

	return lipgloss.JoinVertical(lipgloss.Left, line, panelContent)
}

func (m HexdumpModel) convertUnsigned16Bit() string {
	title := m.styles.title.Render("┌Unsigned 16 Bit")
	line := strings.Repeat("─",  m.styles.converter.GetWidth()-lipgloss.Width(title)-1)
	line = lipgloss.JoinHorizontal(lipgloss.Center, title, line, "┐")
	buf, ok := m.readAt(2)
	if !ok {
		return lipgloss.JoinVertical(lipgloss.Left, line, m.styles.converter.Render(converterShortRead))
	}
	u := m.byteOrder.Uint16(buf)
	panelContent := m.styles.converter.Render(strconv.FormatUint(uint64(u), 10))

	return lipgloss.JoinVertical(lipgloss.Left, line, panelContent)
}

func (m HexdumpModel) convertUnsigned64Bit() string {
	title := m.styles.title.Render("┌Unsigned 64 Bit")
	line := strings.Repeat("─",  m.styles.converter.GetWidth()-lipgloss.Width(title)-1)
	line = lipgloss.JoinHorizontal(lipgloss.Center, title, line, "┐")
	buf, ok := m.readAt(8)
	if !ok {
		return lipgloss.JoinVertical(lipgloss.Left, line, m.styles.converter.Render(converterShortRead))
	}
	u := m.byteOrder.Uint64(buf)
	panelContent := m.styles.converter.Render(strconv.FormatUint(u, 10))
	return lipgloss.JoinVertical(lipgloss.Left, line, panelContent)
}


func (m HexdumpModel) convertStreamLength() string {
	title := m.styles.title.Render("┌Stream Length")
	line := strings.Repeat("─",  m.styles.converter.GetWidth()-lipgloss.Width(title)-1)
	line = lipgloss.JoinHorizontal(lipgloss.Center, title, line, "┐")

	panelContent := m.styles.converter.Render(strconv.FormatInt(int64(m.streamLength), 10))

	return lipgloss.JoinVertical(lipgloss.Left, line, panelContent)
}

func (m HexdumpModel) convertNotification() string {
	title := m.styles.title.Render("┌Notifications")
	line := strings.Repeat("─",  m.styles.converter.GetWidth()-lipgloss.Width(title)-1)
	line = lipgloss.JoinHorizontal(lipgloss.Center, title, line, "┐")

	panelContent := m.styles.converter.Render(m.notifications)

	return lipgloss.JoinVertical(lipgloss.Left, line, panelContent)
}
