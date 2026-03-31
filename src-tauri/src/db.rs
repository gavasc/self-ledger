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
    ")
}

pub fn insert(conn: &Connection, t: &Transaction) -> Result<i64> {
    conn.execute(
        "INSERT INTO transactions (type, desc, cat, val, date)
         VALUES (?1, ?2, ?3, ?4, ?5)",
        params![t.type_, t.desc, t.cat, t.val, t.date],
    )?;
    Ok(conn.last_insert_rowid())
}

pub fn query(conn: &Connection, from: &str, to: &str) -> Result<Vec<Transaction>> {
    let mut stmt = conn.prepare(
        "SELECT id, type, desc, cat, val, date
         FROM transactions
         WHERE date >= ?1 AND date <= ?2
         ORDER BY date DESC"
    )?;
    let rows = stmt.query_map(params![from, to], |row| {
        Ok(Transaction {
            id:    Some(row.get(0)?),
            type_: row.get(1)?,
            desc:  row.get(2)?,
            cat:   row.get(3)?,
            val:   row.get(4)?,
            date:  row.get(5)?,
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
        "SELECT id, type, desc, cat, val, date FROM transactions ORDER BY date DESC"
    )?;
    let rows = stmt.query_map([], |row| {
        Ok(Transaction {
            id:     Some(row.get(0)?),
            type_:  row.get(1)?,
            desc:   row.get(2)?,
            cat:   row.get(3)?,
            val:   row.get(4)?,
            date:   row.get(5)?,
        })
    })?;
    let transactions: Vec<Transaction> = rows.collect::<Result<Vec<_>>>()?;

    Ok(serde_json::to_string_pretty(&transactions).unwrap())
}

pub fn export_csv(conn: &Connection) -> Result<String> {
    let mut stmt = conn.prepare(
        "SELECT id, type, desc, cat, val, date FROM transactions ORDER BY date DESC"
    )?;
    let rows = stmt.query_map([], |row| {
        Ok((
            row.get::<_, i64>(0)?,
            row.get::<_, String>(1)?,
            row.get::<_, String>(2)?,
            row.get::<_, String>(3)?,
            row.get::<_, f64>(4)?,
            row.get::<_, String>(5)?,
        ))
    })?;
    let mut csv = String::from("id,type,desc,cat,val,date\n");

    for row in rows {
        let (id, type_, desc, cat, val, date) = row?;
        csv.push_str(&format!("{},{},{},{},{:.2},{}\n", id, type_, desc, cat, val, date));
    }

    Ok(csv)
}
