package main

import (
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type GeneralModel struct {
	width int
	height int
	applicationName string
	description string 
	repository string
	creator string 
	filename string
	propertiesViewport viewport.Model
	dependenciesViewport viewport.Model
	content string
}

func initializeGeneralModel(width, height int) GeneralModel {
	m := GeneralModel{
		width: width,
		height: height,
		content: "General",
		applicationName: "BinOps",
		description: "The greatest static binary analysis too of all time",
		repository: "https://github.com/baziele/binops",
		creator: "@baziele",
		filename: "maliciousFile.exe",
		propertiesViewport: viewport.New(viewport.WithWidth(40), viewport.WithHeight(15)),
		dependenciesViewport: viewport.New(viewport.WithWidth(60), viewport.WithHeight(7)),
	}

	return m
}


func (m GeneralModel) Init() tea.Cmd{
	return nil
}

func (m GeneralModel) Update(msg tea.Msg) (GeneralModel, tea.Cmd){
	var cmd tea.Cmd
	var cmds []tea.Cmd
	m.propertiesViewport, cmd = m.propertiesViewport.Update(msg)
	cmds = append(cmds, cmd)
	m.dependenciesViewport, cmd = m.dependenciesViewport.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}


func (m GeneralModel) View() string{
	
	applicationName := lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4")).Render(m.applicationName)
	description := lipgloss.NewStyle().Underline(true).Render(m.description) //BorderBottom(true)
	repository := lipgloss.NewStyle().Hyperlink(m.repository).Render(m.repository)
	creator := lipgloss.NewStyle().Italic(true).Foreground(lipgloss.Color("#7D56F4")).Render(m.creator)
	filename := lipgloss.NewStyle().Render(m.filename)
	propertiesViewport := lipgloss.NewStyle().Border(lipgloss.NormalBorder()).Render(m.propertiesViewport.View())
	dependenciesViewport := lipgloss.NewStyle().Border(lipgloss.NormalBorder()).Render(m.dependenciesViewport.View())	

	v := lipgloss.JoinVertical(lipgloss.Center, applicationName, description, repository, creator, filename, propertiesViewport, dependenciesViewport)
	v = lipgloss.PlaceHorizontal(m.width , lipgloss.Center, v)
	return v
}

func (m *GeneralModel) setDimensions(width, height int){
	m.width = width
	m.height = height
	m.propertiesViewport.SetWidth(min(40, width-2))
	m.propertiesViewport.SetHeight(min(15, height-2))
	m.dependenciesViewport.SetWidth(min(60, width-2))
	m.dependenciesViewport.SetHeight(min(7, height-2))
}
