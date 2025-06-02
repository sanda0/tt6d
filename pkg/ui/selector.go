package ui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

type model struct {
	links    []string
	cursor   int
	selected map[int]bool
	viewport struct {
		start int
		size  int
	}
}

func (m model) Init() tea.Cmd {
	// Initialize viewport with a reasonable size
	m.viewport.size = 10
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.MouseMsg:
		switch msg.Type {
		case tea.MouseWheelUp:
			if m.cursor > 0 {
				m.cursor--
			}
		case tea.MouseWheelDown:
			if m.cursor < len(m.links)-1 {
				m.cursor++
			}
		}

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

		case "pgup":
			// Move cursor up by viewport size
			m.cursor -= m.viewport.size
			if m.cursor < 0 {
				m.cursor = 0
			}
			m.viewport.start -= m.viewport.size
			if m.viewport.start < 0 {
				m.viewport.start = 0
			}

		case "pgdown":
			// Move cursor down by viewport size
			m.cursor += m.viewport.size
			if m.cursor >= len(m.links) {
				m.cursor = len(m.links) - 1
			}
			m.viewport.start += m.viewport.size
			maxStart := len(m.links) - m.viewport.size
			if m.viewport.start > maxStart {
				m.viewport.start = maxStart
			}
			if m.viewport.start < 0 {
				m.viewport.start = 0
			}

		case " ":
			if m.selected[m.cursor] {
				delete(m.selected, m.cursor)
			} else {
				m.selected[m.cursor] = true
			}
			// Move cursor down after selection if possible
			if m.cursor < len(m.links)-1 {
				m.cursor++
			}

		case "enter":
			// Return selected files only if at least one is selected
			if len(m.selected) > 0 {
				return m, tea.Quit
			}

		case "a":
			// Select all
			for i := range m.links {
				m.selected[i] = true
			}

		case "n":
			// Deselect all
			m.selected = make(map[int]bool)

		case "pageup":
			// Scroll up
			if m.viewport.start > 0 {
				m.viewport.start--
				if m.viewport.start < m.cursor {
					m.cursor = m.viewport.start
				}
			}

		case "pagedown":
			// Scroll down
			if m.viewport.start+m.viewport.size < len(m.links) {
				m.viewport.start++
				if m.viewport.start+m.viewport.size > m.cursor {
					m.cursor = m.viewport.start + m.viewport.size - 1
				}
			}
		}
	}

	return m, nil
}

func (m model) View() string {
	s := titleStyle.Render("Select MP4 Files to Download") + "\n"

	// Show selection stats
	selectedCount := len(m.selected)
	totalCount := len(m.links)
	s += fmt.Sprintf("\nSelected: %d/%d files", selectedCount, totalCount) + "\n\n"

	// Adjust viewport if cursor is out of view
	if m.cursor < m.viewport.start {
		m.viewport.start = m.cursor
	} else if m.cursor >= m.viewport.start+m.viewport.size {
		m.viewport.start = m.cursor - m.viewport.size + 1
	}

	// Show page indicator if there are more items
	if m.viewport.start > 0 {
		s += "  ↑ More files above ↑\n"
	}

	// Show visible items
	end := min(m.viewport.start+m.viewport.size, len(m.links))
	for i := m.viewport.start; i < end; i++ {
		link := m.links[i]

		cursor := " "
		if m.cursor == i {
			cursor = "▸"
		}

		checked := "[ ]"
		if m.selected[i] {
			checked = "[✓]"
		}

		// Shorten the link for display if it's too long
		displayLink := link
		if len(displayLink) > 70 {
			displayLink = displayLink[:35] + "..." + displayLink[len(displayLink)-32:]
		}

		item := fmt.Sprintf("%s %s %s", cursor, checked, displayLink)

		if m.cursor == i {
			s += selectedItemStyle.Render(item)
		} else {
			s += itemStyle.Render(item)
		}
		s += "\n"
	}

	// Show page indicator if there are more items
	if end < len(m.links) {
		s += "  ↓ More files below ↓\n"
	}

	// Help footer
	s += "\n" + footerStyle.Render("Navigation: ↑/↓ or j/k • PageUp/PageDown")
	s += "\n" + footerStyle.Render("Actions: space: toggle • a: select all • n: none • enter: download • q: quit")

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
