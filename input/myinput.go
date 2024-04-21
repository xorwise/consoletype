package input

import (
	"slices"
	"strings"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rivo/uniseg"
)

type ValidateFunc func(string) error

type KeyMap struct {
	DeleteCharacter key.Binding
}

type Word struct {
	Value []rune
	pos   int
}

type Model struct {
	Err          error
	Text         []*Word
	TextStyle    lipgloss.Style
	ValueStyle   lipgloss.Style
	MistakeStyle lipgloss.Style
	Width        int

	Validate ValidateFunc

	finishedText         []string
	focus                bool
	word                 Word
	currentWordIndex     int
	currentWordCorrect   []int
	currentWordIncorrect []int
}

func New() Model {
	return Model{
		Err:        nil,
		TextStyle:  lipgloss.NewStyle().Foreground(lipgloss.Color("red")),
		ValueStyle: lipgloss.NewStyle(),
		MistakeStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("red")).
			Bold(true),
		Width:                100,
		finishedText:         []string{},
		focus:                false,
		word:                 Word{Value: []rune{}, pos: 0},
		currentWordIndex:     0,
		currentWordCorrect:   []int{},
		currentWordIncorrect: []int{},
	}
}

func (m *Model) Focus() {
	m.focus = true
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m *Model) SetFirstWord() {
	m.word = Word{Value: make([]rune, len(m.Text[0].Value)), pos: 0}
	copy(m.word.Value, m.Text[0].Value)
}

func (m *Model) styleWord(w Word, mistakes []int) string {
	return lipgloss.StyleRunes(string(w.Value), mistakes, m.MistakeStyle.Inline(true), m.ValueStyle.Inline(true))
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeySpace:
			if m.word.pos < len(m.word.Value) {
				for i := m.word.pos; i < len(m.word.Value); i++ {
					m.currentWordIncorrect = append(m.currentWordIncorrect, i)
				}
			}
			m.finishedText = append(m.finishedText, m.styleWord(m.word, m.currentWordIncorrect))
			m.currentWordIndex++
			m.word = Word{Value: make([]rune, len(m.Text[m.currentWordIndex].Value)), pos: 0}
			copy(m.word.Value, m.Text[m.currentWordIndex].Value)
			m.currentWordIncorrect = []int{}
			m.currentWordCorrect = []int{}
		case tea.KeyRunes:
			if m.word.pos >= len(m.word.Value) {
				m.word.Value = append(m.word.Value, msg.Runes...)
				m.currentWordIncorrect = append(m.currentWordIncorrect, m.word.pos)
			} else if m.Text[m.currentWordIndex].Value[m.word.pos] != msg.Runes[0] {
				m.currentWordIncorrect = append(m.currentWordIncorrect, m.word.pos)
			} else {
				m.currentWordCorrect = append(m.currentWordCorrect, m.word.pos)
			}
			m.word.Value[m.word.pos] = msg.Runes[0]
			m.word.pos++
		case tea.KeyTab:
			return m, tea.Quit
		case tea.KeyBackspace:
			if m.word.pos > 0 {
				isFinishedWord := m.word.pos == len(m.word.Value) && len(m.currentWordIncorrect) == 0
				if !isFinishedWord {
					if m.word.pos > len(m.Text[m.currentWordIndex].Value) {
						m.word.Value = m.word.Value[:m.word.pos-1]
					}
					m.word.pos--
					if k := slices.Index(m.currentWordCorrect, m.word.pos); k >= 0 {
						if k+1 < len(m.currentWordCorrect) {
							m.currentWordCorrect = append(m.currentWordCorrect[:k], m.currentWordCorrect[k+1:]...)
						} else {
							m.currentWordCorrect = m.currentWordCorrect[:k]
						}
					}
					if k := slices.Index(m.currentWordIncorrect, m.word.pos); k >= 0 {
						if k+1 < len(m.currentWordIncorrect) {
							m.currentWordIncorrect = append(m.currentWordIncorrect[:k], m.currentWordIncorrect[k+1:]...)
						} else {
							m.currentWordIncorrect = m.currentWordIncorrect[:k]
						}
					}
				}
			}
		}

	}
	return m, nil
}

func StringToWords(s string) []*Word {
	var words []*Word
	for _, w := range strings.Split(s, " ") {
		words = append(words, &Word{Value: []rune(w), pos: 0})
	}
	return words
}

func (m Model) View() string {
	styleText := m.TextStyle.Inline(true).Render

	v := ""
	for i, word := range m.Text {
		if i < m.currentWordIndex {
			v += m.finishedText[i]
			v += styleText(" ")
		} else if i == m.currentWordIndex {
			currentWordValue := lipgloss.StyleRunes(string(m.word.Value[:m.word.pos]), m.currentWordIncorrect, m.MistakeStyle.Inline(true), m.ValueStyle.Inline(true))
			if m.word.pos < len(word.Value) {
				currentWordValue += m.TextStyle.Inline(true).Background(lipgloss.Color("#6C27D8")).Render(string(word.Value[m.word.pos]))
			}
			if m.word.pos+1 < len(word.Value) {
				currentWordValue += styleText(string(word.Value[m.word.pos+1:]))
			}
			v += currentWordValue
			v += styleText(" ")
		} else {
			v += styleText(string(word.Value))
			v += styleText(" ")
		}
	}
	styled := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), true).
		Render(m.handleOverflow(v))
	return styled
}

func (m *Model) handleOverflow(s string) string {
	base := ""
	indexes := []int{}
	for i, w := range m.Text {
		if uniseg.StringWidth(base) > m.Width {
			base = ""
			indexes = append(indexes, i-1)
		}
		base += string(w.Value)
	}
	words := strings.Split(s, " ")
	rows := []string{""}

	j := 0
	for i, w := range words {
		if j < len(indexes) && i == indexes[j] {
			rows = append(rows, "")
			j++
		}
		rows[len(rows)-1] += w + m.TextStyle.Render(" ")
	}

	result := ""
	for i := 0; i < len(rows); i++ {
		result += rows[i] + m.TextStyle.Render("\n")
	}
	return result
}

func Blink() tea.Msg {
	return cursor.Blink()
}
