package forms

import (
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/user/tagging-rugby-cli/pkg/timeutil"
)

// NoteFormResult holds the data returned by a completed note form.
type NoteFormResult struct {
	Text     string
	Category string
	Player   string
	Team     string
}

// HasData returns true if any field in the note form result has a non-empty value.
func (r *NoteFormResult) HasData() bool {
	return r.Text != "" || r.Category != "" || r.Player != "" || r.Team != ""
}

// NewNoteForm creates a huh form for note input with the given timestamp.
// The timestamp is displayed as a header in H:MM:SS format.
// The result pointer is bound to the form fields and will be populated on submit.
func NewNoteForm(timestamp float64, result *NoteFormResult) *huh.Form {
	header := fmt.Sprintf("Add Note @ %s", timeutil.FormatTime(timestamp))

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewNote().Title(header),

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
	).WithTheme(Theme())

	return form
}
