# Self Ledger

A local-first personal finance desktop app. Track expenses, revenues, and account balances. No cloud, no accounts, no internet. All data lives in a SQLite database on your machine.

## Features

- Log expenses and revenues with description, category, date, and account
- Manage multiple accounts (checking, savings, investments, etc.)
- Record transfers between accounts
- View spending and revenue trends via day-by-day line charts
- Expense breakdown by category
- Period comparison -> see how current spending compares to the previous equivalent period
- Export all data to JSON or CSV

## Requirements

- [Rust](https://rustup.rs/)
- [Bun](https://bun.sh/)
- [Tauri prerequisites](https://tauri.app/start/prerequisites/) for your OS (WebView, build tools)

## Running in development

```bash
bun install
bunx tauri dev
```

This starts the Vite dev server and opens the Tauri window with hot reload.

## Building

```bash
bun run tauri build
```

The installable bundle is output to `src-tauri/target/release/bundle/`.

## Data

The database is stored at:

| OS      | Path |
|---------|------|
| Linux   | `~/.config/self-ledger/self_ledger.db` |
| macOS   | `~/Library/Application Support/self-ledger/self_ledger.db` |
| Windows | `%APPDATA%\self-ledger\self_ledger.db` |

## Stack

- **Frontend** — SvelteKit + Svelte 5 (runes), Chart.js
- **Backend** — Tauri 2 (Rust), rusqlite (SQLite)
- **Package manager** — Bun
