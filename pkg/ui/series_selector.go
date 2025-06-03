package ui

import (
	"fmt"
	"sort"

	"tt6d/pkg/extractor"

	tea "github.com/charmbracelet/bubbletea"
)

type seriesModel struct {
	title         string
	seasons       []string
	episodes      map[string][]extractor.Episode
	cursor        int
	selected      map[string]bool
	selectedEps   map[string]bool
	currentState  viewState
	currentSeason string
	viewport      struct {
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
			if len(m.selectedEps) > 0 {
				return m, tea.Quit
			}
			// If no episodes selected, treat as cancel
			m.selectedEps = make(map[string]bool)
			return m, tea.Quit

		case "up", "k":
			items := m.currentItems()
			if len(items) > 0 && m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			items := m.currentItems()
			if len(items) > 0 && m.cursor < len(items)-1 {
				m.cursor++
			}

		case " ":
			switch m.currentState {
			case seasonSelect:
				// Space toggles season selection but doesn't change screen
				m.selected[m.seasons[m.cursor]] = !m.selected[m.seasons[m.cursor]]
				if m.cursor < len(m.seasons)-1 {
					m.cursor++
				}
			case episodeSelect:
				// Space toggles episode selection
				if episodes := m.episodes[m.currentSeason]; len(episodes) > 0 {
					epID := episodes[m.cursor].ID
					m.selectedEps[epID] = !m.selectedEps[epID]
					if m.cursor < len(episodes)-1 {
						m.cursor++
					}
				}
			}

		case "enter":
			switch m.currentState {
			case seasonSelect:
				// Enter moves to episode selection screen
				m.currentState = episodeSelect
				m.currentSeason = m.seasons[m.cursor]
				m.cursor = 0
			case episodeSelect:
				// If we have selections, proceed with download
				if len(m.selectedEps) > 0 {
					return m, tea.Quit
				}
			}

		case "esc":
			if m.currentState > seasonSelect {
				m.currentState--
				m.cursor = 0
			} else {
				// Exit if we're at the season select screen
				m.selectedEps = make(map[string]bool)
				return m, tea.Quit
			}

		case "a":
			if m.currentState == episodeSelect {
				// Select all episodes in current season
				for _, ep := range m.episodes[m.currentSeason] {
					m.selectedEps[ep.ID] = true
				}
			}

		case "n":
			// Deselect all
			if m.currentState == episodeSelect {
				m.selectedEps = make(map[string]bool)
			}
		}
	}

	return m, nil
}

func (m seriesModel) currentItems() []string {
	switch m.currentState {
	case seasonSelect:
		return m.seasons
	case episodeSelect:
		if episodes := m.episodes[m.currentSeason]; len(episodes) > 0 {
			var eps []string
			for _, ep := range episodes {
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
		s += seasonStyle.Render(fmt.Sprintf("Season %s Episodes:", m.currentSeason)) + "\n\n"

		episodes := m.episodes[m.currentSeason]
		for i, ep := range episodes {
			cursor := " "
			if m.cursor == i {
				cursor = "▸"
			}

			checked := "[ ]"
			if m.selectedEps[ep.ID] {
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

	// Help footer
	s += "\n" + footerStyle.Render("Navigation: ↑/↓ or j/k • Enter: next • Esc: back")
	if m.currentState == episodeSelect {
		s += "\n" + footerStyle.Render("Actions: space: select • a: select all • n: none • enter: confirm")
	} else {
		s += "\n" + footerStyle.Render("Actions: space: select • enter: next • q: quit")
	}

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
		selectedEps:  make(map[string]bool),
		currentState: seasonSelect,
	})

	m, err := p.Run()
	if err != nil {
		return nil, fmt.Errorf("failed to run UI: %v", err)
	}

	finalModel := m.(seriesModel)
	if len(finalModel.selectedEps) == 0 {
		return nil, fmt.Errorf("no episodes selected")
	}

	// Use a map to deduplicate links
	linkMap := make(map[string]bool)
	var selectedLinks []string

	// Collect unique links from selected episodes
	for _, episodes := range info.Seasons {
		for _, ep := range episodes {
			if finalModel.selectedEps[ep.ID] {
				// Only add links we haven't seen before
				for _, link := range ep.Links {
					if !linkMap[link] {
						linkMap[link] = true
						selectedLinks = append(selectedLinks, link)
					}
				}
			}
		}
	}

	return selectedLinks, nil
}
