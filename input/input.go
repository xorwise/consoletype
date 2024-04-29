package input

import (
	"fmt"
	"log"
	"slices"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/stopwatch"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rivo/uniseg"
)

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

	finishedText         []string
	focus                bool
	word                 Word
	currentWordIndex     int
	currentWordCorrect   []int
	currentWordIncorrect []int
	indexes              []int
	wpm                  int
	finishedWords        int
	stopwatch            stopwatch.Model
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
		wpm:                  0,
		finishedWords:        0,
		stopwatch:            stopwatch.NewWithInterval(500 * time.Millisecond),
	}
}

func (m *Model) Focus() {
	m.focus = true
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m *Model) calcIndexes() {
	base := ""
	for i, w := range m.Text {
		if uniseg.StringWidth(base) > m.Width {
			base = ""
			m.indexes = append(m.indexes, i-1)
		}
		base += string(w.Value) + " "
	}
}

func (m *Model) SetFirstWord() {
	m.word = Word{Value: make([]rune, len(m.Text[0].Value)), pos: 0}
	copy(m.word.Value, m.Text[0].Value)
	m.calcIndexes()
}

func (m *Model) styleWord(w Word, mistakes []int) string {
	return lipgloss.StyleRunes(string(w.Value), mistakes, m.MistakeStyle.Inline(true), m.ValueStyle.Inline(true))
}

func (m *Model) handleSpace() {
	isFinishedWord := m.word.pos == len(m.word.Value) && len(m.currentWordIncorrect) == 0
	if isFinishedWord {
		log.Println("finished word")
		m.finishedWords++
	}
	if m.word.pos == 0 {
		return
	}
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

}

func (m *Model) handleRunes(msg tea.KeyMsg) {
	if m.word.pos >= len(m.Text[m.currentWordIndex].Value) {
		if m.word.pos <= len(m.Text[m.currentWordIndex].Value)+1 {
			return
		}
		m.word.Value = append(m.word.Value, msg.Runes...)
		m.currentWordIncorrect = append(m.currentWordIncorrect, m.word.pos)
	} else if m.Text[m.currentWordIndex].Value[m.word.pos] != msg.Runes[0] {
		m.currentWordIncorrect = append(m.currentWordIncorrect, m.word.pos)
	} else {
		m.currentWordCorrect = append(m.currentWordCorrect, m.word.pos)
	}
	m.word.Value[m.word.pos] = msg.Runes[0]
	m.word.pos++
}

func (m *Model) handleBackspace() {
	isFinishedWord := m.word.pos == len(m.word.Value) && len(m.currentWordIncorrect) == 0
	if isFinishedWord || m.word.pos == 0 {
		return
	}
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

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if m.stopwatch.Running() && m.stopwatch.Elapsed().Minutes() > 0 {
		m.wpm = int(float64(m.finishedWords) / m.stopwatch.Elapsed().Minutes())
	}
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeySpace:
			m.handleSpace()
		case tea.KeyRunes:
			m.handleRunes(msg)
			if m.currentWordIndex == 0 && m.word.pos == 1 {
				log.Println("started stopwatch")
				return m, m.stopwatch.Start()
			}
		case tea.KeyBackspace:
			m.handleBackspace()
		}

	}
	var cmd tea.Cmd
	m.stopwatch, cmd = m.stopwatch.Update(msg)
	return m, cmd
}

func StringToWords(s string) []*Word {
	var words []*Word
	for _, w := range strings.Split(s, " ") {
		words = append(words, &Word{Value: []rune(w), pos: 0})
	}
	return words
}

func (m Model) View() string {
	v := ""
	v += m.viewWritten()

	v += m.viewCurrentWord()
	currentWordPos := len(v)
	v += m.viewRemaining()

	styled := lipgloss.JoinVertical(
		lipgloss.Top,
		lipgloss.NewStyle().
			Render(m.handleOverflow(v, currentWordPos)),
		lipgloss.NewStyle().
			Render(fmt.Sprintf("WPM: %d", m.wpm)),
	)
	return styled
}

func (m *Model) viewWritten() string {
	s := ""
	for i := 0; i < len(m.finishedText); i++ {
		s += m.TextStyle.Inline(true).Render(m.finishedText[i])
		s += m.TextStyle.Inline(true).Render(" ")
	}
	return s
}

func (m *Model) viewCurrentWord() string {
	s := lipgloss.StyleRunes(string(m.word.Value[:m.word.pos]), m.currentWordIncorrect, m.MistakeStyle.Inline(true), m.ValueStyle.Inline(true))

	if m.word.pos <= len(m.Text[m.currentWordIndex].Value) {
		if m.word.pos == len(m.Text[m.currentWordIndex].Value) {
			if !slices.Contains(m.indexes, m.currentWordIndex) {
				s += m.TextStyle.Inline(true).Background(lipgloss.Color("#6C27D8")).Render(" ")
			} else {
				s += m.TextStyle.Inline(true).Render(" ")
			}
		} else {
			s += m.TextStyle.Inline(true).Background(lipgloss.Color("#6C27D8")).Render(string(m.Text[m.currentWordIndex].Value[m.word.pos]))
			if m.word.pos+1 == len(m.Text[m.currentWordIndex].Value) {
				s += m.TextStyle.Inline(true).Render(" ")
			}
		}
	}
	if m.word.pos+1 < len(m.Text[m.currentWordIndex].Value) {
		s += m.TextStyle.Inline(true).Render(string(m.Text[m.currentWordIndex].Value[m.word.pos+1:]))
		s += m.TextStyle.Inline(true).Render(" ")
	}
	return s
}

func (m *Model) viewRemaining() string {
	s := ""
	for i := m.currentWordIndex + 1; i < len(m.Text); i++ {
		s += m.TextStyle.Inline(true).Render(string(m.Text[i].Value))
		s += m.TextStyle.Inline(true).Render(" ")
	}
	return s
}

func (m *Model) handleOverflow(s string, pos int) string {
	words := strings.Split(s, " ")
	rows := []string{""}

	j := 0
	for i, w := range words {
		if j < len(m.indexes) && i == m.indexes[j]+1 {
			rows = append(rows, "")
			j++
		}
		rows[len(rows)-1] += w + m.TextStyle.Render(" ")
	}

	count, start, end := 0, 0, len(rows)
	for i, row := range rows {
		if pos >= count && pos <= count+len(row) {
			if i == 0 {
				start, end = 0, 3
			} else if i == len(rows)-1 {
				start, end = len(rows)-3, len(rows)
			} else {
				start, end = i-1, i+2
			}
		}
		count += len(row)
	}

	result := ""
	for _, row := range rows[start:end] {
		result += row + m.TextStyle.Render("\n")
	}
	return result
}

func Blink() tea.Msg {
	return cursor.Blink()
}
