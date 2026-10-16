# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.0.1] - 2026-10-17

### Added
- Explicit "Sensitive" Checkbox Tag in Clipboard Quick-Add: added a clean `[ ] Sensitive` tag in the global clipboard quick-add bar with lock indicator (`🔒`) when checked, ensuring entries can be marked sensitive directly upon creation.
- Per-item Sensitive Masking & Eye Toggles: entries tagged as sensitive are masked with exact-length bullet characters (`•`), with reveal/hide eye toggles (`ø` / `⚆`) in the global clipboard modal and permanent masking in search results.
- Automatic Legacy Migration for Sensitive Lists: automatically migrates any clipboard items residing in lists that previously had `features.sensitive = true` to `item.sensitive = true`.
- Strict Non-Empty Validation: enforced non-empty key and value validation across frontend quick-add and backend `ValidateClipboardKey` / `ValidateClipboardValue`.

### Changed
- Unified Clipboard Naming: renamed clipboard lists to "Clipboard" (removing legacy "Dev Endpoints" and "Quick Clipboard" names) and eliminated the redundant subheader inside the clipboard overlay modal.

### Removed
- "Hell No Drama" Cleanup: pruned over-engineered project management layers:
  - Removed Sprint Dates: dropped sprint date range pickers (`sprint_start`, `sprint_end`) and header sprint date pills.
  - Removed Focus Profile Circles & Multi-Tenancy: removed profile circles, profile dropdown menus, list profile tags, and profile filtering from navigation.
  - Removed List-Level Sensitive Mode (`⚿` / `◌`): streamlined privacy strictly to per-item sensitive tagging.
- Auto-check and Overwrite in Windows Installer: enhanced [`scripts/install-win.ps1`](scripts/install-win.ps1) to automatically detect pre-existing `RavenApp` keys in `HKCU\Software\Microsoft\Windows\CurrentVersion\Run` and the Windows Startup shortcut (`Raven.lnk`), inspect their targets, and overwrite them with `-Force` pointing to the current executable path, keeping the Go server binary pure and decoupled from OS registry manipulation.
- Customizable Port Configuration: added `-port <number>` CLI flag to `main.go`, support for `RAVEN_PORT` and `PORT` environment variables in [`server/config.go`](server/config.go), and `-Port` parameter in [`scripts/install-win.ps1`](scripts/install-win.ps1) with full unit test suite in [`server/config_test.go`](server/config_test.go), ensuring the HTTP port is never hardcoded.
- Automated Git Build Scripts with Version Check: created [`scripts/build.ps1`](scripts/build.ps1) and [`scripts/build.sh`](scripts/build.sh) which verify git tags (`git tag -l $Version`) and abort with an error on duplicate releases, execute the full test suite (`go test -v ./...`), compile binary with `-ldflags` (`Version`, `GitCommit`, `BuildDate`), and validate binary output via `-version`.
- Compact 15-copy tiered backup retention: optimized snapshot cap to 15 files max (<250 KB total disk footprint) preserving 5 recent live edits (debounced), 7 daily snapshots for the past week, and 2 weekly snapshots for the past 14 days without sacrificing historical rollback granularity.
- Official MIT License: added [`LICENSE`](LICENSE) granting free, open-source usage, modification, and offline distribution.
- Focus Profile Circles: integrated compact circular profile badges directly into the main navigation bar (`All`, `W` with red background for work, `P` with blue background for personal, `AI` with purple background, etc.), displaying up to 3 direct circles with a dropdown overlay menu for additional profiles, removing the second navigation row.
- "Copy All Items" action in 3-dot list options menu: copies all items in the list to the clipboard in one click with Markdown completion checkboxes (`[ ]` / `[x]`) for tasks or key-value format for clipboard lists.
- Enhanced Date/Day/Week header display: positioned live clock directly next to the Raven logo in a single horizontal row, aligned along the bottom edge of the navigation bar with Day (`Mon`, `Tue`), date, ISO calendar week (`W41`), and local time.
- Default seed lists on fresh install: automatically initializes `data/tasks.json` on first run with rich interactive demonstration lists ("Welcome to Raven", "Sprint 1: Explore Features", and "Quick Clipboard") that showcase all key app features (double-click inline editing, completion checkboxes, reordering, sprint dates, 3-dot configure panel, themes, search, and masked clipboard items).
- Reliable non-admin Windows startup installer: updated [`scripts/install-win.ps1`](scripts/install-win.ps1) to automatically compile the latest binary with Go (`go build -ldflags="-s -w" -o raven.exe main.go`) and configure user-level auto-start via Windows Startup folder shortcut (`Raven.lnk`) and HKCU Run registry without requiring Administrator privileges.
- Dedicated Windows uninstaller script ([`scripts/uninstall-win.ps1`](scripts/uninstall-win.ps1)) to cleanly remove startup shortcuts and registry entries.
- Executable directory fallback in [`main.go`](main.go) ensuring Raven seamlessly resolves `web/` assets and `config.json` even when launched with an arbitrary working directory (such as Windows logon).
- Comprehensive project documentation in [`README.md`](README.md) including architecture diagrams, API reference table, quickstart build guides, and background service autostart setup.
- Expanded backend test suite with 43 passing tests covering list pinning, sprint date updates, feature flag mutations, list reordering, privacy-preserving clipboard search, rolling JSON backup rotation, and default seed lists.
- Pinned list indicator: rendered a subtle vector pin icon beside list titles for all pinned lists with full `aria-label="Pinned list"` and `title="Pinned list"` hover tooltips.
- Comprehensive accessibility and hover identification: added `aria-label` and `title` tooltip attributes across all interactive icons, buttons, inputs, toggle switches, and menus (header controls, global clipboard actions, 3-dots menu triggers, configure options, reordering arrows, checkboxes, and delete triggers) with `aria-hidden="true"` on decorative vector icons.
- Font selection option in Global Settings (`⛭`), allowing users to toggle between Classic Serif (Georgia) and Cosmic Sans (Comic Sans), seamlessly updating body typography, list titles, item descriptions, and inputs with local persistence.
- Comprehensive List Configure panel featuring custom minimal iOS-style toggle switches, in-place list title renaming, sprint date range pickers (start & end), list grid reordering (Move Left / Right), and single-click clear completed tasks.
- Double-click inline item editing with multi-line textarea, Enter-to-save persistence, Escape cancellation, and subtle background highlight for unpersisted/dirty drafts.
- 3-dot (`⋮`) overflow action menu on list cards with unified, solid-filled monochrome vector icons.
- True modal overlay backdrop for the Global Clipboard panel with keyboard and backdrop-click dismissal.

### Changed
- Resolved 3-dot dropdown menu clipping and z-index collision: removed overflow clipping on list cards, raised active card z-index dynamically when the dropdown menu is open, ensuring the 3-dot menu is fully visible even on empty list cards.
- Updated pin icon: replaced the vertical shape with a sleek, crisp angled pushpin (`pin-angle-fill`) both beside pinned list titles and in the 3-dot dropdown menu.
- Refined search matching: global search results now include clipboard items by matching keys only (never matching value text). Sensitive clipboard values in search results are strictly masked (`••••••••`) with no unhide action button; the per-item reveal toggle is exclusively accessible within the dedicated Global Clipboard overlay panel.
- Improved inline item editing: removed textarea scrollbars completely (`overflow: hidden`), added reactive auto-resizing (height dynamically increases as text wraps and contracts as text is removed), and ensured pressing Enter immediately saves changes without inserting unwanted newlines.
- Restored original subtle, minimal 3-dot (`⋮`) menu trigger styling (1.75px radius dots, muted tone, and 0.7 base opacity), removing the bold heavy dots.
- Expanded Global Clipboard overlay modal width to 38vw (35–40% of viewport) with 500px min-width and 850px max-width, eliminating the narrow 400px constraint.
- Moved the quick-add item input to the bottom of list cards and global clipboard, with newly created items appended to the bottom and smooth auto-scrolling into view.
- Reorganized navigation bar: live date/time clock anchored directly next to the Raven logo.
- Moved theme picker dropdown to the right utility cluster.
- Scaled up typography for the Raven brand logo and list card headings.
- Muted, faded background tint (`rgba(255, 255, 255, 0.02)`) and dashed borders for archived lists.
- Widened list card columns across responsive breakpoints for a more spacious and readable layout.
- Initial release of Raven, a "hell minimal" and focused offline-first todo application.
- Standard lists with features: checkboxes, reorder, due dates, last updated badges, sprint dates, and auto-collapse.
- Clipboard lists with features: sensitive data masking and 1-click copy functionality.
- Fluid masonry grid layout that responsive adapts up to 3 columns.
- Sticky header with integrated clock (updates to the second).
- Global text search across list titles, item titles, keys, and non-sensitive clipboard values.
- Dark mode theme engine with 5 distinct palettes (Obsidian, Chlorophyll, Infrared, Ultraviolet, Sepia) and a Random selector.
- Complete Go backend on port 9471 utilizing atomic file writes and automatic JSON backup rotation.
- Robust Go test suite for API endpoints, schema validation, and storage concurrency.
