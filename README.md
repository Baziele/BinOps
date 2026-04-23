# BinOps

Terminal UI for quickly inspecting binaries—geared toward lightweight reverse engineering, triage, and sanity checks without leaving the shell.

**Status:** active development. The goal is first-class support on **Windows, macOS, and Linux**, with **PE**, **Mach-O**, and **ELF** binaries. Today the UI and analysis path focus on **ELF** on **Linux**.

![General page demo](assets/general.gif)

## Features

### Static

![Static page demo](assets/static.gif)

ELF header fields, **GNU** note sections (e.g. build ID), and program header / segment tables for a static picture of how the file is laid out.

### Dynamic

![Dynamic page demo](assets/dynamic.gif)

Reserved for dynamic linking and runtime-oriented inspection. **Coming soon.**

### Strings

![Strings page demo](assets/strings.gif)

Scrollable extracted strings from the mapped file, useful for quick triage of symbols, paths, and embedded text.

### Hexdump

![Hexdump page demo](assets/hexdump.gif)

Raw byte view of the full file with navigation for low-level inspection.

## Requirements

- **Go** [1.26](https://go.dev/dl/) or newer (see `go.mod`).

## Build & run

From the repository root:

```bash
go build -o binops .
./binops /path/to/binary
```

Or without installing a binary:

```bash
go run . /path/to/binary
```

There are no release binaries yet; use `go build` or `go run` as above.

## Controls

- **Tab** / **Shift+Tab** — next / previous page (after the file is loaded)
- **q** or **Ctrl+C** — quit

## Limitations (today)

- Parsing and UI are built around **ELF** (`debug/elf`). Other executable formats are **not** handled yet; they are on the roadmap above.
- Some file metadata helpers are OS-specific (full detail on **Linux**; stubs elsewhere until ported).

The GIFs above were produced with [VHS](https://github.com/charmbracelet/vhs); cassette files are in [`assets/tapes/`](assets/tapes/).

## License

MIT — see [LICENSE](LICENSE).
