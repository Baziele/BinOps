package main

import (
	"strings"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type StringsModel struct {
	height int
	width int
	styles struct{
		title lipgloss.Style
	}
	content  string
	viewport viewport.Model
}

func initializeStringsModel(width, height int) StringsModel {
	m := StringsModel{
		width: width,
		height: height,
		content: "Strings",
	}
	m.styles.title = lipgloss.NewStyle().Bold(true)
	m.viewport = viewport.New(viewport.WithWidth(width), viewport.WithHeight(height - lipgloss.Height(m.styles.title.Render("") )))
	m.viewport.SetContent(`
	expandAVX512_48_outShufHi
expandAVX512_52_outShufLo
expandAVX512_52_outShufHi0
expandAVX512_52_outShufHi1
expandAVX512_56_outShufLo
expandAVX512_56_outShufHi
expandAVX512_60_outShufLo
expandAVX512_60_outShufHi0
expandAVX512_60_outShufHi1
expandAVX512_64_outShufLo
github.com/clipperhouse/uax29/v2/graphemes..dict.lookup[string]
github.com/clipperhouse/uax29/v2/graphemes..dict.splitFunc[string]
github.com/clipperhouse/uax29/v2/graphemes..dict.lookup[[]uint8]
github.com/clipperhouse/uax29/v2/graphemes..dict.splitFunc[[]uint8]
go:itab.*encoding/json.UnsupportedValueError,error
go:itab.internal/poll.errNetClosing,error
go:itab.*internal/poll.DeadlineExceededError,error
sync..dict.OnceValue[bool]
internal/abi..dict.Escape[*internal/poll.splicePipe]
internal/abi..dict.Escape[func(internal/poll.splicePipeFields)]
runtime..dict.callCleanup[internal/poll.splicePipeFields]
internal/abi..dict.TypeFor[internal/poll.splicePipe]
runtime..dict.AddCleanup[internal/poll.splicePipe,internal/poll.splicePipeFields]
go:itab.*compress/zlib.reader,io.ReadCloser
go:itab.*bufio.Reader,compress/flate.Reader
go:itab.*hash/adler32.digest,hash.Hash32
go:itab.image/png.FormatError,error
go:itab.image/png.UnsupportedError,error
go:itab.*image/png.decoder,io.Reader
go:itab.*hash/crc32.digest,hash.Hash32
go:itab.os/user.UnknownUserIdError,error
go:itab.*internal/bisect.parseError,error
internal/abi..dict.TypeFor[net/netip.addrDetail]
sync/atomic..dict.Pointer[unique.indirect[net/netip.addrDetail]]
unique..dict.newIndirectNode[net/netip.addrDetail]
unique..dict.newCanonMap[net/netip.addrDetail]
sync/atomic..dict.Pointer[unique.node[net/netip.addrDetail]]
unique..dict.node[net/netip.addrDetail]
weak..dict.Pointer[net/netip.addrDetail]
sync/atomic..dict.Pointer[unique.entry[net/netip.addrDetail]]
unique..dict.entry[net/netip.addrDetail]
internal/abi..dict.Escape[*net/netip.addrDetail]
weak..dict.Make[net/netip.addrDetail]
unique..dict.newEntryNode[net/netip.addrDetail]
internal/abi..dict.Escape[func(struct {})]
runtime..dict.callCleanup[struct {}]
runtime..dict.AddCleanup[net/netip.addrDetail,struct {}]
unique..dict.indirect[net/netip.addrDetail]
unique..dict.canonMap[net/netip.addrDetail]
internal/abi..dict.EscapeNonString[net/netip.addrDetail]
internal/abi..dict.EscapeToResultNonString[net/netip.addrDetail]
unique..dict.clone[net/netip.addrDetail]
unique..dict.Make[net/netip.addrDetail]
sync/atomic..dict.Pointer[internal/sync.indirect[*internal/abi.Type,interface {}]]
internal/sync..dict.newIndirectNode[*internal/abi.Type,interface {}]
sync/atomic..dict.Pointer[internal/sync.node[*internal/abi.Type,interface {}]]
internal/sync..dict.node[*internal/abi.Type,interface {}]
sync/atomic..dict.Pointer[internal/sync.entry[*internal/abi.Type,interface {}]]
internal/sync..dict.newEntryNode[*internal/abi.Type,interface {}]
internal/sync..dict.entry[*internal/abi.Type,interface {}]
internal/sync..dict.indirect[*internal/abi.Type,interface {}]
internal/sync..dict.HashTrieMap[*internal/abi.Type,interface {}]
go:itab.*regexp.inputReader,regexp.input
go:itab.*regexp.inputString,regexp.input
go:itab.*regexp.inputBytes,regexp.input
cmp..dict.isNaN[int32]
cmp..dict.Less[int32]
slices..dict.insertionSortOrdered[int32]
slices..dict.siftDownOrdered[int32]
slices..dict.heapSortOrdered[int32]
slices..dict.breakPatternsOrdered[int32]
slices..dict.order2Ordered[int32]
slices..dict.medianOrdered[int32]
slices..dict.medianAdjacentOrdered[int32]
slices..dict.choosePivotOrdered[int32]
slices..dict.reverseRangeOrdered[int32]
slices..dict.partialInsertionSortOrdered[int32]
slices..dict.partitionEqualOrdered[int32]
slices..dict.partitionOrdered[int32]
slices..dict.pdqsortOrdered[int32]
go:itab.*compress/flate.decompressor,io.ReadCloser
go:itab.compress/flate.CorruptInputError,error
go:itab.compress/flate.InternalError,error
go:itab.*compress/flate.byFreq,sort.Interface
go:itab.*compress/flate.byLiteral,sort.Interface
x_cgo_pthread_key_created
x_crosscall2_ptr
_crosscall2_ptr
go:itab.*regexp/syntax.Error,error
go:itab.regexp/syntax.ranges,sort.Interface
getpwuid_r
getgrgid_r
getpwnam_r
sysconf
getgrnam_r
free
pthread_create
pthread_sigmask
__vfprintf_chk
abort
__errno_location
pthread_getattr_np
pthread_cond_broadcast
sigaction
setenv
pthread_cond_wait
mmap
pthread_setspecific
nanosleep
pthread_attr_getstack
fputc
pthread_attr_init
pthread_attr_getstacksize
sigemptyset
sigfillset
pthread_attr_setdetachstate
pthread_mutex_unlock
malloc
munmap
pthread_key_create
pthread_self
unsetenv
pthread_attr_destroy
sigismember
fwrite
__fprintf_chk
strerror
sigaddset
pthread_mutex_lock
stderr
x_cgo_inittls
go:fipsinfo
go:main.inittasks
go:runtime.inittasks
runtime.defaultGOROOT.str
runtime.buildVersion.str
runtime.modinfo.str
go:buildinfo
go:buildinfo.ref
type:*
runtime.textsectionmap
.rela
.rela.plt
.tbss
.note.go.buildid
.note.gnu.build-id
.interp
.text
.rodata
.gopclntab
.gnu.version
.gnu.version_r
.hash
.dynstr
.dynsym
.typelink
.itablink
.go.buildinfo
.go.fipsinfo
.dynamic
.got.plt
.go.module
.got
.noptrdata
.data
.bss
.noptrbss
.debug_abbrev
.debug_line
.debug_frame
.debug_gdb_scripts
.debug_info
.debug_loclists
.debug_rnglists
.debug_addr
.symtab
.strtab
.shstrtab

	`)

	return m
}


func (m StringsModel) Init() tea.Cmd{
	
	return nil
}

func (m StringsModel) Update(msg tea.Msg) (StringsModel, tea.Cmd){
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}


func (m StringsModel) View() string{

	// m.viewport.sethe
	return lipgloss.JoinVertical(lipgloss.Left, m.titleView(), m.viewport.View(), m.footerView())
}

func (m StringsModel) titleView() string{
	title := m.styles.title.Render("┌┤Strings├")
	line := strings.Repeat("─", max(0, m.width - lipgloss.Width(title) - 1))
	return lipgloss.JoinHorizontal(lipgloss.Center, title, line, "┐")
}

func (m StringsModel) footerView() string{
	info := m.styles.title.Render(lipgloss.Sprintf("%3.f%%:%3.f%%", m.viewport.ScrollPercent()*100, m.viewport.HorizontalScrollPercent()*100))
	line := strings.Repeat("─", max(0, m.viewport.Width()-lipgloss.Width(info)-3))
	return lipgloss.JoinHorizontal(lipgloss.Center, line,"┤", info, "├┘",)
}

func (m *StringsModel) setDimensions(width, height int){
	m.width = width
	m.height = height
	m.viewport.SetWidth(width)
	m.viewport.SetHeight(height)
}

