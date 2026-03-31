# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What This Is

Self-Ledger is a local-first personal finance desktop app built with **Tauri 2 (Rust) + SvelteKit (TypeScript/Svelte 5)**. All data is stored in SQLite at `~/.config/self-ledger/self_ledger.db`. No network requests — everything runs locally.

## Commands

Uses **bun** as the package manager (not npm).

```bash
bun run dev          # Start Tauri app with hot reload (also starts Vite on :1420)
bun run build        # Build SvelteKit frontend
bun run check        # Run svelte-check type checking
bun run check:watch  # Type checking in watch mode
bun run tauri build  # Build final desktop app bundle
```

There are no test suites currently.

To develop, run `bun run dev` — this starts both the Vite dev server and the Tauri window.

## Architecture

```
Frontend (SvelteKit/Svelte 5)  ──invoke()──►  Backend (Tauri/Rust)  ──►  SQLite
src/routes/+page.svelte                        src-tauri/src/
                                               ├── main.rs       (app init, DB setup, command registration)
                                               ├── commands.rs   (IPC handlers)
                                               └── db.rs         (SQLite operations, Transaction struct)
```

### Frontend (`src/routes/+page.svelte`)

This is the entire UI — a single 1200-line Svelte 5 file. Key patterns:
- Uses Svelte 5 **runes**: `$state`, `$derived`, `$derived.by()`, `$effect`
- Calls Rust via `invoke("command_name", { args })` from `@tauri-apps/api/core`
- Chart.js instances are created/destroyed in `$effect` blocks (watch `filtered` records)
- SSR is disabled (`+layout.ts` sets `ssr = false`, `prerender = true`)

### Backend

**`main.rs`** — Initializes Tauri, opens SQLite at `app_data_dir()`, calls `db::init()`, registers the 5 commands.

**`commands.rs`** — Thin IPC handlers that delegate to `db.rs`: `get_transactions`, `add_transaction`, `delete_transaction`, `export_json`, `export_csv`.

**`db.rs`** — All database logic. The `Transaction` struct maps to:
```sql
CREATE TABLE transactions (
    id    INTEGER PRIMARY KEY AUTOINCREMENT,
    type  TEXT NOT NULL,   -- "expense" or "revenue"
    desc  TEXT NOT NULL,
    cat   TEXT NOT NULL,
    val   REAL NOT NULL,
    date  TEXT NOT NULL    -- "YYYY-MM-DD"
)
```

### Data Flow

1. App starts → Tauri opens SQLite → frontend mounts → calls `setPeriod("3m")` → loads last 3 months
2. Add transaction: form → `addEntry()` validates → `invoke("add_transaction")` → DB insert → reload
3. Charts: `$effect` watches `filtered` (derived from `records`) → destroys old Chart.js instances → creates new ones
4. Export: `invoke("export_json"/"export_csv")` → file written via Tauri dialog

### Key Rust State

`DbConn` (a `Mutex<Connection>`) is stored in Tauri's managed state and accessed in every command handler via `State<'_, DbConn>`.
