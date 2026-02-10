# PRD: Tidy Up UI

## Introduction

Polish the TUI's visual presentation — clean up jagged column borders, reorder Column 1 sections so playback/selected-tag info comes first and controls sit below with labelled group sub-headers, add breathing room between control items with keys on their own line, remove deprecated code, and adopt the Ciapre colour scheme from Gogh for a warm, earthy aesthetic.

## Goals

- Columns render with clean, single-line vertical borders (no jagged/misaligned rows)
- Column 1 layout reordered: Playback → Selected Tag → Controls
- Controls display redesigned: icon + name on one line, shortcut key on the next, with sub-headers per group and spacing between items
- Deprecated input components (`noteinput.go`, `tackleinput.go`) removed
- Colour palette replaced with Ciapre from Gogh

## Ciapre Colour Palette Reference

Source: Gogh terminal themes — `Ciapre.yml`

### Raw palette

| ANSI       | Hex       | Ciapre Label     |
|------------|-----------|------------------|
| Background | `#191C27` | Background       |
| Foreground | `#AEA47A` | Foreground/Text  |
| Cursor     | `#AEA47A` | Cursor           |
| 0  Black   | `#181818` | Host             |
| 1  Red     | `#810009` | Syntax string    |
| 2  Green   | `#48513B` | Command          |
| 3  Yellow  | `#CC8B3F` | Command second   |
| 4  Blue    | `#576D8C` | Path             |
| 5  Magenta | `#724D7C` | Syntax var       |
| 6  Cyan    | `#5C4F4B` | Prompt           |
| 7  White   | `#AEA47F` | White            |
| 8  Bright Black   | `#555555` | —         |
| 9  Bright Red     | `#AC3835` | Command error |
| 10 Bright Green   | `#A6A75D` | Exec        |
| 11 Bright Yellow  | `#DCDF7C` | —           |
| 12 Bright Blue    | `#3097C6` | Folder      |
| 13 Bright Magenta | `#D33061` | —           |
| 14 Bright Cyan    | `#F3DBB2` | —           |
| 15 Bright White   | `#F4F4F4` | —           |

### Mapped to TUI roles

| Current Name    | Current Hex | New Hex   | Ciapre Source        | Notes                                    |
|-----------------|-------------|-----------|----------------------|------------------------------------------|
| DeepPurple      | `#1a1a2e`   | `#191C27` | Background           | Main app background (very close to current) |
| DarkPurple      | `#16213e`   | `#181818` | Black (ANSI 0)       | Secondary background for status bar, panels |
| Purple          | `#4a347d`   | `#5C4F4B` | Cyan/Prompt (ANSI 6) | Borders, column separators, subtle accents  |
| BrightPurple    | `#7b2cbf`   | `#724D7C` | Magenta (ANSI 5)     | Selection highlight, focused elements       |
| Lavender        | `#c77dff`   | `#AEA47A` | Foreground           | Secondary/dim text                          |
| LightLavender   | `#e0aaff`   | `#F3DBB2` | Bright Cyan (ANSI 14)| Primary readable text (cream)               |
| Pink            | `#ff6b9d`   | `#D33061` | Bright Magenta (ANSI 13) | Section headers, special emphasis       |
| Cyan            | `#64dfdf`   | `#3097C6` | Bright Blue (ANSI 12)| Shortcut keys, interactive elements, info    |

Additional colours available for use:

| Hex       | Ciapre Source           | Suggested Use                    |
|-----------|-------------------------|----------------------------------|
| `#CC8B3F` | Yellow (ANSI 3, amber)  | Sub-headers, warm accent         |
| `#AC3835` | Bright Red (ANSI 9)     | Warning, errors                  |
| `#A6A75D` | Bright Green (ANSI 10)  | Success messages                 |
| `#DCDF7C` | Bright Yellow (ANSI 11) | Bright highlight accents         |
| `#F4F4F4` | Bright White (ANSI 15)  | Bold headings if cream isn't enough |
| `#576D8C` | Blue (ANSI 4, slate)    | Alternative header colour        |
| `#810009` | Red (ANSI 1, crimson)   | Severe warnings                  |
| `#48513B` | Green (ANSI 2, olive)   | Muted positive indicators        |
| `#555555` | Bright Black (ANSI 8)   | Disabled/placeholder text        |

## User Stories

### US-001: Fix jagged column borders
**Description:** As a user, I want the three-column layout to have perfectly aligned single-line vertical borders so the UI looks clean and professional.

**Acceptance Criteria:**
- [ ] Every row in the column view uses the same `│` border character at the same horizontal position
- [ ] `padToWidth` correctly handles ANSI-styled text so columns are always the exact expected width
- [ ] No visual gaps, double borders, or misaligned rows at any terminal width (80–200 cols)
- [ ] Column content that overflows is truncated rather than breaking alignment
- [ ] Typecheck/lint passes (`CGO_ENABLED=0 go vet ./...`)

### US-002: Reorder Column 1 — Playback and Selected Tag above Controls
**Description:** As a user, I want to see the playback status and selected tag details first in Column 1, with keyboard controls listed below, so the most dynamic information is at the top.

**Acceptance Criteria:**
- [ ] Column 1 renders sections in this order: (1) Playback, (2) Selected Tag, (3) Controls
- [ ] Each section has a styled header
- [ ] A blank line separates each section
- [ ] Typecheck/lint passes

### US-003: Redesign controls display — sub-headers, key below title, spacing
**Description:** As a user, I want each control group to have a labelled sub-header, each control to show the icon and name on one line with the shortcut key on the next line below, and visible spacing between controls so they're easy to scan.

**Acceptance Criteria:**
- [ ] Each control group has a styled sub-header label:
  - Group 1: "Playback"
  - Group 2: "Navigation"
  - Group 3: "Step / Overlay"
  - Group 4: "Views"
- [ ] Each control renders as two lines:
  - Line 1: `emoji  Name` (with a space between emoji and name)
  - Line 2: `     [Key]` (shortcut indented below, styled in Info/blue)
- [ ] At least one blank line separates each control from the next
- [ ] An extra blank line or divider separates each group
- [ ] The controls section doesn't overflow Column 1 at standard widths (80+ cols)
- [ ] Typecheck/lint passes

### US-004: Remove deprecated input components
**Description:** As a developer, I want to remove the old `noteinput.go` and `tackleinput.go` files since they've been replaced by huh forms, to reduce dead code.

**Acceptance Criteria:**
- [ ] `tui/components/noteinput.go` deleted
- [ ] `tui/components/tackleinput.go` deleted
- [ ] No remaining imports or references to `NoteInput` or `TackleInput` types from these files
- [ ] App builds and runs successfully (`CGO_ENABLED=0 go build ./...`)
- [ ] Typecheck/lint passes

### US-005: Adopt Ciapre colour palette
**Description:** As a user, I want the app to use the Ciapre colour scheme from Gogh so it has a warm, earthy look with good readability.

**Acceptance Criteria:**
- [ ] Replace all 8 colour constants in `tui/styles/styles.go` with Ciapre values (see "Mapped to TUI roles" table)
- [ ] Add new colour constants for additional Ciapre colours used (e.g. `Amber` for `#CC8B3F` sub-headers)
- [ ] Primary text (`#F3DBB2` cream) has sufficient contrast against background (`#191C27`)
- [ ] Secondary text (`#AEA47A` foreground) is readable but clearly subordinate
- [ ] Shortcut keys styled in blue (`#3097C6`)
- [ ] Control group sub-headers use amber (`#CC8B3F`)
- [ ] Selection/highlight uses magenta/purple (`#724D7C`)
- [ ] Borders/separators use `#5C4F4B` (warm brown)
- [ ] huh form theme in `tui/forms/theme.go` updated to match the new palette
- [ ] Status bar, controls, notes list, stats panel, timeline, and help overlay all use the updated palette consistently
- [ ] Typecheck/lint passes

### US-006: Verify full UI consistency
**Description:** As a user, I want the entire TUI to look consistent after all changes — column borders, spacing, colours, and layout should feel cohesive.

**Acceptance Criteria:**
- [ ] All three columns render cleanly at widths 80, 100, 120, and 160
- [ ] Two-column fallback (< 90 cols) still works with clean borders
- [ ] Status bar, timeline, and command input span full width without gaps
- [ ] No raw/unstyled text visible in any panel
- [ ] App builds and runs (`CGO_ENABLED=0 go build ./...`)

## Functional Requirements

- FR-1: `renderColumn1` must render sections in order: Playback → Selected Tag → Controls
- FR-2: Each control group has a sub-header label ("Playback", "Navigation", "Step / Overlay", "Views")
- FR-3: Each control item renders as a two-line block (icon+name, then key) with a blank line after each
- FR-4: Control groups separated by an additional blank line
- FR-5: `padToWidth` must produce exact column widths for all content including styled/emoji text
- FR-6: Column borders use a single `│` character per row at fixed horizontal positions
- FR-7: Deprecated files `noteinput.go` and `tackleinput.go` removed, no dangling references
- FR-8: All colour constants in `styles.go` replaced with Ciapre palette values
- FR-9: huh form theme updated to match the new Ciapre palette
- FR-10: All components reference the shared palette — no hardcoded colour values outside `styles.go`

## Non-Goals

- No changes to the stats view (full-screen tackle table)
- No changes to the help overlay content or layout
- No changes to huh form functionality (note/tackle input logic)
- No new features — this is purely visual polish
- No changes to command input or keybindings

## Design Considerations

- The controls section will be taller with the new two-line-per-control + sub-header layout. At standard heights (~30+ rows), Playback (5–8 lines) + Selected Tag (5–8 lines) takes ~16 lines, leaving ~14+ lines for controls. With 14 controls across 4 groups, each needing 3 lines (name, key, blank) plus sub-headers, it could need ~50 lines total. Controls should clip naturally at column bottom — the most important ones (Playback, Navigation) are first.
- The Ciapre palette blends cool dark backgrounds (`#191C27` is very close to the current `#1a1a2e`) with warm foreground tones (cream, amber, tan). The background shift is minimal; the biggest visual change is in text and accent colours moving from neon purple/pink/cyan to muted earthy tones.

## Technical Considerations

- `lipgloss.Width()` correctly measures ANSI-escaped strings — ensure all `padToWidth` calls use it rather than `len()`
- Emoji characters can be 1 or 2 cells wide depending on terminal; test with common terminals
- Column height is `m.height - 5` — verify this still works after reordering Column 1
- `CGO_ENABLED=0` required for all Go build/vet commands
- Background `#191C27` is nearly identical to the current `#1a1a2e` — minimal visual disruption for the backdrop

## Success Metrics

- Column borders form a clean, unbroken vertical line at any supported terminal width
- Controls section is scannable at a glance — a user can find a shortcut key in under 2 seconds
- Ciapre palette gives a cohesive warm look across all panels
- No visual regressions in existing components (timeline, status bar, command input)
- All deprecated code removed
