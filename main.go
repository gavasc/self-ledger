package main

import (
	"embed"
	"log"
	"os"
	"path/filepath"

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
		Width:                    1800,
		Height:                   950,
		EnableDefaultContextMenu: true,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup: app.startup,
		Bind: []interface{}{
			app,
		},
	})
	if err != nil {
		log.Fatal(err)
	}
}

// resolveDBPath returns ~/.config/self-ledger/self_ledger.db, creating the
// directory if needed. Matches Tauri's app_data_dir() on Linux.
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
