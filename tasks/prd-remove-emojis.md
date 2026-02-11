# PRD: Remove Emojis from TUI

## Introduction

Remove pictographic emoji characters from the TUI while keeping simple ASCII/Unicode symbols (▶, ⏸, ★, ▸, ✓, box-drawing, etc.). The controls bar gets a redesigned layout: emoji icons removed, shortcut key and label on the same line, grouped inside bordered containers. Other components have their pictographic emojis replaced with plain text.

## Goals

- Remove all pictographic emojis (🔇, 📺, 🎬, 🔄, 📝, 📊, 🚪, 🥇, 🥈, 🥉, ⚠, ❓, ➖, ➕, ⏪, ⏩, ⏯, ⏮, ⏭) from every TUI component
- Keep simple ASCII/Unicode symbols (▶, ⏸, ★, ▸, ✓, ◆, ━, █, box-drawing characters, etc.)
- Redesign the controls bar: remove emoji column, show `[Key] Label` on one line per control, wrap each group in a bordered container
- Replace pictographic status indicators (mute, overlay, warning) with short text labels
- Replace medal emojis in leaderboard with plain rank numbers

## User Stories

### US-095: Redesign controls bar with bordered group containers
**Description:** As a user, I want the controls bar to show shortcut key and label on the same line inside bordered group containers, without emoji icons, so controls are clean and scannable.

**Acceptance Criteria:**
- [ ] Remove the `Emoji` field from the `Control` struct in `controls.go`
- [ ] Remove all emoji values from `GetControlGroups()` control definitions
- [ ] Each control renders as a single line: `[Key] Label` (e.g. `[Space] Play`, `[H/←] Back`)
- [ ] Each control group is wrapped in a bordered container using `RenderInfoBox` (or equivalent box-drawing), with the group name as the header (Playback, Navigation, Step / Overlay, Views)
- [ ] Groups are stacked vertically in Column 1
- [ ] Controls within a group are listed one per line inside the container
- [ ] Typecheck/lint passes (`CGO_ENABLED=0 go vet ./...`)

### US-096: Remove pictographic emojis from status bar
**Description:** As a user, I want the status bar to use short text labels instead of pictographic emojis for mute and overlay indicators.

**Acceptance Criteria:**
- [ ] Keep `▶` (play) and `⏸` (pause) icons — these are simple Unicode symbols
- [ ] Replace mute emoji `🔇` with text `MUTED`
- [ ] Replace overlay emoji `📺` with text `OVL`
- [ ] Status bar remains properly aligned with left/right content at all widths
- [ ] Typecheck/lint passes

### US-097: Remove pictographic emojis from mini player
**Description:** As a user, I want the mini player to use text labels instead of pictographic emojis.

**Acceptance Criteria:**
- [ ] Keep `▶` and `⏸` icons — same as status bar
- [ ] Replace mute emoji `🔇` with text `MUTED`
- [ ] Replace warning emoji `⚠` with text `!` or `WARNING:`
- [ ] Mini player auto-sizing still works correctly with new text widths
- [ ] Typecheck/lint passes

### US-098: Remove medal emojis from stats panel
**Description:** As a user, I want the stats panel leaderboard to use plain text rank numbers instead of medal emojis.

**Acceptance Criteria:**
- [ ] Replace `🥇` with `#1`
- [ ] Replace `🥈` with `#2`
- [ ] Replace `🥉` with `#3`
- [ ] Ranks 4 and 5 formatting consistent with 1–3 (e.g. `#4`, `#5`)
- [ ] Bar graph `█` character is kept (block element, not emoji)
- [ ] Column alignment in leaderboard is preserved
- [ ] Typecheck/lint passes

## Functional Requirements

- FR-1: `Control` struct loses its `Emoji` field; all control definitions updated
- FR-2: `ControlsDisplay` renders each group as a bordered box (via `RenderInfoBox`) with controls listed as `[Key] Label` lines
- FR-3: Status bar replaces `🔇` → `MUTED`, `📺` → `OVL`; keeps `▶` and `⏸`
- FR-4: Mini player replaces `🔇` → `MUTED`, `⚠` → `!`; keeps `▶` and `⏸`
- FR-5: Stats panel replaces medal emojis with `#1`–`#5` rank labels
- FR-6: All simple Unicode symbols (▶, ⏸, ★, ▸, ✓, ◆, ━, ╸, ─, ▲, █) and box-drawing characters are kept unchanged

## Non-Goals

- No changes to box-drawing characters (borders, containers)
- No changes to timeline/progress bar characters
- No changes to bar graph block character
- No changes to form theme symbols (▸, ✓)
- No changes to star symbol (★) in notes list
- No layout changes beyond the controls bar redesign
- No new features — this is an emoji cleanup pass
- No changes to keybindings or functionality

## Technical Considerations

- Removing the `Emoji` field from `Control` struct requires updating all call sites
- `RenderInfoBox` already exists in `controls.go` and handles bordered containers with tab-style headers — reuse it for control groups
- Medal emoji replacement changes character width — verify leaderboard column alignment
- `CGO_ENABLED=0` required for all Go build/vet commands

## Success Metrics

- Zero pictographic emoji characters remain in any TUI render output
- Controls bar is easy to scan with grouped bordered containers and `[Key] Label` format
- Status bar and mini player render cleanly without alignment issues
- App builds and runs successfully
