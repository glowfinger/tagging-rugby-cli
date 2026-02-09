// Package forms provides huh-based form components for the TUI.
package forms

import (
	"github.com/charmbracelet/huh"
)

// NewConfirmDiscardForm creates a huh confirm form asking the user whether to discard form data.
// The result pointer is bound to the confirm field value.
func NewConfirmDiscardForm(discard *bool) *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Discard changes?").
				Description("You have unsaved data. Are you sure you want to discard?").
				Affirmative("Yes, discard").
				Negative("No, go back").
				Value(discard),
		),
	).WithTheme(Theme())
}
