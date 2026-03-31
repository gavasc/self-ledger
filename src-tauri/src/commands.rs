use crate::db;
use crate::db::{Transaction, Account, Transfer};

use rusqlite::Connection;
use std::sync::Mutex;
use tauri::State;

pub struct DbConn(pub Mutex<Connection>);

// ── Transactions ──────────────────────────────────────────────────────────────

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
pub fn update_transaction(
    state: State<DbConn>,
    transaction: Transaction,
) -> Result<(), String> {
    let conn = state.0.lock().unwrap();
    db::update(&conn, &transaction).map_err(|e| e.to_string())
}

#[tauri::command]
pub fn delete_transaction(
    state: State<DbConn>,
    id: i64,
) -> Result<(), String> {
    let conn = state.0.lock().unwrap();
    db::delete(&conn, id).map_err(|e| e.to_string())
}

#[tauri::command]
pub fn export_json(state: State<DbConn>) -> Result<String, String> {
    let conn = state.0.lock().unwrap();
    db::export_json(&conn).map_err(|e| e.to_string())
}

#[tauri::command]
pub fn export_csv(state: State<DbConn>) -> Result<String, String> {
    let conn = state.0.lock().unwrap();
    db::export_csv(&conn).map_err(|e| e.to_string())
}

// ── Accounts ──────────────────────────────────────────────────────────────────

#[tauri::command]
pub fn get_accounts(state: State<DbConn>) -> Result<Vec<Account>, String> {
    let conn = state.0.lock().unwrap();
    db::get_accounts(&conn).map_err(|e| e.to_string())
}

#[tauri::command]
pub fn add_account(
    state: State<DbConn>,
    account: Account,
) -> Result<i64, String> {
    let conn = state.0.lock().unwrap();
    db::insert_account(&conn, &account).map_err(|e| e.to_string())
}

#[tauri::command]
pub fn delete_account(
    state: State<DbConn>,
    id: i64,
) -> Result<(), String> {
    let conn = state.0.lock().unwrap();
    db::delete_account(&conn, id).map_err(|e| e.to_string())
}

#[tauri::command]
pub fn get_account_balances(state: State<DbConn>) -> Result<Vec<db::AccountBalance>, String> {
    let conn = state.0.lock().unwrap();
    db::get_account_balances(&conn).map_err(|e| e.to_string())
}

// ── Transfers ─────────────────────────────────────────────────────────────────

#[tauri::command]
pub fn get_transfers(state: State<DbConn>) -> Result<Vec<Transfer>, String> {
    let conn = state.0.lock().unwrap();
    db::get_transfers(&conn).map_err(|e| e.to_string())
}

#[tauri::command]
pub fn add_transfer(
    state: State<DbConn>,
    transfer: Transfer,
) -> Result<i64, String> {
    let conn = state.0.lock().unwrap();
    db::insert_transfer(&conn, &transfer).map_err(|e| e.to_string())
}

#[tauri::command]
pub fn delete_transfer(
    state: State<DbConn>,
    id: i64,
) -> Result<(), String> {
    let conn = state.0.lock().unwrap();
    db::delete_transfer(&conn, id).map_err(|e| e.to_string())
}
