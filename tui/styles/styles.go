// Package styles provides Lipgloss styles for the TUI using a retro purple color palette.
package styles

import "github.com/charmbracelet/lipgloss"

// Color palette - retro purple theme
const (
	// DeepPurple is the darkest background color
	DeepPurple = lipgloss.Color("#1a1a2e")
	// DarkPurple is a secondary dark background
	DarkPurple = lipgloss.Color("#16213e")
	// Purple is the main accent color
	Purple = lipgloss.Color("#4a347d")
	// BrightPurple is used for highlights and focus states
	BrightPurple = lipgloss.Color("#7b2cbf")
	// Lavender is a light accent color
	Lavender = lipgloss.Color("#c77dff")
	// LightLavender is the lightest purple for text
	LightLavender = lipgloss.Color("#e0aaff")
	// Pink is an accent color for special elements
	Pink = lipgloss.Color("#ff6b9d")
	// Cyan is an accent color for information
	Cyan = lipgloss.Color("#64dfdf")
)

// Pre-defined styles using the color palette

// Background is the main background style for the entire TUI
var Background = lipgloss.NewStyle().
	Background(DeepPurple)

// Panel is the style for content panels
var Panel = lipgloss.NewStyle().
	Background(DarkPurple).
	Padding(1, 2)

// Border is the style for bordered panels
var Border = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(Purple)

// Highlight is the style for selected/highlighted items
var Highlight = lipgloss.NewStyle().
	Background(BrightPurple).
	Foreground(LightLavender).
	Bold(true)

// PrimaryText is the style for primary text content
var PrimaryText = lipgloss.NewStyle().
	Foreground(LightLavender)

// SecondaryText is the style for less prominent text
var SecondaryText = lipgloss.NewStyle().
	Foreground(Lavender)

// Warning is the style for warning messages
var Warning = lipgloss.NewStyle().
	Foreground(Pink).
	Bold(true)

// Success is the style for success messages
var Success = lipgloss.NewStyle().
	Foreground(Cyan).
	Bold(true)
