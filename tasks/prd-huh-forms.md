# PRD: Improve Forms Using Huh

## Introduction

Replace the custom form rendering and input handling in the TUI with the [huh](https://github.com/charmbracelet/huh) library. This provides multi-step wizard-style forms, built-in validation, full keyboard navigation, and accessible form components. The note form becomes the parent entity, and specialized forms (like tackle) extend it with additional fields in a wizard flow.

## Goals

- Replace manual field rendering and input handling with huh form components
- Implement multi-step wizard forms where child entities (tackle, etc.) include note fields as the first step
- Provide real-time validation with clear error feedback
- Maintain full keyboard accessibility (Tab, Shift+Tab, arrow keys, Enter, Esc)
- Keep forms inline within existing views (not modal overlays)

## User Stories

### US-052: Add huh dependency
**Description:** As a developer, I need the huh library added to the project so that I can build forms with it.

**Acceptance Criteria:**
- [ ] Add `github.com/charmbracelet/huh` to go.mod
- [ ] Run `go mod tidy` successfully
- [ ] Typecheck passes (`go vet ./...`)

### US-053: Create base note form using huh
**Description:** As a developer, I need a reusable note form built with huh that can be embedded in other forms or used standalone.

**Acceptance Criteria:**
- [ ] Create `tui/forms/noteform.go` with a huh form for note input
- [ ] Form includes fields: Text (required), Category (optional), Player (optional), Team (optional)
- [ ] Text field has validation requiring non-empty input
- [ ] Form displays timestamp header (passed as parameter)
- [ ] Form returns structured data on submit
- [ ] Esc key cancels the form
- [ ] Typecheck passes

### US-054: Create tackle wizard form using huh
**Description:** As a user, I want a multi-step tackle form that captures note details first, then tackle-specific fields, so that data entry flows logically.

**Acceptance Criteria:**
- [ ] Create `tui/forms/tackleform.go` with a multi-step huh wizard
- [ ] Step 1: Note fields (Text, Category, Player, Team)
- [ ] Step 2: Tackle fields (Attempt as number input, Outcome as select)
- [ ] Step 3: Optional fields (Followed, Notes, Zone)
- [ ] Outcome select has options: completed, missed, possible, other
- [ ] Attempt field only accepts numeric input
- [ ] Required fields show validation errors before proceeding
- [ ] Star toggle available via keyboard shortcut (S or *)
- [ ] Progress indicator shows current step
- [ ] Typecheck passes

### US-055: Integrate note form into TUI
**Description:** As a user, I want the note input to use the new huh form so that I get better validation and keyboard navigation.

**Acceptance Criteria:**
- [ ] Replace `NoteInputState` usage in `tui/tui.go` with huh form
- [ ] Form appears inline in the center column when triggered (n key)
- [ ] Tab/Shift+Tab navigate between fields
- [ ] Enter submits when all required fields valid
- [ ] Esc cancels and returns to normal view
- [ ] Submitted data saves to database correctly
- [ ] Remove old `tui/components/noteinput.go` or mark deprecated
- [ ] Typecheck passes

### US-056: Integrate tackle form into TUI
**Description:** As a user, I want the tackle input to use the new multi-step huh form so that I can enter tackle data efficiently with wizard steps.

**Acceptance Criteria:**
- [ ] Replace `TackleInputState` usage in `tui/tui.go` with huh wizard form
- [ ] Form appears inline in the center column when triggered (t key)
- [ ] Wizard shows step progress (e.g., "Step 1 of 3")
- [ ] Can navigate back to previous steps to edit
- [ ] Enter on final step submits when all required fields valid
- [ ] Esc cancels at any step and returns to normal view
- [ ] Submitted data saves note and tackle to database correctly
- [ ] Remove old `tui/components/tackleinput.go` or mark deprecated
- [ ] Typecheck passes

### US-057: Style huh forms to match TUI theme
**Description:** As a user, I want the huh forms to match the existing TUI color scheme so that the interface feels cohesive.

**Acceptance Criteria:**
- [ ] Create custom huh theme in `tui/forms/theme.go`
- [ ] Use existing colors from `tui/styles/styles.go` (Purple, Pink, Lavender, Cyan, etc.)
- [ ] Active field highlights match current style (LightLavender on Purple background)
- [ ] Required field labels use Pink color
- [ ] Optional field labels use Lavender color
- [ ] Error messages use appropriate contrast color
- [ ] Typecheck passes

### US-063: Frame-by-frame navigation with Ctrl modifier
**Description:** As a user, I want to hold Ctrl while pressing H/L so that I can step forward or backward by exactly one frame for precise positioning.

**Acceptance Criteria:**
- [ ] `ctrl+h` steps backward by one frame using `FrameBackStep()`
- [ ] `ctrl+l` steps forward by one frame using `FrameStep()`
- [ ] Regular `h`/`l` continue to seek by the current step size (no regression)
- [ ] Frame step only fires when mpv client is connected
- [ ] Typecheck passes (`CGO_ENABLED=0 go vet ./...`)

### US-064: Remove q as quit shortcut
**Description:** As a user, I don't want pressing `q` to exit the application so that I can use `q` without accidentally quitting.

**Acceptance Criteria:**
- [ ] Remove `"q"` from the quit key case in `tui/tui.go` (line ~203)
- [ ] `ctrl+c` remains as the only way to quit
- [ ] Pressing `q` in normal mode does nothing (no crash, no side effect)
- [ ] Typecheck passes (`CGO_ENABLED=0 go vet ./...`)

## Functional Requirements

- FR-1: Add `github.com/charmbracelet/huh` as a dependency
- FR-2: Create `tui/forms/` package for huh-based form components
- FR-3: Note form includes Text (required), Category, Player, Team fields
- FR-4: Tackle wizard has 3 steps: Note fields, Tackle required fields, Optional fields
- FR-5: All required fields validate before form submission
- FR-6: Tab/Shift+Tab navigate between fields within a step
- FR-7: Enter advances to next step or submits on final step
- FR-8: Esc cancels form at any point
- FR-9: Wizard forms show step progress indicator
- FR-10: Forms render inline in the existing TUI layout (center column)
- FR-11: Custom theme applies existing color palette to huh components
- FR-12: Star toggle available in tackle form via S or * key

## Non-Goals

- No modal/overlay form presentation (forms stay inline)
- No changes to the three-column TUI layout
- No changes to database schema or data model
- No new form types beyond note and tackle
- No autocomplete or suggestion features for text fields

## Technical Considerations

- huh forms are Bubble Tea models that integrate with the existing TUI model
- The main TUI model needs to manage form state and delegate updates when a form is active
- huh provides `huh.NewForm()` for single-page forms and groups for wizard steps
- Custom themes use `huh.ThemeBase()` or `huh.ThemeCharm()` as starting points
- Forms should be sized to fit within the center column width
- Consider using `huh.WithWidth()` and `huh.WithHeight()` for responsive sizing

## Success Metrics

- Forms pass accessibility requirements (full keyboard navigation)
- Required field validation prevents submission of incomplete data
- Wizard steps reduce cognitive load by grouping related fields
- No regressions in form submission or data persistence
- Code is simpler than manual input handling (fewer lines, clearer logic)

## Open Questions

- Should there be a keyboard shortcut to jump directly to a specific wizard step?
- Should field values persist if user cancels partway through (for retry)?
