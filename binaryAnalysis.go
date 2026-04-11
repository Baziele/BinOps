package main

import (
	"encoding/binary"
	"fmt"
	"os"
	"strconv"

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

// ELFAnalysis bundles an opened ELF file and pre-parsed header summary.
type ELFAnalysis struct {
	File   *elf.File
	Header ELFStaticHeader
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

// AnalyzeELF opens the binary, parses it with debug/elf, and builds ELFAnalysis. The caller owns the returned *elf.File and must Close it when finished.
func AnalyzeELF(binaryPath string) (*ELFAnalysis, error) {
	f, err := elf.Open(binaryPath)
	if err != nil {
		return nil, err
	}
	header := ParseELFStaticHeader(binaryPath, f)
	return &ELFAnalysis{File: f, Header: header}, nil
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
