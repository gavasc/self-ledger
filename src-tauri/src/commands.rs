use crate::db;
use crate::db::Transaction;
use rusqlite::Connection;
use std::sync::Mutex;
use tauri::State;

pub struct DbConn(pub Mutex<Connection>);

#[tauri::command]
pub fn get_transactions(
    state: State<DbConn>,
    from: String,
    to: String,
) -> Result<Vec<Transaction>, String> {
    let conn = state.0.lock().unwrap();
    db::query(&conn, &from, &to).map_err(|e| e.to_string())
}

#[tauri::command]
pub fn add_transaction(
    state: State<DbConn>,
    transaction: Transaction,
) -> Result<i64, String> {
    let conn = state.0.lock().unwrap();
    db::insert(&conn, &transaction).map_err(|e| e.to_string())
}

#[tauri::command]
pub fn delete_transaction(
    state: State<DbConn>,
    id: i64,
) -> Result<(), String> {
    let conn = state.0.lock().unwrap();
    db::delete(&conn, id).map_err(|e| e.to_string())
}
