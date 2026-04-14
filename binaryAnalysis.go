package main

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"strings"

	"debug/elf"

	tea "charm.land/bubbletea/v2"
)

// ELFStaticHeader holds display-ready ELF file header fields for the static UI.
type ELFStaticHeader struct {
	Class                string
	Endianness           string
	Version              string
	OSABI                string
	ABIVersion           string
	Type                 string
	Arch                 string
	EntryAddress         string
	ProgHeaderStart      string
	SectionHeaderStart   string
	Flags                string
	EhsizeBytes          string
	PhEntSizeBytes       string
	ProgramHeaderCount   string
}

// GNU note types (owner name "GNU"); see ELF gABI and binutils readelf.
const (
	ntGNUABIVersion       uint32 = 1 // NT_GNU_ABI_TAG
	ntGNUHwCaps           uint32 = 2
	ntGNUBuildID          uint32 = 3 // NT_GNU_BUILD_ID
	ntGNUGoldVersion      uint32 = 4
	ntGNUPropertyType0    uint32 = 5 // NT_GNU_PROPERTY_TYPE_0
	gnuPropertyX86Feat1And uint32 = 0xc0000002
)

// ELFNoteEntry is one note record inside a SHT_NOTE section.
type ELFNoteEntry struct {
	Owner       string
	DataSize    uint32
	Description string
	Details     []string
}

// ELFNoteSection groups notes from one ELF section (e.g. .note.gnu.build-id).
type ELFNoteSection struct {
	SectionName string
	Entries     []ELFNoteEntry
}

// ELFAnalysis bundles an opened ELF file and pre-parsed UI text.
type ELFAnalysis struct {
	File           *elf.File
	Header         ELFStaticHeader
	NoteSections   []ELFNoteSection
	SegmentTables  []ELFStaticTable // same order as StaticModel.fileSegments
}

func elfClassDisplay(c elf.Class) string {
	switch c {
	case elf.ELFCLASS32:
		return "ELF32"
	case elf.ELFCLASS64:
		return "ELF64"
	default:
		return c.String()
	}
}

func elfEndiannessDisplay(order binary.ByteOrder) string {
	switch order {
	case binary.LittleEndian:
		return "Little"
	case binary.BigEndian:
		return "Big"
	default:
		return order.String()
	}
}

func elfVersionDisplay(v elf.Version) string {
	switch v {
	case elf.EV_CURRENT:
		return fmt.Sprintf("%d (current)", int(v))
	case elf.EV_NONE:
		return fmt.Sprintf("%d (none)", int(v))
	default:
		return strconv.Itoa(int(v))
	}
}

// readELFHeaderLayout returns file-layout fields not exposed on debug/elf.FileHeader.
func readELFHeaderLayout(path string, class elf.Class, order binary.ByteOrder) (phoff, shoff uint64, flags uint32, ehsize, phentsize, phnum uint16, ok bool) {
	f, err := os.Open(path)
	if err != nil {
		return 0, 0, 0, 0, 0, 0, false
	}
	defer f.Close()

	var buf [64]byte
	if _, err := f.ReadAt(buf[:], 0); err != nil {
		return 0, 0, 0, 0, 0, 0, false
	}
	if string(buf[0:4]) != "\x7fELF" {
		return 0, 0, 0, 0, 0, 0, false
	}

	switch class {
	case elf.ELFCLASS64:
		phoff = order.Uint64(buf[32:40])
		shoff = order.Uint64(buf[40:48])
		flags = order.Uint32(buf[48:52])
		ehsize = order.Uint16(buf[52:54])
		phentsize = order.Uint16(buf[54:56])
		phnum = order.Uint16(buf[56:58])
		return phoff, shoff, flags, ehsize, phentsize, phnum, true
	case elf.ELFCLASS32:
		phoff = uint64(order.Uint32(buf[28:32]))
		shoff = uint64(order.Uint32(buf[32:36]))
		flags = order.Uint32(buf[36:40])
		ehsize = order.Uint16(buf[40:42])
		phentsize = order.Uint16(buf[42:44])
		phnum = order.Uint16(buf[44:46])
		return phoff, shoff, flags, ehsize, phentsize, phnum, true
	default:
		return 0, 0, 0, 0, 0, 0, false
	}
}

// ParseELFStaticHeader fills ELFStaticHeader from an opened ELF file and optional on-disk path for layout fields.
func ParseELFStaticHeader(binaryPath string, f *elf.File) ELFStaticHeader {
	if f == nil {
		return ELFStaticHeader{}
	}
	fh := f.FileHeader
	h := ELFStaticHeader{
		Class:        elfClassDisplay(fh.Class),
		Endianness:   elfEndiannessDisplay(fh.ByteOrder),
		Version:      elfVersionDisplay(fh.Version),
		OSABI:        fh.OSABI.String(),
		ABIVersion:   strconv.Itoa(int(fh.ABIVersion)),
		Type:         fh.Type.String(),
		Arch:         fh.Machine.String(),
		EntryAddress: fmt.Sprintf("0x%x", fh.Entry),
	}

	phoff, shoff, flags, ehsize, phentsize, phnumDisk, ok := readELFHeaderLayout(binaryPath, fh.Class, fh.ByteOrder)
	phnum := int(phnumDisk)
	if n := len(f.Progs); n > 0 {
		phnum = n
	}
	h.ProgramHeaderCount = strconv.Itoa(phnum)

	if ok {
		h.ProgHeaderStart = fmt.Sprintf("%d (bytes into file)", phoff)
		h.SectionHeaderStart = fmt.Sprintf("%d (bytes into file)", shoff)
		h.Flags = fmt.Sprintf("0x%x", flags)
		h.EhsizeBytes = fmt.Sprintf("%d (bytes)", ehsize)
		h.PhEntSizeBytes = fmt.Sprintf("%d (bytes)", phentsize)
	}
	return h
}

func alignNote4(n int) int {
	return (n + 3) &^ 3
}

func alignNote8(n int) int {
	return (n + 7) &^ 7
}

func elfNoteOSName(os uint32) string {
	switch os {
	case 0:
		return "Linux"
	case 1:
		return "GNU"
	case 2:
		return "Solaris2"
	case 3:
		return "FreeBSD"
	default:
		return fmt.Sprintf("unknown (%d)", os)
	}
}

func parseGNUProperties(desc []byte, order binary.ByteOrder) []string {
	var lines []string
	i := 0
	for i+8 <= len(desc) {
		propType := order.Uint32(desc[i : i+4])
		sz := int(order.Uint32(desc[i+4 : i+8]))
		i += 8
		if sz < 0 || i+sz > len(desc) {
			break
		}
		data := desc[i : i+sz]
		i += sz
		i = alignNote8(i)

		switch propType {
		case gnuPropertyX86Feat1And:
			if len(data) >= 4 {
				feat := order.Uint32(data[:4])
				var parts []string
				if feat&1 != 0 {
					parts = append(parts, "IBT")
				}
				if feat&2 != 0 {
					parts = append(parts, "SHSTK")
				}
				if feat&4 != 0 {
					parts = append(parts, "LAM_U48")
				}
				if feat&8 != 0 {
					parts = append(parts, "LAM_U57")
				}
				if len(parts) > 0 {
					lines = append(lines, fmt.Sprintf("      Properties: x86 feature: %s", strings.Join(parts, ", ")))
				}
			}
		}
	}
	return lines
}

func describeGNUNote(ntype uint32, desc []byte, order binary.ByteOrder) (description string, details []string) {
	switch ntype {
	case ntGNUABIVersion:
		description = "NT_GNU_ABI_TAG (ABI version tag)"
		if len(desc) >= 16 {
			osTag := order.Uint32(desc[0:4])
			major := order.Uint32(desc[4:8])
			minor := order.Uint32(desc[8:12])
			patch := order.Uint32(desc[12:16])
			details = append(details, fmt.Sprintf("    OS: %s, ABI: %d.%d.%d",
				elfNoteOSName(osTag), major, minor, patch))
		}
	case ntGNUBuildID:
		description = "NT_GNU_BUILD_ID (unique build ID bitstring)"
		if len(desc) > 0 {
			details = append(details, fmt.Sprintf("    Build ID: %s", hex.EncodeToString(desc)))
		}
	case ntGNUPropertyType0:
		description = "NT_GNU_PROPERTY_TYPE_0"
		details = append(details, parseGNUProperties(desc, order)...)
	case ntGNUHwCaps:
		description = "NT_GNU_HWCAPS"
	case ntGNUGoldVersion:
		description = "NT_GNU_GOLD_VERSION"
	default:
		description = fmt.Sprintf("Unknown note type: (0x%08x)", ntype)
	}
	return description, details
}

func describeNote(owner string, ntype uint32, desc []byte, order binary.ByteOrder) ELFNoteEntry {
	owner = strings.TrimRight(owner, "\x00")
	if owner == "" {
		owner = "?"
	}
	entry := ELFNoteEntry{
		Owner:    owner,
		DataSize: uint32(len(desc)),
	}
	switch owner {
	case "GNU":
		entry.Description, entry.Details = describeGNUNote(ntype, desc, order)
	default:
		entry.Description = fmt.Sprintf("note type 0x%08x", ntype)
		if len(desc) > 0 && len(desc) <= 64 {
			entry.Details = append(entry.Details, fmt.Sprintf("    %s", hex.EncodeToString(desc)))
		}
	}
	return entry
}

func parseNoteSectionData(data []byte, order binary.ByteOrder) []ELFNoteEntry {
	var entries []ELFNoteEntry
	off := 0
	for off+12 <= len(data) {
		namesz := int(order.Uint32(data[off : off+4]))
		descsz := int(order.Uint32(data[off+4 : off+8]))
		ntype := order.Uint32(data[off+8 : off+12])
		off += 12
		if namesz < 0 || off+namesz > len(data) {
			break
		}
		nameBytes := data[off : off+namesz]
		off += namesz
		off = alignNote4(off)
		if descsz < 0 || off+descsz > len(data) {
			break
		}
		desc := append([]byte(nil), data[off:off+descsz]...)
		off += descsz
		off = alignNote4(off)
		owner := string(bytes.TrimRight(nameBytes, "\x00"))
		entries = append(entries, describeNote(owner, ntype, desc, order))
	}
	return entries
}

// ParseELFNotes collects SHT_NOTE sections and decodes note records like readelf --notes.
func ParseELFNotes(f *elf.File) []ELFNoteSection {
	if f == nil {
		return nil
	}
	var out []ELFNoteSection
	for i, sec := range f.Sections {
		if sec.Type != elf.SHT_NOTE {
			continue
		}
		data, err := sec.Data()
		if err != nil || len(data) == 0 {
			continue
		}
		entries := parseNoteSectionData(data, f.ByteOrder)
		if len(entries) == 0 {
			continue
		}
		name := sec.Name
		if name == "" {
			name = fmt.Sprintf("<section #%d>", i)
		}
		out = append(out, ELFNoteSection{SectionName: name, Entries: entries})
	}
	return out
}


// AnalyzeELF opens the binary, parses it with debug/elf, and builds ELFAnalysis. The caller owns the returned *elf.File and must Close it when finished.
func AnalyzeELF(binaryPath string) (*ELFAnalysis, error) {
	f, err := elf.Open(binaryPath)
	if err != nil {
		return nil, err
	}
	header := ParseELFStaticHeader(binaryPath, f)
	notes := ParseELFNotes(f)
	tables := ParseELFSegmentTables(f)
	return &ELFAnalysis{File: f, Header: header, NoteSections: notes, SegmentTables: tables}, nil
}

func analyzeBinary(binaryName string) tea.Cmd {
	return func() tea.Msg {
		a, err := AnalyzeELF(binaryName)
		if err != nil {
			return err
		}
		return *a
	}
}
