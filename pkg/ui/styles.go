package ui

import (
	"github.com/charmbracelet/lipgloss"
)

// Common styles for UI components
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#00FF00")).
			MarginLeft(2)

	infoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888")).
			MarginLeft(2)

	seasonStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFD700")).
			MarginLeft(4)

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
