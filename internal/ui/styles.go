package ui

import "github.com/charmbracelet/lipgloss"

func TitleStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#50C878")).
		Bold(true).
		MarginTop(1).
		MarginBottom(1)
}

func SectionStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#87CEEB")).
		Bold(true).
		MarginTop(1).
		MarginBottom(1)
}

func BadgeStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#55565B")).
		Background(lipgloss.Color("#9ACD32")).
		Bold(true).
		Padding(0, 1).
		MarginRight(1)
}
