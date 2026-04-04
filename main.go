package main

import (
	"context"
	"embed"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/build
var assets embed.FS

func main() {
	dbPath, err := resolveDBPath()
	if err != nil {
		log.Fatal("cannot resolve db path:", err)
	}

	app, err := NewApp(dbPath)
	if err != nil {
		log.Fatal("cannot init app:", err)
	}
	defer app.Close()

	err = wails.Run(&options.App{
		Title:                    "self-ledger",
		Width:                    1900,
		Height:                   980,
		MaxWidth:                 1840,
		EnableDefaultContextMenu: true,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup: app.startup,
		OnBeforeClose: func(ctx context.Context) bool {
			if app.bm == nil {
				return false
			}
			cfg, err := app.bm.LoadConfig()
			if err != nil || cfg.Repo == "" {
				return false
			}
			done := make(chan error, 1)
			go func() {
				data, err := app.db.ExportJSON()
				if err == nil {
					err = app.bm.BackupNow(data)
				}
				done <- err
			}()
			select {
			case err := <-done:
				if err != nil {
					log.Println("auto-backup on close failed:", err)
				}
			case <-time.After(10 * time.Second):
				log.Println("auto-backup on close timed out")
			}
			return false // false = allow close to proceed
		},
		Bind: []interface{}{
			app,
		},
	})
	if err != nil {
		log.Fatal(err)
	}
}

// resolveDBPath returns ~/.config/self-ledger/self_ledger.db, creating the
// directory if needed.
func resolveDBPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(configDir, "self-ledger")
	if err := os.MkdirAll(dir, 0750); err != nil {
		return "", err
	}
	return filepath.Join(dir, "self_ledger.db"), nil
}
