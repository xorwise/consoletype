package main

import (
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/xorwise/wpm/input"
	"github.com/xorwise/wpm/text"
	"golang.org/x/term"
)

type model struct {
	textInput input.Model
	lang      string
}

func initialModel(lang string) model {
	ti := input.New()
	words := input.StringToWords(text.GenerateText(lang, 100))
	ti.Text = words
	ti.Focus()
	ti.Width = 70
	ti.ValueStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#27D8C4"))
	ti.MistakeStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#D8273B")).Bold(true)
	ti.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("gray"))
	ti.SetFirstWord()

	return model{textInput: ti, lang: lang}
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
			m = initialModel(m.lang)
			return m, nil
		}
	}
	return m, cmd
}

func (m model) View() string {
	physicalWidth, physicalHeight, _ := term.GetSize(int(os.Stdout.Fd()))
	title := lipgloss.NewStyle().
		Width(physicalWidth).
		Align(lipgloss.Center).
		Render("MONKEY TYPE (NOT)")
	text := lipgloss.NewStyle().
		Width(physicalWidth).
		Align(lipgloss.Center).
		MarginBottom(10).
		Render(m.textInput.View())

	return lipgloss.NewStyle().
		Width(physicalWidth).
		Height(physicalHeight).
		Render(
			lipgloss.JoinVertical(lipgloss.Top, title, text),
		)

}

func main() {
	var lang string
	if len(os.Args) < 2 {
		lang = "en"
	} else {
		lang = os.Args[1]
	}
	f, err := os.OpenFile("wpm.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	log.SetOutput(f)
	p := tea.NewProgram(initialModel(lang), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
