package forms

import (
	"fmt"
	"strconv"

	"github.com/charmbracelet/huh"
	"github.com/user/tagging-rugby-cli/pkg/timeutil"
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

// EditTackleFormResult extends TackleFormResult with editable timestamp and end seconds.
type EditTackleFormResult struct {
	TackleFormResult

	// Editable timestamp (displayed and stored as float64 string)
	Timestamp string
	// End seconds — how many seconds after start the end should be
	EndSeconds string
}

// NewTackleForm creates a multi-step huh wizard form for tackle input.
// The timestamp is displayed as a header in H:MM:SS format.
// The result pointer is bound to the form fields and will be populated on submit.
func NewTackleForm(timestamp float64, result *TackleFormResult) *huh.Form {
	header := fmt.Sprintf("Add Tackle @ %s", timeutil.FormatTime(timestamp))

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

// NewEditTackleForm creates a multi-step huh wizard form for editing an existing tackle.
// The form is pre-filled with values from the result, and includes editable Timestamp and End seconds fields.
// The editResult pointer is bound to the form fields and will be populated on submit.
func NewEditTackleForm(timestamp float64, endSeconds float64, result *EditTackleFormResult) *huh.Form {
	// Pre-fill timestamp and end seconds as strings for the form inputs
	result.Timestamp = fmt.Sprintf("%g", timestamp)
	result.EndSeconds = fmt.Sprintf("%g", endSeconds)

	header := fmt.Sprintf("Edit Tackle @ %s", timeutil.FormatTime(timestamp))

	form := huh.NewForm(
		// Step 1: Tackle fields + timing (maps to note_tackles + note_timing)
		huh.NewGroup(
			huh.NewNote().Title(header).Description("Step 1 of 2: Tackle Details"),

			huh.NewInput().
				Title("Timestamp").
				Description("H:MM:SS, MM:SS, or seconds").
				Value(&result.Timestamp).
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("timestamp is required")
					}
					if _, err := timeutil.ParseTimeToSeconds(s); err != nil {
						return fmt.Errorf("invalid time format")
					}
					return nil
				}),

			huh.NewInput().
				Title("End (seconds)").
				Description("Seconds after start for end time").
				Value(&result.EndSeconds).
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("end seconds is required")
					}
					val, err := strconv.ParseFloat(s, 64)
					if err != nil {
						return fmt.Errorf("must be a number")
					}
					if val <= 0 {
						return fmt.Errorf("must be a positive number")
					}
					return nil
				}),

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
