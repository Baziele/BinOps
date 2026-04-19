package main

import (
	"strings"

	"charm.land/lipgloss/v2"
)

func renderVerticalScrollbar(height, total, visible, offset int, trackStyle, thumbStyle lipgloss.Style) string {
	if height <= 0 {
		return ""
	}

	visible = max(1, visible)
	total = max(visible, total)
	offset = clamp(offset, 0, max(0, total-visible))

	thumbHeight := height
	thumbTop := 0
	if total > visible {
		thumbHeight = max(1, (visible*height)/total)
		thumbHeight = min(height, thumbHeight)
		thumbTravel := max(0, height-thumbHeight)
		maxOffset := max(1, total-visible)
		thumbTop = (offset * thumbTravel) / maxOffset
	}

	var b strings.Builder
	for i := 0; i < height; i++ {
		if i > 0 {
			b.WriteByte('\n')
		}
		if i >= thumbTop && i < thumbTop+thumbHeight {
			b.WriteString(thumbStyle.Render("█"))
			continue
		}
		b.WriteString(trackStyle.Render("│"))
	}

	return b.String()
}
