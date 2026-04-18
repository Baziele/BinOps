package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestHexDocumentReadAtUsesDirtyBytes(t *testing.T) {
	doc := newHexDocument([]byte{0x00, 0x11, 0x22, 0x33})
	if ok := doc.SetByteAt(1, 0xAA); !ok {
		t.Fatal("expected edit to succeed")
	}

	got, ok := doc.ReadAt(0, 4)
	if !ok {
		t.Fatal("expected read to succeed")
	}

	want := []byte{0x00, 0xAA, 0x22, 0x33}
	if !bytes.Equal(got, want) {
		t.Fatalf("unexpected read bytes: got %v want %v", got, want)
	}
	if doc.DirtyCount() != 1 {
		t.Fatalf("unexpected dirty count: got %d want 1", doc.DirtyCount())
	}
}

func TestHexDocumentSaveClearsDirtyState(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/sample.bin"
	original := []byte{0x10, 0x20, 0x30}
	if err := os.WriteFile(path, original, 0o644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	doc := newHexDocument(original)
	doc.SetByteAt(1, 0x99)
	if err := doc.Save(path); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read saved file: %v", err)
	}
	want := []byte{0x10, 0x99, 0x30}
	if !bytes.Equal(got, want) {
		t.Fatalf("unexpected saved bytes: got %v want %v", got, want)
	}
	if doc.DirtyCount() != 0 {
		t.Fatalf("expected dirty state to be cleared, got %d", doc.DirtyCount())
	}
}

func TestHexdumpModelApplyHexDigitEditsAndAdvances(t *testing.T) {
	model := initializeHexdumpModel(80, 20, "sample.bin", []byte{0x12, 0x34, 0x56})
	model.view.editMode = true
	model.view.highNibble = true

	model.applyHexDigit(0xA)
	got, _ := model.document.ByteAt(0)
	if got != 0xA2 {
		t.Fatalf("unexpected byte after high nibble edit: got 0x%X want 0xA2", got)
	}
	if model.view.highNibble {
		t.Fatal("expected to move to low nibble after first edit")
	}
	if model.view.cursorOffset != 0 {
		t.Fatalf("unexpected cursor move after first nibble: got %d want 0", model.view.cursorOffset)
	}

	model.applyHexDigit(0xB)
	got, _ = model.document.ByteAt(0)
	if got != 0xAB {
		t.Fatalf("unexpected byte after low nibble edit: got 0x%X want 0xAB", got)
	}
	if !model.view.highNibble {
		t.Fatal("expected next edit to start on high nibble")
	}
	if model.view.cursorOffset != 1 {
		t.Fatalf("unexpected cursor move after completed byte edit: got %d want 1", model.view.cursorOffset)
	}
}

func TestHexdumpModelRenderVisibleRowsUsesWindowedRows(t *testing.T) {
	model := initializeHexdumpModel(80, 20, "sample.bin", []byte{
		0x00, 0x01, 0x02, 0x03,
		0x04, 0x05, 0x06, 0x07,
		0x08, 0x09, 0x0A, 0x0B,
		0x0C, 0x0D, 0x0E, 0x0F,
	})
	model.view.bytesPerRow = 4
	model.view.visibleRows = 2
	model.view.topRow = 1
	model.layout.hexWidth = model.view.bytesPerRow*3 - 1
	model.layout.asciiWidth = model.view.bytesPerRow
	model.invalidateAllRows()

	addr, _, _ := model.renderVisibleRows()
	if strings.Contains(addr, "00000000") {
		t.Fatal("rendered rows included data above the visible window")
	}
	if !strings.Contains(addr, "00000004") || !strings.Contains(addr, "00000008") {
		t.Fatalf("rendered rows did not match visible window: %q", addr)
	}
}

func BenchmarkHexdumpVisibleWindowNavigation(b *testing.B) {
	sizes := []int{
		100 * 1024,
		1024 * 1024,
		10 * 1024 * 1024,
	}

	for _, size := range sizes {
		data := make([]byte, size)
		for i := range data {
			data[i] = byte(i)
		}

		b.Run(fmt.Sprintf("%dKB", size/1024), func(b *testing.B) {
			model := initializeHexdumpModel(120, 40, "sample.bin", data)
			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				if model.view.cursorOffset >= model.document.Len()-1 {
					model.view.cursorOffset = 0
					model.view.topRow = 0
					model.invalidateAllRows()
				}
				model.moveCursor(1)
				model.renderVisibleRows()
			}
		})
	}
}
