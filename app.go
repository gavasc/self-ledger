package main

import (
	"context"
	"fmt"
	"os"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App is the main application struct. All exported methods are bound to the
// frontend as TypeScript functions via Wails' generated bindings.
type App struct {
	ctx context.Context
	db  *DB
	bm  *BackupManager
}

func NewApp(dbPath string) (*App, error) {
	db, err := NewDB(dbPath)
	if err != nil {
		return nil, err
	}
	bm, err := NewBackupManager()
	if err != nil {
		// Non-fatal: app runs normally without backup configured.
		bm = nil
	}
	return &App{db: db, bm: bm}, nil
}

// startup is called by Wails after the app window is ready.
// The context is needed for runtime calls (e.g. SaveFileDialog).
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) Close() {
	if a.db != nil {
		a.db.Close()
	}
}

// ── Transactions ──────────────────────────────────────────────────────────────

func (a *App) GetTransactions(from, to string) ([]Transaction, error) {
	return a.db.QueryTransactions(from, to)
}

func (a *App) AddTransaction(t Transaction) (int64, error) {
	return a.db.InsertTransaction(t)
}

func (a *App) UpdateTransaction(t Transaction) error {
	return a.db.UpdateTransaction(t)
}

func (a *App) DeleteTransaction(id int64) error {
	return a.db.DeleteTransaction(id)
}

// ── Accounts ──────────────────────────────────────────────────────────────────

func (a *App) GetAccounts() ([]Account, error) {
	return a.db.GetAccounts()
}

func (a *App) AddAccount(acc Account) (int64, error) {
	return a.db.InsertAccount(acc)
}

func (a *App) DeleteAccount(id int64) error {
	return a.db.DeleteAccount(id)
}

func (a *App) GetAccountBalances() ([]AccountBalance, error) {
	return a.db.GetAccountBalances()
}

// ── Transfers ─────────────────────────────────────────────────────────────────

func (a *App) GetTransfers() ([]Transfer, error) {
	return a.db.GetTransfers()
}

func (a *App) AddTransfer(t Transfer) (int64, error) {
	return a.db.InsertTransfer(t)
}

func (a *App) DeleteTransfer(id int64) error {
	return a.db.DeleteTransfer(id)
}

// ── Notes ─────────────────────────────────────────────────────────────────────

func (a *App) GetNote(section, from, to string) (string, error) {
	return a.db.GetNote(section, from, to)
}

func (a *App) SaveNote(section, from, to, content string) error {
	return a.db.UpsertNote(section, from, to, content)
}

// ── Installments ──────────────────────────────────────────────────────────────

func (a *App) AddInstallment(inst Installment) (int64, error) {
	return a.db.InsertInstallment(inst)
}

func (a *App) GetInstallments() ([]Installment, error) {
	return a.db.GetInstallments()
}

func (a *App) DeleteInstallment(id int64) error {
	return a.db.DeleteInstallment(id)
}

// ── Export ────────────────────────────────────────────────────────────────────

// ExportJSON opens a native save dialog and writes the full JSON export.
// Go handles both the dialog and the file write; the frontend just calls this.
func (a *App) ExportJSON() error {
	data, err := a.db.ExportJSON()
	if err != nil {
		return err
	}
	path, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "Export JSON",
		DefaultFilename: "self-ledger-export.json",
		Filters: []runtime.FileFilter{
			{DisplayName: "JSON Files (*.json)", Pattern: "*.json"},
		},
	})
	if err != nil || path == "" {
		return err // nil when user cancelled
	}
	return os.WriteFile(path, []byte(data), 0644)
}

// ── Backup ────────────────────────────────────────────────────────────────────

func (a *App) GetBackupConfig() (BackupConfig, error) {
	if a.bm == nil {
		return BackupConfig{}, nil
	}
	return a.bm.LoadConfig()
}

func (a *App) SaveBackupConfig(cfg BackupConfig) error {
	if a.bm == nil {
		return fmt.Errorf("backup manager unavailable")
	}
	return a.bm.SaveConfig(cfg)
}

// AutoRestoreIfNeeded checks on startup whether the local database is empty
// and a backup is configured. If so, it fetches and imports the remote backup
// automatically. Returns true if a restore was performed.
//
// Errors from fetching (e.g. remote not yet initialised) are treated as
// non-fatal — the app just starts with an empty DB as usual.
func (a *App) AutoRestoreIfNeeded() (bool, error) {
	if a.bm == nil {
		return false, nil
	}
	cfg, err := a.bm.LoadConfig()
	if err != nil || cfg.Repo == "" {
		return false, nil
	}
	empty, err := a.db.IsEmpty()
	if err != nil || !empty {
		return false, err
	}
	jsonData, err := a.bm.FetchBackup()
	if err != nil {
		// Remote may not have a backup yet — not an error the user needs to see.
		return false, nil
	}
	if err := a.db.ImportJSON(jsonData); err != nil {
		return false, err
	}
	return true, nil
}

func (a *App) BackupNow() error {
	if a.bm == nil {
		return fmt.Errorf("backup manager unavailable")
	}
	data, err := a.db.ExportJSON()
	if err != nil {
		return err
	}
	return a.bm.BackupNow(data)
}

func (a *App) RestoreFromBackup() error {
	if a.bm == nil {
		return fmt.Errorf("backup manager unavailable")
	}
	jsonData, err := a.bm.FetchBackup()
	if err != nil {
		return err
	}
	return a.db.ImportJSON(jsonData)
}

// ExportCSV opens a native save dialog and writes the full CSV export.
func (a *App) ExportCSV() error {
	data, err := a.db.ExportCSV()
	if err != nil {
		return err
	}
	path, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "Export CSV",
		DefaultFilename: "self-ledger-export.csv",
		Filters: []runtime.FileFilter{
			{DisplayName: "CSV Files (*.csv)", Pattern: "*.csv"},
		},
	})
	if err != nil || path == "" {
		return err
	}
	return os.WriteFile(path, []byte(data), 0644)
}
