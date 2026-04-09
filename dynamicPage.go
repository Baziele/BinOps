package main

import tea "charm.land/bubbletea/v2"

type DynamicModel struct {
	content string
}

func initializeDynamicModel() *DynamicModel {
	return &DynamicModel{content: "Dynamic"}
}


func (m DynamicModel) Init() tea.Cmd{
	return nil
}

func (m DynamicModel) Update(msg tea.Msg) (tea.Model, tea.Cmd){
	return m, nil
}


func (m DynamicModel) View() tea.View{
	return tea.NewView(m.content)
}
