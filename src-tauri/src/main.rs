#![cfg_attr(not(debug_assertions), windows_subsystem = "windows")]

mod commands;
mod db;

use rusqlite::Connection;
use tauri::Manager;

fn main() {
    tauri::Builder::default()
        .setup(|app| {
            let db_path = app.path().app_data_dir()?.join("self_ledger.db");
            let conn = Connection::open(&db_path).expect("failed to open db");
            self_ledger_lib::db::init(&conn).expect("failed to init db");
            app.manage(commands::DbConn(std::sync::Mutex::new(conn)));
            Ok(())
        })
        .plugin(tauri_plugin_dialog::init())
        .plugin(tauri_plugin_fs::init())
        .invoke_handler(tauri::generate_handler![
            commands::get_transactions,
            commands::add_transaction,
            commands::delete_transaction,
            commands::export_json,
            commands::export_csv,
        ])
        .run(tauri::generate_context!())
        .expect("error running tauri app");
}
