package forms

import (
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/user/tagging-rugby-cli/tui/styles"
)

// Theme returns a custom huh theme that matches the TUI color palette.
func Theme() *huh.Theme {
	t := huh.ThemeBase()

	// Focused field styles
	t.Focused.Base = t.Focused.Base.
		BorderStyle(lipgloss.ThickBorder()).
		BorderLeft(true).
		BorderForeground(styles.BrightPurple).
		PaddingLeft(1)

	t.Focused.Title = lipgloss.NewStyle().
		Foreground(styles.Pink).
		Bold(true)

	t.Focused.Description = lipgloss.NewStyle().
		Foreground(styles.Lavender)

	t.Focused.ErrorIndicator = lipgloss.NewStyle().
		Foreground(styles.Pink).
		Bold(true)

	t.Focused.ErrorMessage = lipgloss.NewStyle().
		Foreground(styles.Pink)

	t.Focused.SelectSelector = lipgloss.NewStyle().
		SetString("▸ ").
		Foreground(styles.Cyan)

	t.Focused.Option = lipgloss.NewStyle().
		Foreground(styles.LightLavender)

	t.Focused.NextIndicator = lipgloss.NewStyle().
		Foreground(styles.Lavender)

	t.Focused.PrevIndicator = lipgloss.NewStyle().
		Foreground(styles.Lavender)

	t.Focused.MultiSelectSelector = lipgloss.NewStyle().
		SetString("▸ ").
		Foreground(styles.Cyan)

	t.Focused.SelectedOption = lipgloss.NewStyle().
		Foreground(styles.Cyan)

	t.Focused.SelectedPrefix = lipgloss.NewStyle().
		SetString("[✓] ").
		Foreground(styles.Cyan)

	t.Focused.UnselectedOption = lipgloss.NewStyle().
		Foreground(styles.Lavender)

	t.Focused.UnselectedPrefix = lipgloss.NewStyle().
		SetString("[ ] ").
		Foreground(styles.Lavender)

	t.Focused.TextInput.Cursor = lipgloss.NewStyle().
		Foreground(styles.Cyan)

	t.Focused.TextInput.Placeholder = lipgloss.NewStyle().
		Foreground(styles.Purple)

	t.Focused.TextInput.Prompt = lipgloss.NewStyle().
		Foreground(styles.Cyan)

	t.Focused.TextInput.Text = lipgloss.NewStyle().
		Foreground(styles.LightLavender)

	t.Focused.FocusedButton = lipgloss.NewStyle().
		Background(styles.BrightPurple).
		Foreground(styles.LightLavender).
		Bold(true).
		Padding(0, 1)

	t.Focused.BlurredButton = lipgloss.NewStyle().
		Background(styles.Purple).
		Foreground(styles.Lavender).
		Padding(0, 1)

	t.Focused.Card = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(styles.Purple).
		Padding(0, 1)

	t.Focused.NoteTitle = lipgloss.NewStyle().
		Foreground(styles.Cyan).
		Bold(true)

	t.Focused.Next = t.Focused.FocusedButton

	// Blurred field styles
	t.Blurred.Base = t.Blurred.Base.
		BorderStyle(lipgloss.HiddenBorder()).
		BorderLeft(true).
		PaddingLeft(1)

	t.Blurred.Title = lipgloss.NewStyle().
		Foreground(styles.Lavender)

	t.Blurred.Description = lipgloss.NewStyle().
		Foreground(styles.Purple)

	t.Blurred.ErrorIndicator = lipgloss.NewStyle().
		Foreground(styles.Pink)

	t.Blurred.ErrorMessage = lipgloss.NewStyle().
		Foreground(styles.Pink)

	t.Blurred.SelectSelector = lipgloss.NewStyle().
		SetString("  ")

	t.Blurred.Option = lipgloss.NewStyle().
		Foreground(styles.Lavender)

	t.Blurred.MultiSelectSelector = lipgloss.NewStyle().
		SetString("  ")

	t.Blurred.SelectedOption = lipgloss.NewStyle().
		Foreground(styles.Lavender)

	t.Blurred.SelectedPrefix = lipgloss.NewStyle().
		SetString("[✓] ").
		Foreground(styles.Lavender)

	t.Blurred.UnselectedOption = lipgloss.NewStyle().
		Foreground(styles.Purple)

	t.Blurred.UnselectedPrefix = lipgloss.NewStyle().
		SetString("[ ] ").
		Foreground(styles.Purple)

	t.Blurred.TextInput.Cursor = lipgloss.NewStyle().
		Foreground(styles.Purple)

	t.Blurred.TextInput.Placeholder = lipgloss.NewStyle().
		Foreground(styles.Purple)

	t.Blurred.TextInput.Prompt = lipgloss.NewStyle().
		Foreground(styles.Purple)

	t.Blurred.TextInput.Text = lipgloss.NewStyle().
		Foreground(styles.Lavender)

	t.Blurred.FocusedButton = lipgloss.NewStyle().
		Background(styles.Purple).
		Foreground(styles.Lavender).
		Padding(0, 1)

	t.Blurred.BlurredButton = lipgloss.NewStyle().
		Background(styles.DeepPurple).
		Foreground(styles.Purple).
		Padding(0, 1)

	t.Blurred.Card = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(styles.DeepPurple).
		Padding(0, 1)

	t.Blurred.NoteTitle = lipgloss.NewStyle().
		Foreground(styles.Lavender)

	t.Blurred.Next = t.Blurred.FocusedButton

	return t
}
