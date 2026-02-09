package forms

import (
	"fmt"
	"strconv"

	"github.com/charmbracelet/huh"
)

// TackleFormResult holds the data returned by a completed tackle wizard.
type TackleFormResult struct {
	// Step 1: Note fields
	Text     string
	Category string
	Player   string
	Team     string

	// Step 2: Tackle fields
	Attempt string
	Outcome string

	// Step 3: Optional fields
	Followed string
	Notes    string
	Zone     string

	// Star toggle
	Star bool
}

// NewTackleForm creates a multi-step huh wizard form for tackle input.
// The timestamp is displayed as a header in MM:SS format.
// The result pointer is bound to the form fields and will be populated on submit.
func NewTackleForm(timestamp float64, result *TackleFormResult) *huh.Form {
	totalSeconds := int(timestamp)
	mins := totalSeconds / 60
	secs := totalSeconds % 60
	header := fmt.Sprintf("Add Tackle @ %d:%02d", mins, secs)

	form := huh.NewForm(
		// Step 1: Note fields
		huh.NewGroup(
			huh.NewNote().Title(header).Description("Step 1 of 3: Note Details"),

			huh.NewInput().
				Title("Text").
				Description("Required").
				Value(&result.Text).
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("text is required")
					}
					return nil
				}),

			huh.NewInput().
				Title("Category").
				Description("Optional").
				Value(&result.Category),

			huh.NewInput().
				Title("Player").
				Description("Optional").
				Value(&result.Player),

			huh.NewInput().
				Title("Team").
				Description("Optional").
				Value(&result.Team),
		),

		// Step 2: Tackle fields
		huh.NewGroup(
			huh.NewNote().Title(header).Description("Step 2 of 3: Tackle Details"),

			huh.NewInput().
				Title("Attempt").
				Description("Required - number only").
				Value(&result.Attempt).
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("attempt is required")
					}
					if _, err := strconv.Atoi(s); err != nil {
						return fmt.Errorf("attempt must be a number")
					}
					return nil
				}),

			huh.NewSelect[string]().
				Title("Outcome").
				Description("Required").
				Options(
					huh.NewOption("Completed", "completed"),
					huh.NewOption("Missed", "missed"),
					huh.NewOption("Possible", "possible"),
					huh.NewOption("Other", "other"),
				).
				Value(&result.Outcome),
		),

		// Step 3: Optional fields
		huh.NewGroup(
			huh.NewNote().Title(header).Description("Step 3 of 3: Optional Details"),

			huh.NewInput().
				Title("Followed").
				Description("Optional - who followed up").
				Value(&result.Followed),

			huh.NewInput().
				Title("Notes").
				Description("Optional - additional notes").
				Value(&result.Notes),

			huh.NewInput().
				Title("Zone").
				Description("Optional - field zone").
				Value(&result.Zone),

			huh.NewConfirm().
				Title("Star").
				Description("Mark as highlighted").
				Value(&result.Star),
		),
	).WithTheme(Theme())

	return form
}
