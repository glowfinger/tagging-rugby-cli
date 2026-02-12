# PRD: Extract Layout Package & Document TUI Architecture

## Introduction

The main TUI file (`tui/tui.go`, 2173 lines) contains ~170 lines of layout logic — column rendering, responsive breakpoints, line normalization, and width padding — interleaved with the Bubble Tea model, Update, and View logic. This makes the file harder to navigate and the layout system harder to reason about in isolation.

This feature extracts all layout concerns into a dedicated `tui/layout/` package and documents the full TUI layer in a `TUI-ARCHITECTURE.md` file so contributors can understand the rendering pipeline, component contracts, and layout system at a glance.

## Goals

- Move all layout logic out of `tui/tui.go` into a new `tui/layout/` package
- Create a reusable `Container` component that guarantees exact-height output with scroll indicators
- Extract responsive breakpoint logic into a dedicated layout manager
- Reduce `tui/tui.go` by ~170 lines, keeping it focused on state and message handling
- Document the full TUI layer (`tui/` directory) in `TUI-ARCHITECTURE.md`
- Zero behavior change — pure refactor with identical rendered output

## User Stories

### US-001: Create `tui/layout/` package with `padToWidth` and `normalizeLines`

**Description:** As a developer, I want low-level layout utilities in their own package so column rendering code can use them without depending on the main TUI model.

**Acceptance Criteria:**
- [ ] New file `tui/layout/helpers.go` created
- [ ] `padToWidth(s string, width int) string` moved from `tui/tui.go` to `layout.PadToWidth` (exported)
- [ ] `normalizeLines(lines []string, height int) []string` moved from `tui/tui.go` to `layout.NormalizeLines` (exported)
- [ ] All call sites in `tui/tui.go` updated to use `layout.PadToWidth` / `layout.NormalizeLines`
- [ ] Existing tests (if any) still pass; `go vet` passes with `CGO_ENABLED=0`

### US-002: Create `Container` component in `tui/layout/`

**Description:** As a developer, I want a `Container` struct that wraps content to an exact Width x Height bounding box, truncating overflow and padding underflow, so every column renderer produces dimension-safe output.

**Acceptance Criteria:**
- [ ] New file `tui/layout/container.go` created
- [ ] `Container` struct with `Width int` and `Height int` fields
- [ ] `Render(content string) string` method that: splits content into lines, truncates to Height lines, pads to Height lines if fewer, pads/truncates each line to Width using `PadToWidth`
- [ ] Adds a scroll indicator (e.g. `"↓ More..."`) on the last visible line when content is truncated
- [ ] Typecheck/vet passes with `CGO_ENABLED=0`

### US-003: Extract responsive column layout into `tui/layout/columns.go`

**Description:** As a developer, I want the responsive multi-column layout logic (breakpoints, width calculations, column joining with borders) in a dedicated file so the View method is concise and the layout rules are easy to find.

**Acceptance Criteria:**
- [ ] New file `tui/layout/columns.go` created
- [ ] Constants `MinTerminalWidth` (80), `Col3HideThreshold` (90), `Col3MinWidth` (18) exported from `layout` package
- [ ] New function `layout.ComputeColumnWidths(termWidth int) (col1, col2, col3 int, showCol3 bool)` — encapsulates the responsive width math
- [ ] New function `layout.JoinColumns(columns []string, widths []int, height int) string` — takes rendered column strings, normalizes lines, pads to width, joins with purple `│` border
- [ ] `tui/tui.go` View method calls `layout.ComputeColumnWidths` and `layout.JoinColumns` instead of inline logic
- [ ] Rendered output is byte-identical to before for any given terminal width
- [ ] Typecheck/vet passes with `CGO_ENABLED=0`

### US-004: Extract `renderColumn1` to use `Container`

**Description:** As a developer, I want `renderColumn1` to wrap its output in a `layout.Container` so it guarantees exact dimensions and shows a scroll indicator when content overflows.

**Acceptance Criteria:**
- [ ] `renderColumn1` wraps its return value in `layout.Container{Width: width, Height: height}.Render(...)`
- [ ] The manual `normalizeLines` call for column 1 in the View method is no longer needed (Container handles it)
- [ ] Scroll indicator appears when column 1 content exceeds available height
- [ ] Visual output is identical when content fits; adds `↓ More...` only when truncated
- [ ] Typecheck/vet passes with `CGO_ENABLED=0`

### US-005: Extract `renderColumn2` and `renderColumn3` to use `Container`

**Description:** As a developer, I want columns 2 and 3 to also use `layout.Container` for consistent height/width guarantees.

**Acceptance Criteria:**
- [ ] `renderColumn2` wraps its return value in `layout.Container{Width: width, Height: height}.Render(...)`
- [ ] `renderColumn3` wraps its return value in `layout.Container{Width: width, Height: height}.Render(...)`
- [ ] The manual `normalizeLines` calls for columns 2 and 3 in the View method are no longer needed
- [ ] `JoinColumns` no longer needs to call `NormalizeLines` or `PadToWidth` internally if each column is already containerized (decide during implementation — either approach is fine as long as output matches)
- [ ] Typecheck/vet passes with `CGO_ENABLED=0`

### US-006: Move `renderColumn1/2/3` and `formatStepSize` out of `tui.go`

**Description:** As a developer, I want the column render methods and their helpers in a separate file so `tui.go` stays focused on Model, Update, and top-level View orchestration.

**Acceptance Criteria:**
- [ ] New file `tui/columns.go` (same `tui` package, not `layout`) containing `renderColumn1`, `renderColumn2`, `renderColumn3`, and `formatStepSize`
- [ ] These remain methods on `*Model` (they access model state) — only the file changes, not the package
- [ ] `tui/tui.go` no longer contains any `renderColumn*` or `formatStepSize` functions
- [ ] Typecheck/vet passes with `CGO_ENABLED=0`

### US-007: Write `TUI-ARCHITECTURE.md`

**Description:** As a contributor, I want a single document that explains how the TUI layer is structured so I can understand the rendering pipeline, component API contracts, and layout system without reading every file.

**Acceptance Criteria:**
- [ ] File created at `tui/TUI-ARCHITECTURE.md`
- [ ] Covers the following sections:
  - **Overview** — purpose of the TUI layer, key dependencies (Bubble Tea, Lip Gloss, huh)
  - **Directory structure** — `tui/tui.go`, `tui/columns.go`, `tui/components/`, `tui/forms/`, `tui/styles/`, `tui/layout/`
  - **Rendering pipeline** — how `View()` orchestrates status bar → columns → timeline → command input
  - **Layout system** — `tui/layout/` package: `Container`, `ComputeColumnWidths`, `JoinColumns`, `PadToWidth`, `NormalizeLines`; responsive breakpoints (mini player < 80, 2-col 80-89, 3-col 90+)
  - **Component contracts** — each component's public function signature, state struct, and rendering guarantees (fixed height, full-width, overlay, etc.)
  - **Forms** — how huh forms integrate (init, delegate messages, check state)
  - **Styles** — color palette in `tui/styles/styles.go`
- [ ] No invented details — every statement matches the actual code
- [ ] Concise — aim for ~150-250 lines of markdown

## Functional Requirements

- FR-1: A new Go package `tui/layout` must be created with files `helpers.go`, `container.go`, and `columns.go`
- FR-2: `layout.PadToWidth` must use `ansi.Truncate` from `charmbracelet/x/ansi` for ANSI+grapheme-aware truncation (same as current implementation)
- FR-3: `layout.NormalizeLines` must pad or truncate a `[]string` to exactly the given height (same as current implementation)
- FR-4: `layout.Container{Width, Height}.Render(content)` must produce output that is exactly `Height` lines, each exactly `Width` visual columns
- FR-5: `layout.Container` must append a scroll indicator (`↓ More...`) on the last line when content is truncated
- FR-6: `layout.ComputeColumnWidths(termWidth)` must return column widths matching the current breakpoint logic: equal thirds at >=120, min-width col3 at 90-119, 2-col at 80-89
- FR-7: `layout.JoinColumns` must join column content with a purple `│` separator, matching current rendering
- FR-8: Column render methods (`renderColumn1/2/3`) must remain on `*Model` in the `tui` package (new file `tui/columns.go`), not in `tui/layout`
- FR-9: `TUI-ARCHITECTURE.md` must be placed at `tui/TUI-ARCHITECTURE.md`
- FR-10: All existing functionality must be preserved — this is a pure refactor with no behavior changes (except the addition of scroll indicators via Container)

## Non-Goals

- No new UI features, controls, or keybindings
- No changes to the Bubble Tea Model struct or Update logic
- No changes to component rendering logic (NotesList, Timeline, StatusBar, etc.)
- No changes to the forms package or styles package
- No performance optimization work
- No test suite creation (though existing tests must still pass)

## Technical Considerations

- The `layout` package must not import `tui` (avoid circular dependencies) — it can import `tui/styles` and third-party libraries (`lipgloss`, `charmbracelet/x/ansi`)
- `renderColumn1/2/3` access `m.statusBar`, `m.notesList`, `m.statsView` etc. — they must stay as methods on `*Model` in the `tui` package
- `formatStepSize` is a pure function but only used by `renderColumn1` — colocate in `tui/columns.go`
- Use `CGO_ENABLED=0` for all `go build` / `go vet` commands (modernc.org/sqlite requirement)

## Success Metrics

- `tui/tui.go` reduced by ~170 lines (from ~2173 to ~2000)
- New `tui/layout/` package with 3 files totaling ~150-200 lines
- New `tui/columns.go` with ~120 lines
- `TUI-ARCHITECTURE.md` at ~150-250 lines covering all TUI layer sections
- `CGO_ENABLED=0 go vet ./...` passes
- `CGO_ENABLED=0 go build ./...` passes
- Visual output identical to current (verified manually)

## Open Questions

- Should `Container` support horizontal scroll indicators for very narrow columns, or only vertical overflow?
- Should `TUI-ARCHITECTURE.md` include a mermaid diagram of the rendering pipeline, or keep it text-only?
