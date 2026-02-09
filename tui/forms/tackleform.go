package forms

import (
	"fmt"
	"strconv"

	"github.com/charmbracelet/huh"
)

// TackleFormResult holds the data returned by a completed tackle wizard.
type TackleFormResult struct {
	// Step 1: Tackle fields (maps to note_tackles)
	Player  string
	Attempt string
	Outcome string

	// Step 2: Optional fields
	Followed string // maps to note_detail type="followed"
	Notes    string // maps to note_detail type="notes"
	Zone     string // maps to note_zones
	Star     bool   // maps to note_highlights type="star"
}

// HasData returns true if any user-entered field in the tackle form has data.
// Excludes Outcome (auto-populated by select widget) and Star (defaults to false).
func (r *TackleFormResult) HasData() bool {
	return r.Player != "" || r.Attempt != "" ||
		r.Followed != "" || r.Notes != "" || r.Zone != ""
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
		// Step 1: Tackle fields (maps to note_tackles)
		huh.NewGroup(
			huh.NewNote().Title(header).Description("Step 1 of 2: Tackle Details"),

			huh.NewInput().
				Title("Player").
				Description("Required").
				Value(&result.Player).
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("player is required")
					}
					return nil
				}),

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

		// Step 2: Optional fields (maps to note_details, note_zones, note_highlights)
		huh.NewGroup(
			huh.NewNote().Title(header).Description("Step 2 of 2: Optional Details"),

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
