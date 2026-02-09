// Package styles provides Lipgloss styles for the TUI using the Ciapre colour palette.
package styles

import "github.com/charmbracelet/lipgloss"

// Color palette - Ciapre (warm, earthy) theme from Gogh
const (
	// DeepPurple is the main background colour (Ciapre background)
	DeepPurple = lipgloss.Color("#191C27")
	// DarkPurple is a secondary dark background (Ciapre ANSI 0 black)
	DarkPurple = lipgloss.Color("#181818")
	// Purple is the border/dim accent colour (Ciapre ANSI 6 brown)
	Purple = lipgloss.Color("#5C4F4B")
	// BrightPurple is used for highlights and focus states (Ciapre ANSI 5 magenta)
	BrightPurple = lipgloss.Color("#724D7C")
	// Lavender is a secondary text colour (Ciapre foreground)
	Lavender = lipgloss.Color("#AEA47A")
	// LightLavender is the primary text colour (Ciapre ANSI 14 cream)
	LightLavender = lipgloss.Color("#F3DBB2")
	// Pink is an accent colour for headers and special elements (Ciapre ANSI 13 bright magenta)
	Pink = lipgloss.Color("#D33061")
	// Cyan is an accent colour for information and interactive elements (Ciapre ANSI 12 bright blue)
	Cyan = lipgloss.Color("#3097C6")
	// Amber is a warm accent for sub-headers (Ciapre derived)
	Amber = lipgloss.Color("#CC8B3F")
	// Red is used for warnings and errors (Ciapre ANSI 1)
	Red = lipgloss.Color("#AC3835")
	// Green is used for success messages (Ciapre ANSI 2)
	Green = lipgloss.Color("#A6A75D")
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
	Foreground(Red).
	Bold(true)

// Success is the style for success messages
var Success = lipgloss.NewStyle().
	Foreground(Green).
	Bold(true)
