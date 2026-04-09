package main

import tea "charm.land/bubbletea/v2"

type StaticModel struct {
	content string
}

func initializeStaticModel() *StaticModel {
	return &StaticModel{content: "Static"}
}


func (m StaticModel) Init() tea.Cmd{
	return nil
}

func (m StaticModel) Update(msg tea.Msg) (tea.Model, tea.Cmd){
	return m, nil
}


func (m StaticModel) View() tea.View{
	return tea.NewView(m.content)
}
