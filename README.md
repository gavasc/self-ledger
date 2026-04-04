# Self Ledger

A local-first personal finance desktop app. Track expenses, revenues, and account balances. No cloud, no accounts, no internet. All data lives in a SQLite database on your machine.

## Features

- Log expenses and revenues with description, category, date, and account
- Manage multiple accounts (checking, savings, investments, etc.)
- Record transfers between accounts
- Split a purchase into N monthly transactions
- View spending and revenue trends via day-by-day line charts
- Expense breakdown by category
- See how current spending compares to the previous equivalent period
- Period notes for expenses and revenues
- Export all data to JSON or CSV
- Backup your data to a git repo as JSON

## Requirements

- [Go](https://go.dev/dl/) 1.21+
- [Wails v2](https://wails.io/docs/gettingstarted/installation) (`go install github.com/wailsapp/wails/v2/cmd/wails@latest`)
- [Bun](https://bun.sh/)
- Linux: `gtk3-devel` and `webkit2gtk4.1-devel` (or equivalent for your distro)

## Running in development

```bash
cd frontend && bun install && cd ..
wails dev -tags webkit2_4_1        # Linux with webkit2gtk 4.1
wails dev                          # macOS / Windows / Linux with webkit2gtk 4.0
```

This starts the Vite dev server on `:5173` and opens the Wails window with hot reload.

## Building

```bash
wails build -tags webkit2_4_1     # Linux with webkit2gtk 4.1
wails build                        # macOS / Windows / Linux with webkit2gtk 4.0
```

The binary is output to `build/bin/self-ledger`.

## Packaging (Linux)

`.deb` and `.rpm` packages are produced by [nfpm](https://nfpm.goreleaser.com/). The packages install:

| File | Destination |
|------|-------------|
| `build/bin/self-ledger` | `/usr/local/bin/self-ledger` |
| `frontend/static/favicon.png` | `/usr/share/pixmaps/self-ledger.png` |
| `self-ledger.desktop` | `/usr/share/applications/self-ledger.desktop` |

```bash
go install github.com/goreleaser/nfpm/v2/cmd/nfpm@latest
VERSION=1.0.0 nfpm package --packager deb --target build/packages/
VERSION=1.0.0 nfpm package --packager rpm --target build/packages/
```

## Data

The database is stored at:

| OS      | Path |
|---------|------|
| Linux   | `~/.config/self-ledger/self_ledger.db` |
| macOS   | `~/Library/Application Support/self-ledger/self_ledger.db` |
| Windows | `%APPDATA%\self-ledger\self_ledger.db` |

## Stack

- **Frontend** — SvelteKit + Svelte 5 (runes), Chart.js, Bun
- **Backend** — Wails v2 (Go), modernc/sqlite (pure-Go SQLite, no CGO)
