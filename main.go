package main

import (
	"fmt"
	"log"
	"os"

	// "github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/xorwise/wpm/input"
	"golang.org/x/term"
)

type model struct {
	textInput input.Model
}

func initialModel() model {
	ti := input.New()
	words := input.StringToWords("banana apple car house sky tree book dog cat moon sun road river beach park tree chair table lamp bed pillow window door floor wall ceiling rug carpet sofa chair shelf curtain plant flower vase clock mirror towel soap shampoo brush toothpaste toothbrush razor scissors")
	ti.Text = words
	ti.Focus()
	ti.Width = 70
	ti.ValueStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#27D8C4"))
	ti.MistakeStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#D8273B")).Bold(true)
	ti.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("gray"))
	ti.SetFirstWord()

	return model{textInput: ti}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter, tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyTab:
			return m, nil
		}
	}
	return m, cmd
}

func (m model) View() string {
	physicalWidth, _, _ := term.GetSize(int(os.Stdout.Fd()))
	title := lipgloss.NewStyle().
		Width(physicalWidth).
		Align(lipgloss.Center).
		Render("MONKEY TYPE (NOT)")
	text := lipgloss.NewStyle().
		Width(physicalWidth).
		Align(lipgloss.Center).
		MarginBottom(10).
		Render(m.textInput.View())
	return lipgloss.JoinVertical(lipgloss.Top, title, text)

}

func main() {
	f, err := os.OpenFile("wpm.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	log.SetOutput(f)
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
