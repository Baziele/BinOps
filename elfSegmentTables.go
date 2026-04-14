package main

import (
	"fmt"
	"strconv"

	"debug/elf"
)

// ELFStaticTable holds column titles and row cells for one static-analysis table.
// Rows hold compact display values; CellDetails holds the same shape with long
// descriptions (empty string when nothing more than the cell text is needed).
type ELFStaticTable struct {
	Headers     []string
	Rows        [][]string
	CellDetails [][]string
}

func emptyELFTable(message string) ELFStaticTable {
	if message == "" {
		return ELFStaticTable{}
	}
	return ELFStaticTable{
		Headers:     []string{"Note"},
		Rows:        [][]string{{message}},
		CellDetails: nil,
	}
}

// ParseELFSegmentTables returns six tables in the same order as StaticModel.fileSegments.
func ParseELFSegmentTables(f *elf.File) []ELFStaticTable {
	out := make([]ELFStaticTable, 6)
	if f == nil {
		for i := range out {
			out[i] = emptyELFTable("No ELF file loaded.")
		}
		return out
	}
	out[0] = parseProgramHeaders(f)
	out[1] = parseSectionHeaders(f)
	out[2] = parseSymbolsTable(f)
	out[3] = parseDynamicSymbolsTable(f)
	out[4] = parseDynamicTable(f)
	out[5] = parseRelocationsTable(f)
	return out
}

func parseProgramHeaders(f *elf.File) ELFStaticTable {
	t := ELFStaticTable{
		Headers: []string{"Type", "Flags", "Offset", "VAddr", "PAddr", "FileSz", "MemSz", "Align"},
	}
	for _, p := range f.Progs {
		t.Rows = append(t.Rows, []string{
			compactProgType(p.Type),
			compactProgFlags(p.Flags),
			fmt.Sprintf("0x%x", p.Off),
			fmt.Sprintf("0x%x", p.Vaddr),
			fmt.Sprintf("0x%x", p.Paddr),
			fmt.Sprintf("0x%x", p.Filesz),
			fmt.Sprintf("0x%x", p.Memsz),
			fmt.Sprintf("0x%x", p.Align),
		})
		t.CellDetails = append(t.CellDetails, []string{
			readableProgType(p.Type),
			readableProgFlags(p.Flags),
			"",
			"",
			"",
			"",
			"",
			"",
		})
	}
	if len(t.Rows) == 0 {
		return emptyELFTable("No program headers.")
	}
	return t
}

func parseSectionHeaders(f *elf.File) ELFStaticTable {
	t := ELFStaticTable{
		Headers: []string{"#", "Name", "Type", "Flags", "Addr", "Off", "Size", "Link", "Info", "Align", "EntSize"},
	}
	for i, s := range f.Sections {
		name := s.Name
		if name == "" {
			name = fmt.Sprintf("<section #%d>", i)
		}
		t.Rows = append(t.Rows, []string{
			strconv.Itoa(i),
			name,
			compactSectionType(s.Type),
			compactSectionFlags(s.Flags),
			fmt.Sprintf("0x%x", s.Addr),
			fmt.Sprintf("0x%x", s.Offset),
			fmt.Sprintf("0x%x", s.Size),
			fmt.Sprintf("%d", s.Link),
			fmt.Sprintf("%d", s.Info),
			fmt.Sprintf("0x%x", s.Addralign),
			fmt.Sprintf("0x%x", s.Entsize),
		})
		t.CellDetails = append(t.CellDetails, []string{
			"",
			"",
			readableSectionType(s.Type),
			readableSectionFlags(s.Flags),
			"",
			"",
			"",
			"",
			"",
			"",
			"",
		})
	}
	if len(t.Rows) == 0 {
		return emptyELFTable("No section headers.")
	}
	return t
}

func parseSymbolsTable(f *elf.File) ELFStaticTable {
	t := ELFStaticTable{
		Headers: []string{"Name", "Value", "Size", "Type", "Bind", "Vis", "Section"},
	}
	syms, err := f.Symbols()
	if err != nil {
		if err == elf.ErrNoSymbols {
			return emptyELFTable("No symbol table (.symtab).")
		}
		return emptyELFTable(fmt.Sprintf("Symbols: %v", err))
	}
	for _, s := range syms {
		row, det := symbolRow(s, false)
		t.Rows = append(t.Rows, row)
		t.CellDetails = append(t.CellDetails, det)
	}
	return t
}

func parseDynamicSymbolsTable(f *elf.File) ELFStaticTable {
	t := ELFStaticTable{
		Headers: []string{"Name", "Value", "Size", "Type", "Bind", "Vis", "Section", "Version", "Library"},
	}
	syms, err := f.DynamicSymbols()
	if err != nil {
		if err == elf.ErrNoSymbols {
			return emptyELFTable("No dynamic symbol table (.dynsym).")
		}
		return emptyELFTable(fmt.Sprintf("Dynamic symbols: %v", err))
	}
	for _, s := range syms {
		row, det := symbolRow(s, true)
		t.Rows = append(t.Rows, row)
		t.CellDetails = append(t.CellDetails, det)
	}
	return t
}

func symbolRow(s elf.Symbol, dyn bool) (row []string, detail []string) {
	st := elf.ST_TYPE(s.Info)
	row = []string{
		s.Name,
		fmt.Sprintf("0x%x", s.Value),
		fmt.Sprintf("0x%x", s.Size),
		compactSymType(st),
		readableSymBind(elf.ST_BIND(s.Info)),
		readableSymVis(elf.ST_VISIBILITY(s.Other)),
		compactSectionIndex(s.Section),
	}
	detail = []string{
		"",
		"",
		"",
		readableSymType(st),
		"",
		"",
		readableSectionIndex(s.Section),
	}
	if dyn {
		ver, lib := "", ""
		if s.HasVersion {
			ver = s.Version
			lib = s.Library
		}
		row = append(row, ver, lib)
		detail = append(detail, "", "")
	}
	return row, detail
}

func elfStrTabBytes(f *elf.File, link uint32) []byte {
	if link >= uint32(len(f.Sections)) {
		return nil
	}
	b, err := f.Sections[link].Data()
	if err != nil {
		return nil
	}
	return b
}

func stringFromStrtab(tab []byte, off int) string {
	if tab == nil || off < 0 || off >= len(tab) {
		return ""
	}
	end := off
	for end < len(tab) && tab[end] != 0 {
		end++
	}
	return string(tab[off:end])
}

func dynTagValueString(tag elf.DynTag, val uint64, strtab []byte) string {
	switch tag {
	case elf.DT_NULL:
		return "0"
	case elf.DT_NEEDED, elf.DT_SONAME, elf.DT_RPATH, elf.DT_RUNPATH, elf.DT_AUXILIARY, elf.DT_FILTER:
		s := stringFromStrtab(strtab, int(val))
		if s != "" {
			return s
		}
		return fmt.Sprintf("0x%x", val)
	case elf.DT_PLTREL:
		return elf.DynTag(val).String()
	default:
		return fmt.Sprintf("0x%x", val)
	}
}

func parseDynamicTable(f *elf.File) ELFStaticTable {
	t := ELFStaticTable{
		Headers: []string{"Tag", "Value"},
	}
	ds := f.SectionByType(elf.SHT_DYNAMIC)
	if ds == nil {
		return emptyELFTable("No dynamic section (.dynamic).")
	}
	d, err := ds.Data()
	if err != nil {
		return emptyELFTable(fmt.Sprintf(".dynamic: %v", err))
	}
	strtab := elfStrTabBytes(f, ds.Link)

	ent := 8
	if f.Class == elf.ELFCLASS64 {
		ent = 16
	}
	if len(d)%ent != 0 {
		return emptyELFTable("Invalid .dynamic entry size.")
	}

	for len(d) > 0 {
		var tag elf.DynTag
		var val uint64
		switch f.Class {
		case elf.ELFCLASS32:
			tag = elf.DynTag(f.ByteOrder.Uint32(d[0:4]))
			val = uint64(f.ByteOrder.Uint32(d[4:8]))
			d = d[8:]
		case elf.ELFCLASS64:
			tag = elf.DynTag(f.ByteOrder.Uint64(d[0:8]))
			val = f.ByteOrder.Uint64(d[8:16])
			d = d[16:]
		default:
			return emptyELFTable("Unsupported ELF class for .dynamic.")
		}
		if tag == elf.DT_NULL {
			break
		}
		t.Rows = append(t.Rows, []string{
			compactDynTag(tag),
			dynTagValueString(tag, val, strtab),
		})
		t.CellDetails = append(t.CellDetails, []string{
			readableDynTag(tag),
			"",
		})
	}
	if len(t.Rows) == 0 {
		return emptyELFTable("Empty .dynamic section.")
	}
	return t
}

func relocTypeString(f *elf.File, typ uint32) string {
	switch f.Machine {
	case elf.EM_X86_64:
		return elf.R_X86_64(typ).String()
	case elf.EM_386:
		return elf.R_386(typ).String()
	case elf.EM_AARCH64:
		return elf.R_AARCH64(typ).String()
	case elf.EM_ARM:
		return elf.R_ARM(typ).String()
	case elf.EM_RISCV:
		return elf.R_RISCV(typ).String()
	case elf.EM_PPC64:
		return elf.R_PPC64(typ).String()
	case elf.EM_PPC:
		return elf.R_PPC(typ).String()
	case elf.EM_S390:
		return elf.R_390(typ).String()
	default:
		return fmt.Sprintf("0x%x", typ)
	}
}

func symbolNameAt(f *elf.File, symSec *elf.Section, symIdx uint32) string {
	if symSec == nil || symIdx == 0 {
		return ""
	}
	data, err := symSec.Data()
	if err != nil {
		return "?"
	}
	strdata := elfStrTabBytes(f, symSec.Link)
	switch f.Class {
	case elf.ELFCLASS64:
		off := int(symIdx) * elf.Sym64Size
		if off+elf.Sym64Size > len(data) {
			return "?"
		}
		nameOff := f.ByteOrder.Uint32(data[off : off+4])
		return stringFromStrtab(strdata, int(nameOff))
	case elf.ELFCLASS32:
		off := int(symIdx) * elf.Sym32Size
		if off+elf.Sym32Size > len(data) {
			return "?"
		}
		nameOff := f.ByteOrder.Uint32(data[off : off+4])
		return stringFromStrtab(strdata, int(nameOff))
	default:
		return "?"
	}
}

func relocTargetSectionName(f *elf.File, relSec *elf.Section) string {
	info := relSec.Info
	if info >= uint32(len(f.Sections)) {
		return "?"
	}
	n := f.Sections[info].Name
	if n == "" {
		return fmt.Sprintf("#%d", info)
	}
	return n
}

func parseRelocationsTable(f *elf.File) ELFStaticTable {
	t := ELFStaticTable{
		Headers: []string{"Section", "Kind", "Offset", "SymIdx", "Symbol", "Type", "Addend"},
	}
	for _, sec := range f.Sections {
		if sec.Type != elf.SHT_REL && sec.Type != elf.SHT_RELA {
			continue
		}
		if sec.Link >= uint32(len(f.Sections)) {
			continue
		}
		data, err := sec.Data()
		if err != nil || len(data) == 0 {
			continue
		}
		symSec := f.Sections[sec.Link]
		secName := sec.Name
		if secName == "" {
			secName = fmt.Sprintf("<reloc #%d>", sec.Offset)
		}
		target := relocTargetSectionName(f, sec)
		loc := secName
		if target != "" && target != "?" {
			loc = secName + " → " + target
		}

		switch f.Class {
		case elf.ELFCLASS64:
			if sec.Type == elf.SHT_RELA {
				const rela64 = 8 + 8 + 8
				for i := 0; i+rela64 <= len(data); i += rela64 {
					chunk := data[i : i+rela64]
					off := f.ByteOrder.Uint64(chunk[0:8])
					info := f.ByteOrder.Uint64(chunk[8:16])
					addend := f.ByteOrder.Uint64(chunk[16:24])
					symIdx := elf.R_SYM64(info)
					rtyp := elf.R_TYPE64(info)
					t.Rows = append(t.Rows, []string{
						loc,
						"RELA",
						fmt.Sprintf("0x%x", off),
						fmt.Sprintf("%d", symIdx),
						symbolNameAt(f, symSec, symIdx),
						compactRelocType(f, rtyp),
						fmt.Sprintf("0x%x", addend),
					})
					t.CellDetails = append(t.CellDetails, []string{
						"", "", "", "", "", relocTypeString(f, rtyp), "",
					})
				}
			} else {
				const rel64 = 8 + 8
				for i := 0; i+rel64 <= len(data); i += rel64 {
					chunk := data[i : i+rel64]
					off := f.ByteOrder.Uint64(chunk[0:8])
					info := f.ByteOrder.Uint64(chunk[8:16])
					symIdx := elf.R_SYM64(info)
					rtyp := elf.R_TYPE64(info)
					t.Rows = append(t.Rows, []string{
						loc,
						"REL",
						fmt.Sprintf("0x%x", off),
						fmt.Sprintf("%d", symIdx),
						symbolNameAt(f, symSec, symIdx),
						compactRelocType(f, rtyp),
						"—",
					})
					t.CellDetails = append(t.CellDetails, []string{
						"", "", "", "", "", relocTypeString(f, rtyp), "",
					})
				}
			}
		case elf.ELFCLASS32:
			if sec.Type == elf.SHT_RELA {
				for i := 0; i+12 <= len(data); i += 12 {
					chunk := data[i : i+12]
					off := f.ByteOrder.Uint32(chunk[0:4])
					info := f.ByteOrder.Uint32(chunk[4:8])
					addend := int32(f.ByteOrder.Uint32(chunk[8:12]))
					symIdx := elf.R_SYM32(info)
					rtyp := elf.R_TYPE32(info)
					t.Rows = append(t.Rows, []string{
						loc,
						"RELA",
						fmt.Sprintf("0x%x", off),
						fmt.Sprintf("%d", symIdx),
						symbolNameAt(f, symSec, symIdx),
						compactRelocType(f, rtyp),
						fmt.Sprintf("0x%x", uint32(addend)),
					})
					t.CellDetails = append(t.CellDetails, []string{
						"", "", "", "", "", relocTypeString(f, rtyp), "",
					})
				}
			} else {
				for i := 0; i+8 <= len(data); i += 8 {
					chunk := data[i : i+8]
					off := f.ByteOrder.Uint32(chunk[0:4])
					info := f.ByteOrder.Uint32(chunk[4:8])
					symIdx := elf.R_SYM32(info)
					rtyp := elf.R_TYPE32(info)
					t.Rows = append(t.Rows, []string{
						loc,
						"REL",
						fmt.Sprintf("0x%x", off),
						fmt.Sprintf("%d", symIdx),
						symbolNameAt(f, symSec, symIdx),
						compactRelocType(f, rtyp),
						"—",
					})
					t.CellDetails = append(t.CellDetails, []string{
						"", "", "", "", "", relocTypeString(f, rtyp), "",
					})
				}
			}
		}
	}
	if len(t.Rows) == 0 {
		return emptyELFTable("No relocation sections (SHT_REL / SHT_RELA).")
	}
	return t
}
