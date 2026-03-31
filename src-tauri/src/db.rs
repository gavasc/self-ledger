use rusqlite::{Connection, Result, params};
use serde::{Deserialize, Serialize};

#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct Transaction {
    pub id: Option<i64>,
    #[serde(rename="type")]
    pub type_: String,
    pub desc: String,
    pub cat: String,
    pub val: f64,
    pub date: String,
    pub account_id: Option<i64>,
}

#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct Account {
    pub id: Option<i64>,
    pub name: String,
    pub initial_balance: f64,
}

#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct AccountBalance {
    pub id: i64,
    pub name: String,
    pub balance: f64,
}

#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct Transfer {
    pub id: Option<i64>,
    pub from_account_id: i64,
    pub to_account_id: i64,
    pub from_account_name: Option<String>,
    pub to_account_name: Option<String>,
    pub amount: f64,
    pub date: String,
    pub desc: String,
}

pub fn init(conn: &Connection) -> Result<()> {
    conn.execute_batch("
        CREATE TABLE IF NOT EXISTS transactions (
            id    INTEGER PRIMARY KEY AUTOINCREMENT,
            type  TEXT    NOT NULL,
            desc  TEXT    NOT NULL,
            cat   TEXT    NOT NULL,
            val   REAL    NOT NULL,
            date  TEXT    NOT NULL
        );
        CREATE TABLE IF NOT EXISTS accounts (
            id              INTEGER PRIMARY KEY AUTOINCREMENT,
            name            TEXT    NOT NULL UNIQUE,
            initial_balance REAL    NOT NULL DEFAULT 0
        );
        CREATE TABLE IF NOT EXISTS transfers (
            id              INTEGER PRIMARY KEY AUTOINCREMENT,
            from_account_id INTEGER NOT NULL REFERENCES accounts(id),
            to_account_id   INTEGER NOT NULL REFERENCES accounts(id),
            amount          REAL    NOT NULL,
            date            TEXT    NOT NULL,
            desc            TEXT    NOT NULL DEFAULT ''
        );
        CREATE TABLE IF NOT EXISTS notes (
            section     TEXT NOT NULL,
            period_from TEXT NOT NULL,
            period_to   TEXT NOT NULL,
            content     TEXT NOT NULL DEFAULT '',
            PRIMARY KEY (section, period_from, period_to)
        );
    ")?;
    // Migration: add account_id to transactions (ignored if column already exists)
    let _ = conn.execute(
        "ALTER TABLE transactions ADD COLUMN account_id INTEGER REFERENCES accounts(id)",
        [],
    );
    Ok(())
}

// ── Transactions ──────────────────────────────────────────────────────────────

pub fn insert(conn: &Connection, t: &Transaction) -> Result<i64> {
    conn.execute(
        "INSERT INTO transactions (type, desc, cat, val, date, account_id)
         VALUES (?1, ?2, ?3, ?4, ?5, ?6)",
        params![t.type_, t.desc, t.cat, t.val, t.date, t.account_id],
    )?;
    Ok(conn.last_insert_rowid())
}

pub fn update(conn: &Connection, t: &Transaction) -> Result<()> {
    conn.execute(
        "UPDATE transactions
         SET type = ?1, desc = ?2, cat = ?3, val = ?4, date = ?5, account_id = ?6
         WHERE id = ?7",
        params![t.type_, t.desc, t.cat, t.val, t.date, t.account_id, t.id],
    )?;
    Ok(())
}

pub fn query(conn: &Connection, from: &str, to: &str) -> Result<Vec<Transaction>> {
    let mut stmt = conn.prepare(
        "SELECT id, type, desc, cat, val, date, account_id
         FROM transactions
         WHERE date >= ?1 AND date <= ?2
         ORDER BY date DESC"
    )?;
    let rows = stmt.query_map(params![from, to], |row| {
        Ok(Transaction {
            id:         Some(row.get(0)?),
            type_:      row.get(1)?,
            desc:       row.get(2)?,
            cat:        row.get(3)?,
            val:        row.get(4)?,
            date:       row.get(5)?,
            account_id: row.get(6)?,
        })
    })?;
    rows.collect()
}

pub fn delete(conn: &Connection, id: i64) -> Result<()> {
    conn.execute("DELETE FROM transactions WHERE id = ?1", params![id])?;
    Ok(())
}

pub fn export_json(conn: &Connection) -> Result<String> {
    let mut stmt = conn.prepare(
        "SELECT id, type, desc, cat, val, date, account_id FROM transactions ORDER BY date DESC"
    )?;
    let transactions: Vec<Transaction> = stmt.query_map([], |row| {
        Ok(Transaction {
            id:         Some(row.get(0)?),
            type_:      row.get(1)?,
            desc:       row.get(2)?,
            cat:        row.get(3)?,
            val:        row.get(4)?,
            date:       row.get(5)?,
            account_id: row.get(6)?,
        })
    })?.collect::<Result<Vec<_>>>()?;

    let accounts = get_account_balances(conn)?;

    let mut tstmt = conn.prepare("
        SELECT t.id, t.from_account_id, t.to_account_id,
               fa.name, ta.name, t.amount, t.date, t.desc
        FROM transfers t
        JOIN accounts fa ON fa.id = t.from_account_id
        JOIN accounts ta ON ta.id = t.to_account_id
        ORDER BY t.date DESC, t.id DESC
    ")?;
    let transfers: Vec<Transfer> = tstmt.query_map([], |row| {
        Ok(Transfer {
            id:                Some(row.get(0)?),
            from_account_id:   row.get(1)?,
            to_account_id:     row.get(2)?,
            from_account_name: Some(row.get(3)?),
            to_account_name:   Some(row.get(4)?),
            amount:            row.get(5)?,
            date:              row.get(6)?,
            desc:              row.get(7)?,
        })
    })?.collect::<Result<Vec<_>>>()?;

    let mut nstmt = conn.prepare(
        "SELECT section, period_from, period_to, content FROM notes WHERE content != '' ORDER BY section, period_from"
    )?;
    let notes: Vec<serde_json::Value> = nstmt.query_map([], |row| {
        Ok((
            row.get::<_, String>(0)?,
            row.get::<_, String>(1)?,
            row.get::<_, String>(2)?,
            row.get::<_, String>(3)?,
        ))
    })?.collect::<Result<Vec<_>>>()?
    .into_iter()
    .map(|(section, from, to, content)| serde_json::json!({
        "section":   section,
        "date-from": from,
        "date-to":   to,
        "note":      content,
    }))
    .collect();

    let out = serde_json::json!({
        "transactions": transactions,
        "accounts":     accounts,
        "transfers":    transfers,
        "notes":        notes,
    });
    Ok(serde_json::to_string_pretty(&out).unwrap())
}

pub fn export_csv(conn: &Connection) -> Result<String> {
    let mut csv = String::new();

    // transactions
    csv.push_str("transactions\n");
    csv.push_str("id,type,desc,cat,val,date,account_id\n");
    let mut stmt = conn.prepare(
        "SELECT id, type, desc, cat, val, date, account_id FROM transactions ORDER BY date DESC"
    )?;
    for row in stmt.query_map([], |row| {
        Ok((
            row.get::<_, i64>(0)?,
            row.get::<_, String>(1)?,
            row.get::<_, String>(2)?,
            row.get::<_, String>(3)?,
            row.get::<_, f64>(4)?,
            row.get::<_, String>(5)?,
            row.get::<_, Option<i64>>(6)?,
        ))
    })? {
        let (id, type_, desc, cat, val, date, account_id) = row?;
        let acct = account_id.map(|v| v.to_string()).unwrap_or_default();
        csv.push_str(&format!("{},{},{},{},{:.2},{},{}\n", id, type_, desc, cat, val, date, acct));
    }

    // accounts
    csv.push_str("\naccounts\n");
    csv.push_str("id,name,balance\n");
    for a in get_account_balances(conn)? {
        csv.push_str(&format!("{},{},{:.2}\n", a.id, a.name, a.balance));
    }

    // transfers
    csv.push_str("\ntransfers\n");
    csv.push_str("id,from_account,to_account,amount,date,desc\n");
    let mut stmt = conn.prepare("
        SELECT t.id, fa.name, ta.name, t.amount, t.date, t.desc
        FROM transfers t
        JOIN accounts fa ON fa.id = t.from_account_id
        JOIN accounts ta ON ta.id = t.to_account_id
        ORDER BY t.date DESC, t.id DESC
    ")?;
    for row in stmt.query_map([], |row| {
        Ok((
            row.get::<_, i64>(0)?,
            row.get::<_, String>(1)?,
            row.get::<_, String>(2)?,
            row.get::<_, f64>(3)?,
            row.get::<_, String>(4)?,
            row.get::<_, String>(5)?,
        ))
    })? {
        let (id, from, to, amount, date, desc) = row?;
        csv.push_str(&format!("{},{},{},{:.2},{},{}\n", id, from, to, amount, date, desc));
    }

    // notes
    csv.push_str("\nnotes\n");
    csv.push_str("section,date-from,date-to,note\n");
    let mut nstmt = conn.prepare(
        "SELECT section, period_from, period_to, content FROM notes WHERE content != '' ORDER BY section, period_from"
    )?;
    for row in nstmt.query_map([], |row| {
        Ok((
            row.get::<_, String>(0)?,
            row.get::<_, String>(1)?,
            row.get::<_, String>(2)?,
            row.get::<_, String>(3)?,
        ))
    })? {
        let (section, from, to, content) = row?;
        let escaped = content.replace('"', "\"\"");
        csv.push_str(&format!("{},{},{},\"{}\"\n", section, from, to, escaped));
    }

    Ok(csv)
}

// ── Accounts ──────────────────────────────────────────────────────────────────

pub fn insert_account(conn: &Connection, a: &Account) -> Result<i64> {
    conn.execute(
        "INSERT INTO accounts (name, initial_balance) VALUES (?1, ?2)",
        params![a.name, a.initial_balance],
    )?;
    Ok(conn.last_insert_rowid())
}

pub fn get_accounts(conn: &Connection) -> Result<Vec<Account>> {
    let mut stmt = conn.prepare(
        "SELECT id, name, initial_balance FROM accounts ORDER BY name"
    )?;
    let rows = stmt.query_map([], |row| {
        Ok(Account {
            id:              Some(row.get(0)?),
            name:            row.get(1)?,
            initial_balance: row.get(2)?,
        })
    })?;
    rows.collect()
}

pub fn delete_account(conn: &Connection, id: i64) -> Result<()> {
    conn.execute("DELETE FROM accounts WHERE id = ?1", params![id])?;
    Ok(())
}

pub fn get_account_balances(conn: &Connection) -> Result<Vec<AccountBalance>> {
    let mut stmt = conn.prepare("
        SELECT
            a.id,
            a.name,
            a.initial_balance
                + COALESCE((SELECT SUM(amount) FROM transfers    WHERE to_account_id   = a.id), 0)
                - COALESCE((SELECT SUM(amount) FROM transfers    WHERE from_account_id = a.id), 0)
                + COALESCE((SELECT SUM(val)    FROM transactions WHERE account_id = a.id AND type = 'revenue'), 0)
                - COALESCE((SELECT SUM(val)    FROM transactions WHERE account_id = a.id AND type = 'expense'), 0)
                AS balance
        FROM accounts a
        ORDER BY a.name
    ")?;
    let rows = stmt.query_map([], |row| {
        Ok(AccountBalance {
            id:      row.get(0)?,
            name:    row.get(1)?,
            balance: row.get(2)?,
        })
    })?;
    rows.collect()
}

// ── Transfers ─────────────────────────────────────────────────────────────────

pub fn insert_transfer(conn: &Connection, t: &Transfer) -> Result<i64> {
    conn.execute(
        "INSERT INTO transfers (from_account_id, to_account_id, amount, date, desc)
         VALUES (?1, ?2, ?3, ?4, ?5)",
        params![t.from_account_id, t.to_account_id, t.amount, t.date, t.desc],
    )?;
    Ok(conn.last_insert_rowid())
}

pub fn get_transfers(conn: &Connection) -> Result<Vec<Transfer>> {
    let mut stmt = conn.prepare("
        SELECT
            t.id, t.from_account_id, t.to_account_id,
            fa.name, ta.name,
            t.amount, t.date, t.desc
        FROM transfers t
        JOIN accounts fa ON fa.id = t.from_account_id
        JOIN accounts ta ON ta.id = t.to_account_id
        ORDER BY t.date DESC, t.id DESC
    ")?;
    let rows = stmt.query_map([], |row| {
        Ok(Transfer {
            id:                Some(row.get(0)?),
            from_account_id:   row.get(1)?,
            to_account_id:     row.get(2)?,
            from_account_name: Some(row.get(3)?),
            to_account_name:   Some(row.get(4)?),
            amount:            row.get(5)?,
            date:              row.get(6)?,
            desc:              row.get(7)?,
        })
    })?;
    rows.collect()
}

pub fn delete_transfer(conn: &Connection, id: i64) -> Result<()> {
    conn.execute("DELETE FROM transfers WHERE id = ?1", params![id])?;
    Ok(())
}

// ── Notes ─────────────────────────────────────────────────────────────────────

pub fn get_note(conn: &Connection, section: &str, from: &str, to: &str) -> Result<String> {
    let result = conn.query_row(
        "SELECT content FROM notes WHERE section = ?1 AND period_from = ?2 AND period_to = ?3",
        params![section, from, to],
        |row| row.get::<_, String>(0),
    );
    match result {
        Ok(content) => Ok(content),
        Err(rusqlite::Error::QueryReturnedNoRows) => Ok(String::new()),
        Err(e) => Err(e),
    }
}

pub fn upsert_note(conn: &Connection, section: &str, from: &str, to: &str, content: &str) -> Result<()> {
    conn.execute(
        "INSERT INTO notes (section, period_from, period_to, content)
         VALUES (?1, ?2, ?3, ?4)
         ON CONFLICT(section, period_from, period_to) DO UPDATE SET content = excluded.content",
        params![section, from, to, content],
    )?;
    Ok(())
}
