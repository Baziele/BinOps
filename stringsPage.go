package main

import tea "charm.land/bubbletea/v2"

type StringsModel struct {
	content string
}

func initializeStringsModel() *StringsModel {
	return &StringsModel{content: "Strings"}
}


func (m StringsModel) Init() tea.Cmd{
	return nil
}

func (m StringsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd){
	return m, nil
}


func (m StringsModel) View() tea.View{
	return tea.NewView(m.content)
}
