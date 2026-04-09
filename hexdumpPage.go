package main

import tea "charm.land/bubbletea/v2"

type HexdumpModel struct {
	content string
}

func initializeHexdumpModel() *HexdumpModel {
	return &HexdumpModel{content: "Hexdump"}
}


func (m HexdumpModel) Init() tea.Cmd{
	return nil
}

func (m HexdumpModel) Update(msg tea.Msg) (tea.Model, tea.Cmd){
	return m, nil
}


func (m HexdumpModel) View() tea.View{
	return tea.NewView(m.content)
}
