package main

import (
	"fmt"
	"strings"

	"debug/elf"
)

// --- Compact display strings for table cells (full text via readable* in CellDetails) ---

func compactProgType(t elf.ProgType) string {
	return t.String()
}

func compactProgFlags(f elf.ProgFlag) string {
	if f == 0 {
		return "—"
	}
	var b strings.Builder
	if f&elf.PF_R != 0 {
		b.WriteByte('r')
	} else {
		b.WriteByte('-')
	}
	if f&elf.PF_W != 0 {
		b.WriteByte('w')
	} else {
		b.WriteByte('-')
	}
	if f&elf.PF_X != 0 {
		b.WriteByte('x')
	} else {
		b.WriteByte('-')
	}
	rest := f &^ (elf.PF_R | elf.PF_W | elf.PF_X)
	if rest != 0 {
		b.WriteString(fmt.Sprintf("+0x%x", uint32(rest)))
	}
	return b.String()
}

func compactSectionType(t elf.SectionType) string {
	return t.String()
}

func compactSectionFlags(f elf.SectionFlag) string {
	if f == 0 {
		return "—"
	}
	return fmt.Sprintf("0x%x", uint64(f))
}

func compactSymType(t elf.SymType) string {
	switch t {
	case elf.STT_NOTYPE:
		return "none"
	case elf.STT_OBJECT:
		return "obj"
	case elf.STT_FUNC:
		return "func"
	case elf.STT_SECTION:
		return "sect"
	case elf.STT_FILE:
		return "file"
	case elf.STT_COMMON:
		return "common"
	case elf.STT_TLS:
		return "tls"
	case elf.STT_GNU_IFUNC:
		return "ifunc"
	default:
		return t.String()
	}
}

func compactSectionIndex(i elf.SectionIndex) string {
	switch i {
	case elf.SHN_UNDEF:
		return "UND"
	case elf.SHN_ABS:
		return "ABS"
	case elf.SHN_COMMON:
		return "COM"
	case elf.SHN_XINDEX:
		return "XIX"
	default:
		if i < 0xff00 {
			return fmt.Sprintf("%d", int(i))
		}
		return fmt.Sprintf("0x%x", uint32(i))
	}
}

func compactDynTag(tag elf.DynTag) string {
	return tag.String()
}

func compactRelocType(f *elf.File, typ uint32) string {
	return fmt.Sprintf("0x%x", typ)
}

// readableSectionType maps standard ELF section types to short descriptions.
func readableSectionType(t elf.SectionType) string {
	switch t {
	case elf.SHT_NULL:
		return "Inactive (unused)"
	case elf.SHT_PROGBITS:
		return "Program data or code"
	case elf.SHT_SYMTAB:
		return "Symbol table"
	case elf.SHT_STRTAB:
		return "String table"
	case elf.SHT_RELA:
		return "Relocations (with addends)"
	case elf.SHT_HASH:
		return "Symbol hash table"
	case elf.SHT_DYNAMIC:
		return "Dynamic linking info"
	case elf.SHT_NOTE:
		return "Notes"
	case elf.SHT_NOBITS:
		return "No bits in file (BSS-like)"
	case elf.SHT_REL:
		return "Relocations (no addends)"
	case elf.SHT_SHLIB:
		return "Reserved (shared lib)"
	case elf.SHT_DYNSYM:
		return "Dynamic symbol table"
	case elf.SHT_INIT_ARRAY:
		return "Initialization function pointers"
	case elf.SHT_FINI_ARRAY:
		return "Termination function pointers"
	case elf.SHT_PREINIT_ARRAY:
		return "Pre-initialization function pointers"
	case elf.SHT_GROUP:
		return "Section group"
	case elf.SHT_SYMTAB_SHNDX:
		return "Extended section indexes"
	case elf.SHT_GNU_ATTRIBUTES:
		return "Object attributes (GNU)"
	case elf.SHT_GNU_HASH:
		return "GNU hash table"
	case elf.SHT_GNU_LIBLIST:
		return "GNU prelink library list"
	case elf.SHT_GNU_VERDEF:
		return "GNU version definitions"
	case elf.SHT_GNU_VERNEED:
		return "GNU version requirements"
	case elf.SHT_GNU_VERSYM:
		return "GNU version symbol indices"
	case elf.SHT_RISCV_ATTRIBUTES:
		return "RISC-V attributes"
	case elf.SHT_MIPS_ABIFLAGS:
		return "MIPS ABI flags"
	default:
		return t.String()
	}
}

// readableSectionFlags expands SHF_* bits into plain language.
func readableSectionFlags(f elf.SectionFlag) string {
	if f == 0 {
		return "—"
	}
	var parts []string
	var known elf.SectionFlag
	add := func(mask elf.SectionFlag, label string) {
		if f&mask != 0 {
			parts = append(parts, label)
			known |= mask
		}
	}
	add(elf.SHF_WRITE, "Writable")
	add(elf.SHF_ALLOC, "Loaded into memory")
	add(elf.SHF_EXECINSTR, "Executable")
	add(elf.SHF_MERGE, "Mergeable")
	add(elf.SHF_STRINGS, "Nul-terminated strings")
	add(elf.SHF_INFO_LINK, "sh_info links to section")
	add(elf.SHF_LINK_ORDER, "Link order")
	add(elf.SHF_OS_NONCONFORMING, "OS-specific handling")
	add(elf.SHF_GROUP, "Section group member")
	add(elf.SHF_TLS, "Thread-local storage")
	add(elf.SHF_COMPRESSED, "Compressed")
	if x := f & elf.SHF_MASKOS; x != 0 {
		parts = append(parts, fmt.Sprintf("OS-specific field 0x%x", uint32(x)))
		known |= x
	}
	if x := f & elf.SHF_MASKPROC; x != 0 {
		parts = append(parts, fmt.Sprintf("CPU-specific field 0x%x", uint32(x)))
		known |= x
	}
	if rest := f &^ known; rest != 0 {
		parts = append(parts, fmt.Sprintf("0x%x", uint32(rest)))
	}
	return strings.Join(parts, ", ")
}

// readableProgType maps standard program header types to short descriptions.
func readableProgType(t elf.ProgType) string {
	switch t {
	case elf.PT_NULL:
		return "Unused"
	case elf.PT_LOAD:
		return "Loadable segment"
	case elf.PT_DYNAMIC:
		return "Dynamic linking"
	case elf.PT_INTERP:
		return "Interpreter path"
	case elf.PT_NOTE:
		return "Notes / auxiliary info"
	case elf.PT_SHLIB:
		return "Reserved (not used)"
	case elf.PT_PHDR:
		return "Program headers in file"
	case elf.PT_TLS:
		return "Thread-local storage template"
	case elf.PT_GNU_EH_FRAME:
		return "Exception unwind (.eh_frame_hdr)"
	case elf.PT_GNU_STACK:
		return "Stack permissions"
	case elf.PT_GNU_RELRO:
		return "Read-only after relocations"
	case elf.PT_GNU_PROPERTY:
		return "GNU properties"
	case elf.PT_PAX_FLAGS:
		return "PAX flags"
	case elf.PT_OPENBSD_RANDOMIZE:
		return "OpenBSD randomization data"
	case elf.PT_OPENBSD_WXNEEDED:
		return "OpenBSD W^X marker"
	case elf.PT_OPENBSD_BOOTDATA:
		return "OpenBSD boot data"
	default:
		return t.String()
	}
}

// readableProgFlags maps PF_* to r/w/x style.
func readableProgFlags(f elf.ProgFlag) string {
	if f == 0 {
		return "—"
	}
	var parts []string
	var known elf.ProgFlag
	add := func(mask elf.ProgFlag, label string) {
		if f&mask != 0 {
			parts = append(parts, label)
			known |= mask
		}
	}
	add(elf.PF_R, "Read")
	add(elf.PF_W, "Write")
	add(elf.PF_X, "Execute")
	if x := f & elf.PF_MASKOS; x != 0 {
		parts = append(parts, fmt.Sprintf("OS-specific field 0x%x", uint32(x)))
		known |= x
	}
	if x := f & elf.PF_MASKPROC; x != 0 {
		parts = append(parts, fmt.Sprintf("CPU-specific field 0x%x", uint32(x)))
		known |= x
	}
	if rest := f &^ known; rest != 0 {
		parts = append(parts, fmt.Sprintf("0x%x", uint32(rest)))
	}
	return strings.Join(parts, ", ")
}

func readableSymType(t elf.SymType) string {
	switch t {
	case elf.STT_NOTYPE:
		return "None"
	case elf.STT_OBJECT:
		return "Data object"
	case elf.STT_FUNC:
		return "Function"
	case elf.STT_SECTION:
		return "Section"
	case elf.STT_FILE:
		return "Source file"
	case elf.STT_COMMON:
		return "Common block"
	case elf.STT_TLS:
		return "TLS object"
	case elf.STT_GNU_IFUNC:
		return "Indirect function"
	case elf.STT_RELC, elf.STT_SRELC:
		return t.String()
	default:
		return t.String()
	}
}

func readableSymBind(b elf.SymBind) string {
	switch b {
	case elf.STB_LOCAL:
		return "Local"
	case elf.STB_GLOBAL:
		return "Global"
	case elf.STB_WEAK:
		return "Weak"
	default:
		return b.String()
	}
}

func readableSymVis(v elf.SymVis) string {
	switch v {
	case elf.STV_DEFAULT:
		return "Default"
	case elf.STV_INTERNAL:
		return "Internal"
	case elf.STV_HIDDEN:
		return "Hidden"
	case elf.STV_PROTECTED:
		return "Protected"
	default:
		return v.String()
	}
}

// readableSectionIndex turns SHN_* / numeric section refs into short text.
func readableSectionIndex(i elf.SectionIndex) string {
	switch i {
	case elf.SHN_UNDEF:
		return "Undefined"
	case elf.SHN_ABS:
		return "Absolute"
	case elf.SHN_COMMON:
		return "Common"
	case elf.SHN_XINDEX:
		return "Extended index"
	default:
		if i < 0xff00 {
			return fmt.Sprintf("Section #%d", int(i))
		}
		return i.String()
	}
}

// readableDynTag adds a gloss for common dynamic tags; falls back to elf name.
func readableDynTag(tag elf.DynTag) string {
	switch tag {
	case elf.DT_NULL:
		return "End of dynamic section"
	case elf.DT_NEEDED:
		return "Needed shared library"
	case elf.DT_PLTRELSZ:
		return "Total size of PLT relocations"
	case elf.DT_PLTGOT:
		return "Procedure linkage table / GOT"
	case elf.DT_HASH:
		return "Symbol hash table address"
	case elf.DT_STRTAB:
		return "Dynamic string table address"
	case elf.DT_SYMTAB:
		return "Dynamic symbol table address"
	case elf.DT_RELA:
		return "Rela relocations address"
	case elf.DT_RELASZ:
		return "Total Rela relocation size"
	case elf.DT_RELAENT:
		return "One Rela relocation entry size"
	case elf.DT_STRSZ:
		return "Dynamic string table size"
	case elf.DT_SYMENT:
		return "One dynamic symbol entry size"
	case elf.DT_INIT:
		return "Initialization function"
	case elf.DT_FINI:
		return "Finalization function"
	case elf.DT_SONAME:
		return "Shared object name"
	case elf.DT_RPATH:
		return "Library search path (deprecated)"
	case elf.DT_SYMBOLIC:
		return "Symbolic linking"
	case elf.DT_REL:
		return "Rel relocations address"
	case elf.DT_RELSZ:
		return "Total Rel relocation size"
	case elf.DT_RELENT:
		return "One Rel relocation entry size"
	case elf.DT_PLTREL:
		return "PLT relocation type"
	case elf.DT_DEBUG:
		return "Debug (historical)"
	case elf.DT_TEXTREL:
		return "Relocations may modify read-only text"
	case elf.DT_JMPREL:
		return "PLT relocations address"
	case elf.DT_BIND_NOW:
		return "Bind now"
	case elf.DT_INIT_ARRAY:
		return "Initialization functions array"
	case elf.DT_FINI_ARRAY:
		return "Finalization functions array"
	case elf.DT_INIT_ARRAYSZ:
		return "Initialization functions array size"
	case elf.DT_FINI_ARRAYSZ:
		return "Finalization functions array size"
	case elf.DT_RUNPATH:
		return "Library search path (DT_RUNPATH)"
	case elf.DT_FLAGS:
		return "Object-specific flags"
	case elf.DT_PREINIT_ARRAY:
		return "Pre-initialization functions"
	case elf.DT_PREINIT_ARRAYSZ:
		return "Pre-initialization array size"
	case elf.DT_GNU_HASH:
		return "GNU hash table address"
	default:
		return tag.String()
	}
}

// formatDynTagDisplay shows a short description plus the standard DT_* name when they differ.
func formatDynTagDisplay(tag elf.DynTag) string {
	raw := tag.String()
	human := readableDynTag(tag)
	if human == raw {
		return raw
	}
	return fmt.Sprintf("%s (%s)", human, raw)
}
