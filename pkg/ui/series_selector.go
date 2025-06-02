package ui

import (
	"fmt"
	"sort"

	"tt6d/pkg/extractor"

	tea "github.com/charmbracelet/bubbletea"
)

type seriesModel struct {
	title        string
	seasons      []string
	episodes     map[string][]extractor.Episode
	cursor       int
	selected     map[string]bool
	currentState viewState
	viewport     struct {
		start int
		size  int
	}
}

type viewState int

const (
	seasonSelect viewState = iota
	episodeSelect
	confirmSelect
)

func (m seriesModel) Init() tea.Cmd {
	m.viewport.size = 10
	return nil
}

func (m seriesModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			if m.cursor < len(m.currentItems())-1 {
				m.cursor++
			}

		case "enter":
			switch m.currentState {
			case seasonSelect:
				m.currentState = episodeSelect
				m.cursor = 0
			case episodeSelect:
				season := m.seasons[m.cursor]
				m.selected[season] = !m.selected[season]
			case confirmSelect:
				if len(m.selected) > 0 {
					return m, tea.Quit
				}
			}

		case "esc":
			if m.currentState > seasonSelect {
				m.currentState--
				m.cursor = 0
			}

		case "a":
			switch m.currentState {
			case seasonSelect:
				m.selected = make(map[string]bool)
				for _, season := range m.seasons {
					m.selected[season] = true
				}
			case episodeSelect:
				season := m.seasons[m.cursor]
				for _, ep := range m.episodes[season] {
					m.selected[ep.ID] = true
				}
			}

		case "n":
			// Deselect all
			m.selected = make(map[string]bool)
		}
	}

	return m, nil
}

func (m seriesModel) currentItems() []string {
	switch m.currentState {
	case seasonSelect:
		return m.seasons
	case episodeSelect:
		if m.cursor < len(m.seasons) {
			var eps []string
			for _, ep := range m.episodes[m.seasons[m.cursor]] {
				eps = append(eps, ep.ID)
			}
			return eps
		}
	}
	return nil
}

func (m seriesModel) View() string {
	s := titleStyle.Render(m.title) + "\n\n"

	switch m.currentState {
	case seasonSelect:
		s += infoStyle.Render("Select a season:") + "\n\n"
		for i, season := range m.seasons {
			cursor := " "
			if m.cursor == i {
				cursor = "▸"
			}

			checked := "[ ]"
			if m.selected[season] {
				checked = "[✓]"
			}

			item := fmt.Sprintf("%s %s Season %s", cursor, checked, season)

			if m.cursor == i {
				s += seasonStyle.Render(item)
			} else {
				s += itemStyle.Render(item)
			}
			s += "\n"
		}

	case episodeSelect:
		if m.cursor < len(m.seasons) {
			season := m.seasons[m.cursor]
			s += seasonStyle.Render(fmt.Sprintf("Season %s Episodes:", season)) + "\n\n"

			for i, ep := range m.episodes[season] {
				cursor := " "
				if m.cursor == i {
					cursor = "▸"
				}

				checked := "[ ]"
				if m.selected[ep.ID] {
					checked = "[✓]"
				}

				item := fmt.Sprintf("%s %s %s", cursor, checked, ep.ID)

				if m.cursor == i {
					s += selectedItemStyle.Render(item)
				} else {
					s += itemStyle.Render(item)
				}
				s += "\n"
			}
		}
	}

	// Help footer
	s += "\n" + footerStyle.Render("Navigation: ↑/↓ or j/k • Enter: select • Esc: back")
	s += "\n" + footerStyle.Render("Actions: space: toggle • a: select all • n: none • q: quit")

	return s
}

func SelectTVSeriesEpisodes(info *extractor.TVSeriesInfo) ([]string, error) {
	var seasons []string
	for season := range info.Seasons {
		seasons = append(seasons, season)
	}

	// Sort seasons
	sort.Strings(seasons)

	p := tea.NewProgram(seriesModel{
		title:        info.Title,
		seasons:      seasons,
		episodes:     info.Seasons,
		selected:     make(map[string]bool),
		currentState: seasonSelect,
	})

	m, err := p.Run()
	if err != nil {
		return nil, fmt.Errorf("failed to run UI: %v", err)
	}

	finalModel := m.(seriesModel)
	if len(finalModel.selected) == 0 {
		return nil, fmt.Errorf("no episodes selected")
	}

	var selectedLinks []string
	for season := range finalModel.selected {
		for _, ep := range info.Seasons[season] {
			selectedLinks = append(selectedLinks, ep.Links...)
		}
	}

	return selectedLinks, nil
}
