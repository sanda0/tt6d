package ui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#00FF00")).
			MarginLeft(2)

	itemStyle = lipgloss.NewStyle().
			PaddingLeft(4)

	selectedItemStyle = lipgloss.NewStyle().
				PaddingLeft(2).
				Foreground(lipgloss.Color("#00FF00")).
				SetString("▸ ")

	footerStyle = lipgloss.NewStyle().
			MarginLeft(2).
			MarginTop(1).
			Foreground(lipgloss.Color("#888888"))
)

type model struct {
	links    []string
	cursor   int
	selected map[int]bool
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.links)-1 {
				m.cursor++
			}

		case " ", "enter":
			if m.selected[m.cursor] {
				delete(m.selected, m.cursor)
			} else {
				m.selected[m.cursor] = true
			}

		case "a":
			// Select all
			for i := range m.links {
				m.selected[i] = true
			}

		case "n":
			// Deselect all
			m.selected = make(map[int]bool)
		}
	}

	return m, nil
}

func (m model) View() string {
	s := titleStyle.Render("Select MP4 files to download") + "\n\n"

	for i, link := range m.links {
		cursor := " "
		if m.cursor == i {
			cursor = "▸"
		}

		checked := " "
		if m.selected[i] {
			checked = "✓"
		}

		// Shorten the link for display if it's too long
		displayLink := link
		if len(displayLink) > 70 {
			displayLink = displayLink[:35] + "..." + displayLink[len(displayLink)-32:]
		}

		item := fmt.Sprintf("%s [%s] %s", cursor, checked, displayLink)

		if m.cursor == i {
			s += selectedItemStyle.Render(item)
		} else {
			s += itemStyle.Render(item)
		}
		s += "\n"
	}

	s += "\n" + footerStyle.Render("↑/↓: navigate • space: toggle • a: select all • n: deselect all • enter: confirm • q: quit")

	return s
}

func GetSelectedLinks(links []string) ([]string, error) {
	p := tea.NewProgram(model{
		links:    links,
		selected: make(map[int]bool),
	})

	m, err := p.Run()
	if err != nil {
		return nil, fmt.Errorf("failed to run UI: %v", err)
	}

	finalModel := m.(model)
	if len(finalModel.selected) == 0 {
		return nil, fmt.Errorf("no files selected")
	}

	var selectedLinks []string
	for i := range finalModel.selected {
		selectedLinks = append(selectedLinks, links[i])
	}

	return selectedLinks, nil
}
