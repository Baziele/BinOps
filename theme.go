package main

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

type ThemeName string

const (
	ThemeNebula         ThemeName = "nebula"
	ThemeSunsetTerminal ThemeName = "sunset-terminal"
	ThemeForestNight    ThemeName = "forest-night"
	ThemeArcticGlow     ThemeName = "arctic-glow"
	ThemeSynthwaveMuted ThemeName = "synthwave-muted"
	ThemeMonoPlusAccent ThemeName = "mono-plus-accent"
	ThemePrismDrive     ThemeName = "prism-drive"
	ThemeHarborLights   ThemeName = "harbor-lights"
	ThemeAmberCircuit   ThemeName = "amber-circuit"
)

// ActiveTheme controls the global CLI theme.
// Switch this constant to any ThemeName preset.
const ActiveTheme = ThemeForestNight

type HexdumpTheme struct {
	Null        color.Color
	Space       color.Color
	High        color.Color
	Control     color.Color
	Printable   color.Color
	Standard    color.Color
	SelectedBG  color.Color
	SelectedFG  color.Color
}

type Theme struct {
	Name ThemeName

	NavAccent      color.Color
	Border         color.Color
	PanelTitle     color.Color
	Label          color.Color
	Description    color.Color
	Creator        color.Color
	SelectionBG    color.Color
	SelectionFG    color.Color
	Muted          color.Color
	MutedStrong    color.Color
	Subtle         color.Color
	Body           color.Color
	BodyAlt        color.Color
	Warning        color.Color

	StringsTitleFG  color.Color
	StringsTitleBG  color.Color
	StringsFooterFG color.Color
	StringsFooterBG color.Color
	StringsOffset   color.Color

	Hexdump HexdumpTheme
}

func GetTheme(name ThemeName) Theme {
	switch name {
	case ThemeSunsetTerminal:
		return themeSunsetTerminal()
	case ThemeForestNight:
		return themeForestNight()
	case ThemeArcticGlow:
		return themeArcticGlow()
	case ThemeSynthwaveMuted:
		return themeSynthwaveMuted()
	case ThemeMonoPlusAccent:
		return themeMonoPlusAccent()
	case ThemePrismDrive:
		return themePrismDrive()
	case ThemeHarborLights:
		return themeHarborLights()
	case ThemeAmberCircuit:
		return themeAmberCircuit()
	case ThemeNebula:
		fallthrough
	default:
		return themeNebula()
	}
}

func themeNebula() Theme {
	return Theme{
		Name:           ThemeNebula,
		NavAccent:      lipgloss.Color("#8D6BF5"),
		Border:         lipgloss.Color("#4A5672"),
		PanelTitle:     lipgloss.Color("#5FD7FF"),
		Label:          lipgloss.Color("#6BCBFF"),
		Description:    lipgloss.Color("#E08BFF"),
		Creator:        lipgloss.Color("#B89CFF"),
		SelectionBG:    lipgloss.Color("#4B5263"),
		SelectionFG:    lipgloss.Color("#E6EAF2"),
		Muted:          lipgloss.Color("#95A0B8"),
		MutedStrong:    lipgloss.Color("#AAB3C7"),
		Subtle:         lipgloss.Color("#7A859D"),
		Body:           lipgloss.Color("#D0D7E5"),
		BodyAlt:        lipgloss.Color("#BFC8DB"),
		Warning:        lipgloss.Color("#E5B07A"),
		StringsTitleFG: lipgloss.Color("#F6F0FF"),
		StringsTitleBG: lipgloss.Color("#6D56D8"),
		StringsFooterFG: lipgloss.Color("#FFF2C2"),
		StringsFooterBG: lipgloss.Color("#6D56D8"),
		StringsOffset:   lipgloss.Color("#8BC3FF"),
		Hexdump: HexdumpTheme{
			Null:       lipgloss.Color("#5C6370"),
			Space:      lipgloss.Color("#7A8FA3"),
			High:       lipgloss.Color("#9A84B8"),
			Control:    lipgloss.Color("#B08B68"),
			Printable:  lipgloss.Color("#7FAE7B"),
			Standard:   lipgloss.Color("#C8CCD4"),
			SelectedBG: lipgloss.Color("#4B5263"),
			SelectedFG: lipgloss.Color("#E6EAF2"),
		},
	}
}

func themeSunsetTerminal() Theme {
	return Theme{
		Name:            ThemeSunsetTerminal,
		NavAccent:       lipgloss.Color("#F38E70"),
		Border:          lipgloss.Color("#6B4E49"),
		PanelTitle:      lipgloss.Color("#FFB278"),
		Label:           lipgloss.Color("#FFD08A"),
		Description:     lipgloss.Color("#FF92A5"),
		Creator:         lipgloss.Color("#FFC0A3"),
		SelectionBG:     lipgloss.Color("#5B403A"),
		SelectionFG:     lipgloss.Color("#FFF0DF"),
		Muted:           lipgloss.Color("#C5AA9C"),
		MutedStrong:     lipgloss.Color("#D5B8A8"),
		Subtle:          lipgloss.Color("#A5887A"),
		Body:            lipgloss.Color("#F4E2D7"),
		BodyAlt:         lipgloss.Color("#E5D1C5"),
		Warning:         lipgloss.Color("#F4C27F"),
		StringsTitleFG:  lipgloss.Color("#FFF4E6"),
		StringsTitleBG:  lipgloss.Color("#B35D4E"),
		StringsFooterFG: lipgloss.Color("#FFF0C9"),
		StringsFooterBG: lipgloss.Color("#B35D4E"),
		StringsOffset:   lipgloss.Color("#FFC284"),
		Hexdump: HexdumpTheme{
			Null:       lipgloss.Color("#8C6A5A"),
			Space:      lipgloss.Color("#A78472"),
			High:       lipgloss.Color("#C47B92"),
			Control:    lipgloss.Color("#D6A46D"),
			Printable:  lipgloss.Color("#89B07D"),
			Standard:   lipgloss.Color("#EBD8CB"),
			SelectedBG: lipgloss.Color("#5B403A"),
			SelectedFG: lipgloss.Color("#FFF0DF"),
		},
	}
}

func themeForestNight() Theme {
	return Theme{
		Name:            ThemeForestNight,
		NavAccent:       lipgloss.Color("#4FC18D"),
		Border:          lipgloss.Color("#3B5D54"),
		PanelTitle:      lipgloss.Color("#73D3B0"),
		Label:           lipgloss.Color("#87E8C4"),
		Description:     lipgloss.Color("#A5DFAF"),
		Creator:         lipgloss.Color("#7BC79D"),
		SelectionBG:     lipgloss.Color("#304940"),
		SelectionFG:     lipgloss.Color("#EAF8F0"),
		Muted:           lipgloss.Color("#9DB9AB"),
		MutedStrong:     lipgloss.Color("#B2CDBF"),
		Subtle:          lipgloss.Color("#7E9D90"),
		Body:            lipgloss.Color("#D1E7DC"),
		BodyAlt:         lipgloss.Color("#C0D8CC"),
		Warning:         lipgloss.Color("#CFA56A"),
		StringsTitleFG:  lipgloss.Color("#ECFFF7"),
		StringsTitleBG:  lipgloss.Color("#2F7C63"),
		StringsFooterFG: lipgloss.Color("#F6F3C9"),
		StringsFooterBG: lipgloss.Color("#2F7C63"),
		StringsOffset:   lipgloss.Color("#9AE7D0"),
		Hexdump: HexdumpTheme{
			Null:       lipgloss.Color("#6A8077"),
			Space:      lipgloss.Color("#82A396"),
			High:       lipgloss.Color("#90BFA9"),
			Control:    lipgloss.Color("#B5976A"),
			Printable:  lipgloss.Color("#7AC28F"),
			Standard:   lipgloss.Color("#D2E4DA"),
			SelectedBG: lipgloss.Color("#304940"),
			SelectedFG: lipgloss.Color("#EAF8F0"),
		},
	}
}

func themeArcticGlow() Theme {
	return Theme{
		Name:            ThemeArcticGlow,
		NavAccent:       lipgloss.Color("#79B7FF"),
		Border:          lipgloss.Color("#4B617A"),
		PanelTitle:      lipgloss.Color("#8CD8FF"),
		Label:           lipgloss.Color("#9FDFFF"),
		Description:     lipgloss.Color("#9AB7FF"),
		Creator:         lipgloss.Color("#A8CCFF"),
		SelectionBG:     lipgloss.Color("#3C4D61"),
		SelectionFG:     lipgloss.Color("#ECF6FF"),
		Muted:           lipgloss.Color("#A7B8CA"),
		MutedStrong:     lipgloss.Color("#BBC9D9"),
		Subtle:          lipgloss.Color("#8B9EB2"),
		Body:            lipgloss.Color("#D8E6F3"),
		BodyAlt:         lipgloss.Color("#C8D8E8"),
		Warning:         lipgloss.Color("#D8B17A"),
		StringsTitleFG:  lipgloss.Color("#F0F7FF"),
		StringsTitleBG:  lipgloss.Color("#4D6FA3"),
		StringsFooterFG: lipgloss.Color("#F5F1CF"),
		StringsFooterBG: lipgloss.Color("#4D6FA3"),
		StringsOffset:   lipgloss.Color("#9BD9FF"),
		Hexdump: HexdumpTheme{
			Null:       lipgloss.Color("#708399"),
			Space:      lipgloss.Color("#8EA3B9"),
			High:       lipgloss.Color("#9EA9D8"),
			Control:    lipgloss.Color("#C7A57B"),
			Printable:  lipgloss.Color("#86C1A9"),
			Standard:   lipgloss.Color("#D5DFEA"),
			SelectedBG: lipgloss.Color("#3C4D61"),
			SelectedFG: lipgloss.Color("#ECF6FF"),
		},
	}
}

func themeSynthwaveMuted() Theme {
	return Theme{
		Name:            ThemeSynthwaveMuted,
		NavAccent:       lipgloss.Color("#C57BFF"),
		Border:          lipgloss.Color("#5A4C77"),
		PanelTitle:      lipgloss.Color("#E07BFF"),
		Label:           lipgloss.Color("#7CDFFF"),
		Description:     lipgloss.Color("#FF8DC9"),
		Creator:         lipgloss.Color("#D7A0FF"),
		SelectionBG:     lipgloss.Color("#493A64"),
		SelectionFG:     lipgloss.Color("#F4ECFF"),
		Muted:           lipgloss.Color("#ADA1C2"),
		MutedStrong:     lipgloss.Color("#C0B4D6"),
		Subtle:          lipgloss.Color("#8D82A5"),
		Body:            lipgloss.Color("#DCCFEA"),
		BodyAlt:         lipgloss.Color("#CCBFE0"),
		Warning:         lipgloss.Color("#E5B07A"),
		StringsTitleFG:  lipgloss.Color("#F9F1FF"),
		StringsTitleBG:  lipgloss.Color("#7A4CA5"),
		StringsFooterFG: lipgloss.Color("#FFF0C2"),
		StringsFooterBG: lipgloss.Color("#7A4CA5"),
		StringsOffset:   lipgloss.Color("#91D4FF"),
		Hexdump: HexdumpTheme{
			Null:       lipgloss.Color("#7B6C94"),
			Space:      lipgloss.Color("#9A8AB4"),
			High:       lipgloss.Color("#C58AD8"),
			Control:    lipgloss.Color("#D1A06F"),
			Printable:  lipgloss.Color("#8CC798"),
			Standard:   lipgloss.Color("#D8CAE8"),
			SelectedBG: lipgloss.Color("#493A64"),
			SelectedFG: lipgloss.Color("#F4ECFF"),
		},
	}
}

func themeMonoPlusAccent() Theme {
	return Theme{
		Name:            ThemeMonoPlusAccent,
		NavAccent:       lipgloss.Color("#8AA4FF"),
		Border:          lipgloss.Color("#5A5F6D"),
		PanelTitle:      lipgloss.Color("#B2BEDE"),
		Label:           lipgloss.Color("#9FB1DF"),
		Description:     lipgloss.Color("#C7B6E2"),
		Creator:         lipgloss.Color("#A4AED0"),
		SelectionBG:     lipgloss.Color("#474C59"),
		SelectionFG:     lipgloss.Color("#F0F2F6"),
		Muted:           lipgloss.Color("#A1A6B3"),
		MutedStrong:     lipgloss.Color("#BCC1CE"),
		Subtle:          lipgloss.Color("#848A98"),
		Body:            lipgloss.Color("#D4D8E1"),
		BodyAlt:         lipgloss.Color("#C2C7D2"),
		Warning:         lipgloss.Color("#C5A372"),
		StringsTitleFG:  lipgloss.Color("#F2F4F9"),
		StringsTitleBG:  lipgloss.Color("#596289"),
		StringsFooterFG: lipgloss.Color("#ECE8C7"),
		StringsFooterBG: lipgloss.Color("#596289"),
		StringsOffset:   lipgloss.Color("#A5B8EA"),
		Hexdump: HexdumpTheme{
			Null:       lipgloss.Color("#7A808D"),
			Space:      lipgloss.Color("#959CAA"),
			High:       lipgloss.Color("#A69CC3"),
			Control:    lipgloss.Color("#B89C76"),
			Printable:  lipgloss.Color("#8EB29A"),
			Standard:   lipgloss.Color("#CBD1DC"),
			SelectedBG: lipgloss.Color("#474C59"),
			SelectedFG: lipgloss.Color("#F0F2F6"),
		},
	}
}

func themePrismDrive() Theme {
	return Theme{
		Name:            ThemePrismDrive,
		NavAccent:       lipgloss.Color("#00D1FF"),
		Border:          lipgloss.Color("#7A5CFF"),
		PanelTitle:      lipgloss.Color("#FF73C6"),
		Label:           lipgloss.Color("#7EF2FF"),
		Description:     lipgloss.Color("#FFA3D8"),
		Creator:         lipgloss.Color("#B6A0FF"),
		SelectionBG:     lipgloss.Color("#5A3F8F"),
		SelectionFG:     lipgloss.Color("#FFF3FF"),
		Muted:           lipgloss.Color("#A29CC3"),
		MutedStrong:     lipgloss.Color("#B9B2D6"),
		Subtle:          lipgloss.Color("#847BA9"),
		Body:            lipgloss.Color("#E2DBF6"),
		BodyAlt:         lipgloss.Color("#D1C9EA"),
		Warning:         lipgloss.Color("#FFC47A"),
		StringsTitleFG:  lipgloss.Color("#FFF4FF"),
		StringsTitleBG:  lipgloss.Color("#6A4BC0"),
		StringsFooterFG: lipgloss.Color("#FFF2C9"),
		StringsFooterBG: lipgloss.Color("#0E8FB4"),
		StringsOffset:   lipgloss.Color("#7DEBFF"),
		Hexdump: HexdumpTheme{
			Null:       lipgloss.Color("#7C6D9B"),
			Space:      lipgloss.Color("#58A9C2"),
			High:       lipgloss.Color("#C18BE6"),
			Control:    lipgloss.Color("#E6A15D"),
			Printable:  lipgloss.Color("#69CF9A"),
			Standard:   lipgloss.Color("#D8D0EC"),
			SelectedBG: lipgloss.Color("#5A3F8F"),
			SelectedFG: lipgloss.Color("#FFF3FF"),
		},
	}
}

func themeHarborLights() Theme {
	return Theme{
		Name:            ThemeHarborLights,
		NavAccent:       lipgloss.Color("#4DD0E1"),
		Border:          lipgloss.Color("#2E6F95"),
		PanelTitle:      lipgloss.Color("#FFA94D"),
		Label:           lipgloss.Color("#7CE7E7"),
		Description:     lipgloss.Color("#FF8A9E"),
		Creator:         lipgloss.Color("#9FD0FF"),
		SelectionBG:     lipgloss.Color("#2F4E73"),
		SelectionFG:     lipgloss.Color("#EFF8FF"),
		Muted:           lipgloss.Color("#93AFC5"),
		MutedStrong:     lipgloss.Color("#A8C1D5"),
		Subtle:          lipgloss.Color("#7893AA"),
		Body:            lipgloss.Color("#D8E7F3"),
		BodyAlt:         lipgloss.Color("#C7D9E8"),
		Warning:         lipgloss.Color("#F3BE77"),
		StringsTitleFG:  lipgloss.Color("#FFF7EC"),
		StringsTitleBG:  lipgloss.Color("#2D84A6"),
		StringsFooterFG: lipgloss.Color("#FFF1CE"),
		StringsFooterBG: lipgloss.Color("#955F2A"),
		StringsOffset:   lipgloss.Color("#98EEFF"),
		Hexdump: HexdumpTheme{
			Null:       lipgloss.Color("#647F96"),
			Space:      lipgloss.Color("#6E9FB2"),
			High:       lipgloss.Color("#A98FDD"),
			Control:    lipgloss.Color("#D9A369"),
			Printable:  lipgloss.Color("#72C89B"),
			Standard:   lipgloss.Color("#D0DEE9"),
			SelectedBG: lipgloss.Color("#2F4E73"),
			SelectedFG: lipgloss.Color("#EFF8FF"),
		},
	}
}

func themeAmberCircuit() Theme {
	return Theme{
		Name:            ThemeAmberCircuit,
		NavAccent:       lipgloss.Color("#FFC857"),
		Border:          lipgloss.Color("#5E7CE2"),
		PanelTitle:      lipgloss.Color("#8BE9FD"),
		Label:           lipgloss.Color("#FFD97D"),
		Description:     lipgloss.Color("#FF9E80"),
		Creator:         lipgloss.Color("#A9B9FF"),
		SelectionBG:     lipgloss.Color("#3B4F9A"),
		SelectionFG:     lipgloss.Color("#FFF8E8"),
		Muted:           lipgloss.Color("#A8AFC2"),
		MutedStrong:     lipgloss.Color("#BEC4D5"),
		Subtle:          lipgloss.Color("#8A91A5"),
		Body:            lipgloss.Color("#E1E5F2"),
		BodyAlt:         lipgloss.Color("#D0D7E8"),
		Warning:         lipgloss.Color("#FFBE55"),
		StringsTitleFG:  lipgloss.Color("#FFF6E0"),
		StringsTitleBG:  lipgloss.Color("#4E67C8"),
		StringsFooterFG: lipgloss.Color("#FFF0C0"),
		StringsFooterBG: lipgloss.Color("#A86C1A"),
		StringsOffset:   lipgloss.Color("#FFE08A"),
		Hexdump: HexdumpTheme{
			Null:       lipgloss.Color("#7E8496"),
			Space:      lipgloss.Color("#A3A7B7"),
			High:       lipgloss.Color("#8DA2F0"),
			Control:    lipgloss.Color("#E1A15A"),
			Printable:  lipgloss.Color("#83C88C"),
			Standard:   lipgloss.Color("#D5D9E6"),
			SelectedBG: lipgloss.Color("#3B4F9A"),
			SelectedFG: lipgloss.Color("#FFF8E8"),
		},
	}
}
