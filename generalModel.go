package main

import tea "charm.land/bubbletea/v2"

type GeneralModel struct {
	content string
}

func initializeGeneralModel() *GeneralModel {
	return &GeneralModel{content: "General"}
}


func (m GeneralModel) Init() tea.Cmd{
	return nil
}

func (m GeneralModel) Update(msg tea.Msg) (tea.Model, tea.Cmd){
	return m, nil
}


func (m GeneralModel) View() tea.View{
	return tea.NewView(m.content)
}
