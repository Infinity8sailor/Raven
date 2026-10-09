# 🐦‍⬛ Raven

> **Hell-minimal, distraction-free, local-first todo and clipboard application.**

Raven is an ultra-lightweight, single-binary personal organizer built in pure Go and Alpine.js. It pairs structured todo lists with an encrypted, quick-access key-value clipboard. No telemetry, no cloud dependencies, no npm bloat.

---

## ✨ Features

- **Local-First & Offline**: All data resides locally in clean, human-readable JSON (`data/tasks.json`) with atomic temp-file writes and rolling backup rotations.
- **Dual List Architecture**:
  - **Standard Lists**: Tasks, completion checkboxes, manual reordering (`↑`/`↓`), sprint date ranges, relative time badges, auto-collapse, and inline editing.
  - **Global Clipboard**: Key-value snippet store with one-click copying (`⎘`), sensitive value masking (`••••••••`), and per-item reveal toggles (`ø`/`⚆`).
- **Seamless Inline Editing**: Double-click any task title to trigger an auto-expanding inline textarea with zero scrollbars, instant `Enter` key persistence, and unpersisted draft highlights.
- **Privacy-Centric Search**:
  - Instant text filter across standard lists and tasks.
  - Searches clipboard entries strictly against **keys** (value text is never searched to prevent accidental secret leakage).
  - Sensitive clipboard values remain masked in search results with no unhide option.
- **Fluid Masonry Layout**: Column-based grid that dynamically stacks cards tightly with zero vertical dead space across 1 to 5 columns on desktop and 4K displays.
- **Customization Engine**:
  - **5 Palettes**: Nero (Obsidian), Chlorophyll (forest green), Infrared (dark red), Ultraviolet (deep purple), and Sepia (warm amber) + Randomizer.
  - **Typography Options**: Switch between Classic Serif (Georgia) and Cosmic Sans (Comic Sans) in Global Settings (`⛭`).
- **Subtle Visual Feedback**: Pinned lists display a vector pin icon (`📌`) directly beside the title; action menus fold away behind a clean 3-dot (`⋮`) trigger.

---

## 🏗️ Architecture & Tech Stack

```
Raven/
├── main.go               # Application entrypoint & HTTP server bootstrap
├── config.json           # Port & directory configuration
├── data.example.json     # Initial default schema template
├── server/               # Pure Go backend (zero external dependencies)
│   ├── config.go         # Configuration loader
│   ├── models.go         # Domain models, request types, validation
│   ├── storage.go        # Thread-safe atomic file storage & backups
│   ├── migration.go      # Schema migration pipeline
│   ├── router.go         # Standard library HTTP mux (Go 1.22+ routing)
│   ├── lists.go          # List CRUD, feature flags, reordering
│   ├── items.go          # Item CRUD, completion toggles, reordering
│   ├── search.go         # Privacy-preserving search query handler
│   ├── settings.go       # Theme, backup mode, and health check handlers
│   └── *_test.go         # Comprehensive unit & integration test suites
├── web/                  # Minimalist web frontend (no node_modules)
│   ├── index.html        # Single-page UI with Alpine.js
│   ├── app.js            # Reactive controller & API client
│   ├── styles.css        # Responsive CSS layout & dark mode styles
│   ├── themes.css        # CSS variable palettes
│   └── vendor/
│       └── alpine.min.js # Alpine.js v3 runtime
└── scripts/              # OS startup service installers
    ├── install-win.ps1   # Windows Scheduled Task installer
    ├── install-linux.sh  # Linux systemd user service installer
    └── install-mac.sh    # macOS launchd daemon installer
```

---

## 🚀 Quickstart

### Prerequisites
- [Go 1.22+](https://go.dev/dl/) installed.

### 1. Build the Binary
```bash
# Build the binary
go build -o raven.exe .
```

### 2. Run Raven
```bash
./raven.exe
```
By default, Raven starts on `http://localhost:9471`.

### 3. Open in Browser
Navigate to:
```
http://localhost:9471
```

---

## ⚙️ Configuration (`config.json`)

```json
{
  "port": 9471,
  "data_dir": "data",
  "web_dir": "web"
}
```

- `port`: Local HTTP port to bind to (default: `9471`).
- `data_dir`: Directory where `tasks.json` and rolling backups are stored.
- `web_dir`: Directory containing static HTML, CSS, and JS assets.

---

## 📡 REST API Reference

| Method | Endpoint | Description |
|---|---|---|
| `GET` | `/api/lists` | Fetch all lists (filter with `?show_archived=true`) |
| `POST` | `/api/lists` | Create a list (`{"type": "standard"\|"clipboard", "title": "..."}`) |
| `PUT` | `/api/lists/:id` | Update list metadata (`title`, `pinned`, `archived`, sprint dates) |
| `DELETE` | `/api/lists/:id` | Delete list (requires `?confirm=true`) |
| `PATCH` | `/api/lists/:id/features` | Toggle features (`checkboxes`, `reorder`, `sensitive`, etc.) |
| `POST` | `/api/lists/reorder` | Reorder list position (`{"id": "...", "direction": "up"\|"down"}`) |
| `POST` | `/api/lists/:id/items` | Add item or key-value pair to a list |
| `PUT` | `/api/lists/:id/items/:itemId` | Update item title, key, value, or status |
| `DELETE` | `/api/lists/:id/items/:itemId` | Delete item (requires `?confirm=true`) |
| `PATCH` | `/api/lists/:id/items/:itemId/status` | Toggle item completion status (`todo` ↔ `done`) |
| `POST` | `/api/lists/:id/items/reorder` | Reorder item inside a list |
| `GET` | `/api/search?q={query}` | Global search across lists, tasks, and clipboard keys |
| `GET` | `/api/settings` | Retrieve active application settings |
| `PUT` | `/api/settings` | Update settings (`theme`, `backup_mode`, `show_archived`) |
| `POST` | `/api/backup` | Trigger an immediate rolling JSON backup |
| `GET` | `/api/health` | Uptime diagnostic and system status |

---

## 🧪 Testing

Raven includes a comprehensive test suite covering schema validation, file atomicity, backup rotation, concurrent access, API routes, and privacy-preserving search.

Run all tests:
```bash
go test -v ./...
```

Run tests with race detection:
```bash
go test -race ./...
```

---

## 📦 Autostart / Background Service Installation

- **Windows**: Run `powershell -ExecutionPolicy Bypass -File scripts/install-win.ps1` to register a logon task.
- **Linux**: Run `bash scripts/install-linux.sh` to install a systemd user unit.
- **macOS**: Run `bash scripts/install-mac.sh` to create a `launchd` plist.

---

## 📄 License

MIT. Free to use, adapt, and run offline anywhere.
