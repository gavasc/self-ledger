package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	_ "modernc.org/sqlite"
)

// ── Structs ───────────────────────────────────────────────────────────────────

type Transaction struct {
	ID               *int64   `json:"id"`
	Type             string   `json:"type"`
	Desc             string   `json:"desc"`
	Cat              string   `json:"cat"`
	Val              float64  `json:"val"`
	Date             string   `json:"date"`
	AccountID        *int64   `json:"account_id"`
	InstallmentID    *int64   `json:"installment_id"`
	InstallmentIndex *int64   `json:"installment_index"`
}

type Installment struct {
	ID            *int64   `json:"id"`
	Desc          string   `json:"desc"`
	Cat           string   `json:"cat"`
	TotalVal      float64  `json:"total_val"`
	NInstallments int64    `json:"n_installments"`
	StartDate     string   `json:"start_date"`
	AccountID     *int64   `json:"account_id"`
	PaidCount     *int64   `json:"paid_count"`
	MonthlyVal    *float64 `json:"monthly_val"`
}

type Account struct {
	ID             *int64  `json:"id"`
	Name           string  `json:"name"`
	InitialBalance float64 `json:"initial_balance"`
}

type AccountBalance struct {
	ID      int64   `json:"id"`
	Name    string  `json:"name"`
	Balance float64 `json:"balance"`
}

type Transfer struct {
	ID              *int64   `json:"id"`
	FromAccountID   int64    `json:"from_account_id"`
	ToAccountID     int64    `json:"to_account_id"`
	FromAccountName *string  `json:"from_account_name"`
	ToAccountName   *string  `json:"to_account_name"`
	Amount          float64  `json:"amount"`
	Date            string   `json:"date"`
	Desc            string   `json:"desc"`
}

// ── DB ────────────────────────────────────────────────────────────────────────

type DB struct {
	conn *sql.DB
}

func NewDB(path string) (*DB, error) {
	conn, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	// Serialize all access through a single connection (mirrors Rust's Mutex<Connection>)
	conn.SetMaxOpenConns(1)
	db := &DB{conn: conn}
	if err := db.init(); err != nil {
		conn.Close()
		return nil, err
	}
	return db, nil
}

func (db *DB) Close() { db.conn.Close() }

func (db *DB) init() error {
	_, err := db.conn.Exec(`
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
		CREATE TABLE IF NOT EXISTS installments (
			id             INTEGER PRIMARY KEY AUTOINCREMENT,
			desc           TEXT    NOT NULL,
			cat            TEXT    NOT NULL,
			total_val      REAL    NOT NULL,
			n_installments INTEGER NOT NULL,
			start_date     TEXT    NOT NULL,
			account_id     INTEGER REFERENCES accounts(id)
		);
	`)
	if err != nil {
		return err
	}
	// Migrations — errors silently discarded (column may already exist)
	db.conn.Exec("ALTER TABLE transactions ADD COLUMN account_id INTEGER REFERENCES accounts(id)")
	db.conn.Exec("ALTER TABLE transactions ADD COLUMN installment_id INTEGER REFERENCES installments(id)")
	db.conn.Exec("ALTER TABLE transactions ADD COLUMN installment_index INTEGER")
	return nil
}

// ── Transactions ──────────────────────────────────────────────────────────────

func (db *DB) InsertTransaction(t Transaction) (int64, error) {
	res, err := db.conn.Exec(
		"INSERT INTO transactions (type, desc, cat, val, date, account_id) VALUES (?, ?, ?, ?, ?, ?)",
		t.Type, t.Desc, t.Cat, t.Val, t.Date, t.AccountID,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (db *DB) UpdateTransaction(t Transaction) error {
	_, err := db.conn.Exec(
		"UPDATE transactions SET type = ?, desc = ?, cat = ?, val = ?, date = ?, account_id = ? WHERE id = ?",
		t.Type, t.Desc, t.Cat, t.Val, t.Date, t.AccountID, t.ID,
	)
	return err
}

func (db *DB) QueryTransactions(from, to string) ([]Transaction, error) {
	rows, err := db.conn.Query(
		`SELECT id, type, desc, cat, val, date, account_id, installment_id, installment_index
		 FROM transactions
		 WHERE date >= ? AND date <= ?
		 ORDER BY date DESC`,
		from, to,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanTransactions(rows)
}

func (db *DB) DeleteTransaction(id int64) error {
	_, err := db.conn.Exec("DELETE FROM transactions WHERE id = ?", id)
	return err
}

func scanTransactions(rows *sql.Rows) ([]Transaction, error) {
	var out []Transaction
	for rows.Next() {
		var t Transaction
		if err := rows.Scan(
			&t.ID, &t.Type, &t.Desc, &t.Cat, &t.Val, &t.Date,
			&t.AccountID, &t.InstallmentID, &t.InstallmentIndex,
		); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

// ── Accounts ──────────────────────────────────────────────────────────────────

func (db *DB) InsertAccount(a Account) (int64, error) {
	res, err := db.conn.Exec(
		"INSERT INTO accounts (name, initial_balance) VALUES (?, ?)",
		a.Name, a.InitialBalance,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (db *DB) GetAccounts() ([]Account, error) {
	rows, err := db.conn.Query("SELECT id, name, initial_balance FROM accounts ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Account
	for rows.Next() {
		var a Account
		if err := rows.Scan(&a.ID, &a.Name, &a.InitialBalance); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (db *DB) DeleteAccount(id int64) error {
	_, err := db.conn.Exec("DELETE FROM accounts WHERE id = ?", id)
	return err
}

func (db *DB) GetAccountBalances() ([]AccountBalance, error) {
	rows, err := db.conn.Query(`
		SELECT
			a.id,
			a.name,
			a.initial_balance
				+ COALESCE((SELECT SUM(amount) FROM transfers    WHERE to_account_id   = a.id), 0)
				- COALESCE((SELECT SUM(amount) FROM transfers    WHERE from_account_id = a.id), 0)
				+ COALESCE((SELECT SUM(val)    FROM transactions WHERE account_id = a.id AND type = 'revenue' AND date <= date('now')), 0)
				- COALESCE((SELECT SUM(val)    FROM transactions WHERE account_id = a.id AND type = 'expense' AND date <= date('now')), 0)
				AS balance
		FROM accounts a
		ORDER BY a.name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []AccountBalance
	for rows.Next() {
		var ab AccountBalance
		if err := rows.Scan(&ab.ID, &ab.Name, &ab.Balance); err != nil {
			return nil, err
		}
		out = append(out, ab)
	}
	return out, rows.Err()
}

// ── Transfers ─────────────────────────────────────────────────────────────────

func (db *DB) InsertTransfer(t Transfer) (int64, error) {
	res, err := db.conn.Exec(
		"INSERT INTO transfers (from_account_id, to_account_id, amount, date, desc) VALUES (?, ?, ?, ?, ?)",
		t.FromAccountID, t.ToAccountID, t.Amount, t.Date, t.Desc,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (db *DB) GetTransfers() ([]Transfer, error) {
	rows, err := db.conn.Query(`
		SELECT
			t.id, t.from_account_id, t.to_account_id,
			fa.name, ta.name,
			t.amount, t.date, t.desc
		FROM transfers t
		JOIN accounts fa ON fa.id = t.from_account_id
		JOIN accounts ta ON ta.id = t.to_account_id
		ORDER BY t.date DESC, t.id DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Transfer
	for rows.Next() {
		var tr Transfer
		if err := rows.Scan(
			&tr.ID, &tr.FromAccountID, &tr.ToAccountID,
			&tr.FromAccountName, &tr.ToAccountName,
			&tr.Amount, &tr.Date, &tr.Desc,
		); err != nil {
			return nil, err
		}
		out = append(out, tr)
	}
	return out, rows.Err()
}

func (db *DB) DeleteTransfer(id int64) error {
	_, err := db.conn.Exec("DELETE FROM transfers WHERE id = ?", id)
	return err
}

// ── Installments ──────────────────────────────────────────────────────────────

func (db *DB) InsertInstallment(inst Installment) (int64, error) {
	res, err := db.conn.Exec(
		"INSERT INTO installments (desc, cat, total_val, n_installments, start_date, account_id) VALUES (?, ?, ?, ?, ?, ?)",
		inst.Desc, inst.Cat, inst.TotalVal, inst.NInstallments, inst.StartDate, inst.AccountID,
	)
	if err != nil {
		return 0, err
	}
	installmentID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	monthly := inst.TotalVal / float64(inst.NInstallments)

	// Parse start_date "YYYY-MM-DD" using integer arithmetic (matches Rust behavior exactly)
	parts := strings.SplitN(inst.StartDate, "-", 3)
	year := parseIntOrDefault(parts, 0, 2024)
	month := parseIntOrDefault(parts, 1, 1)
	day := parseIntOrDefault(parts, 2, 1)

	for i := int64(0); i < inst.NInstallments; i++ {
		date := fmt.Sprintf("%04d-%02d-%02d", year, month, day)
		_, err = db.conn.Exec(
			`INSERT INTO transactions (type, desc, cat, val, date, account_id, installment_id, installment_index)
			 VALUES ('expense', ?, ?, ?, ?, ?, ?, ?)`,
			inst.Desc, inst.Cat, monthly, date, inst.AccountID, installmentID, i+1,
		)
		if err != nil {
			return 0, err
		}
		month++
		if month > 12 {
			month = 1
			year++
		}
	}
	return installmentID, nil
}

func parseIntOrDefault(parts []string, idx, def int) int {
	if idx >= len(parts) {
		return def
	}
	var v int
	fmt.Sscanf(parts[idx], "%d", &v)
	if v == 0 {
		return def
	}
	return v
}

func (db *DB) GetInstallments() ([]Installment, error) {
	rows, err := db.conn.Query(`
		SELECT i.id, i.desc, i.cat, i.total_val, i.n_installments, i.start_date, i.account_id,
		       COUNT(t.id) as paid_count
		FROM installments i
		LEFT JOIN transactions t ON t.installment_id = i.id
		GROUP BY i.id
		ORDER BY i.start_date DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Installment
	for rows.Next() {
		var inst Installment
		var paidCount int64
		if err := rows.Scan(
			&inst.ID, &inst.Desc, &inst.Cat, &inst.TotalVal, &inst.NInstallments,
			&inst.StartDate, &inst.AccountID, &paidCount,
		); err != nil {
			return nil, err
		}
		monthly := inst.TotalVal / float64(inst.NInstallments)
		inst.PaidCount = &paidCount
		inst.MonthlyVal = &monthly
		out = append(out, inst)
	}
	return out, rows.Err()
}

func (db *DB) DeleteInstallment(id int64) error {
	if _, err := db.conn.Exec("DELETE FROM transactions WHERE installment_id = ?", id); err != nil {
		return err
	}
	_, err := db.conn.Exec("DELETE FROM installments WHERE id = ?", id)
	return err
}

// ── Notes ─────────────────────────────────────────────────────────────────────

func (db *DB) GetNote(section, from, to string) (string, error) {
	var content string
	err := db.conn.QueryRow(
		"SELECT content FROM notes WHERE section = ? AND period_from = ? AND period_to = ?",
		section, from, to,
	).Scan(&content)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return content, err
}

func (db *DB) UpsertNote(section, from, to, content string) error {
	_, err := db.conn.Exec(
		`INSERT INTO notes (section, period_from, period_to, content) VALUES (?, ?, ?, ?)
		 ON CONFLICT(section, period_from, period_to) DO UPDATE SET content = excluded.content`,
		section, from, to, content,
	)
	return err
}

// ── Export ────────────────────────────────────────────────────────────────────

func (db *DB) ExportJSON() (string, error) {
	// Transactions
	rows, err := db.conn.Query(
		"SELECT id, type, desc, cat, val, date, account_id, installment_id, installment_index FROM transactions ORDER BY date DESC",
	)
	if err != nil {
		return "", err
	}
	defer rows.Close()
	transactions, err := scanTransactions(rows)
	if err != nil {
		return "", err
	}

	// Accounts
	accounts, err := db.GetAccountBalances()
	if err != nil {
		return "", err
	}

	// Transfers
	trows, err := db.conn.Query(`
		SELECT t.id, t.from_account_id, t.to_account_id,
		       fa.name, ta.name, t.amount, t.date, t.desc
		FROM transfers t
		JOIN accounts fa ON fa.id = t.from_account_id
		JOIN accounts ta ON ta.id = t.to_account_id
		ORDER BY t.date DESC, t.id DESC
	`)
	if err != nil {
		return "", err
	}
	defer trows.Close()
	var transfers []Transfer
	for trows.Next() {
		var tr Transfer
		if err := trows.Scan(
			&tr.ID, &tr.FromAccountID, &tr.ToAccountID,
			&tr.FromAccountName, &tr.ToAccountName,
			&tr.Amount, &tr.Date, &tr.Desc,
		); err != nil {
			return "", err
		}
		transfers = append(transfers, tr)
	}

	// Notes
	nrows, err := db.conn.Query(
		"SELECT section, period_from, period_to, content FROM notes WHERE content != '' ORDER BY section, period_from",
	)
	if err != nil {
		return "", err
	}
	defer nrows.Close()
	var notes []map[string]string
	for nrows.Next() {
		var section, periodFrom, periodTo, content string
		if err := nrows.Scan(&section, &periodFrom, &periodTo, &content); err != nil {
			return "", err
		}
		notes = append(notes, map[string]string{
			"section":   section,
			"date-from": periodFrom,
			"date-to":   periodTo,
			"note":      content,
		})
	}

	// Installments
	installments, err := db.GetInstallments()
	if err != nil {
		return "", err
	}

	out := map[string]any{
		"transactions": transactions,
		"accounts":     accounts,
		"transfers":    transfers,
		"notes":        notes,
		"installments": installments,
	}
	b, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (db *DB) ExportCSV() (string, error) {
	var csv strings.Builder

	// transactions
	csv.WriteString("transactions\n")
	csv.WriteString("id,type,desc,cat,val,date,account_id,installment_id,installment_index\n")
	rows, err := db.conn.Query(
		"SELECT id, type, desc, cat, val, date, account_id, installment_id, installment_index FROM transactions ORDER BY date DESC",
	)
	if err != nil {
		return "", err
	}
	defer rows.Close()
	for rows.Next() {
		var id int64
		var typ, desc, cat, date string
		var val float64
		var accountID, installmentID, installmentIndex *int64
		if err := rows.Scan(&id, &typ, &desc, &cat, &val, &date, &accountID, &installmentID, &installmentIndex); err != nil {
			return "", err
		}
		csv.WriteString(fmt.Sprintf("%d,%s,%s,%s,%.2f,%s,%s,%s,%s\n",
			id, typ, desc, cat, val, date,
			int64PtrStr(accountID), int64PtrStr(installmentID), int64PtrStr(installmentIndex),
		))
	}

	// accounts
	csv.WriteString("\naccounts\n")
	csv.WriteString("id,name,balance\n")
	accounts, err := db.GetAccountBalances()
	if err != nil {
		return "", err
	}
	for _, a := range accounts {
		csv.WriteString(fmt.Sprintf("%d,%s,%.2f\n", a.ID, a.Name, a.Balance))
	}

	// transfers
	csv.WriteString("\ntransfers\n")
	csv.WriteString("id,from_account,to_account,amount,date,desc\n")
	trows, err := db.conn.Query(`
		SELECT t.id, fa.name, ta.name, t.amount, t.date, t.desc
		FROM transfers t
		JOIN accounts fa ON fa.id = t.from_account_id
		JOIN accounts ta ON ta.id = t.to_account_id
		ORDER BY t.date DESC, t.id DESC
	`)
	if err != nil {
		return "", err
	}
	defer trows.Close()
	for trows.Next() {
		var id int64
		var from, to, date, desc string
		var amount float64
		if err := trows.Scan(&id, &from, &to, &amount, &date, &desc); err != nil {
			return "", err
		}
		csv.WriteString(fmt.Sprintf("%d,%s,%s,%.2f,%s,%s\n", id, from, to, amount, date, desc))
	}

	// notes
	csv.WriteString("\nnotes\n")
	csv.WriteString("section,date-from,date-to,note\n")
	nrows, err := db.conn.Query(
		"SELECT section, period_from, period_to, content FROM notes WHERE content != '' ORDER BY section, period_from",
	)
	if err != nil {
		return "", err
	}
	defer nrows.Close()
	for nrows.Next() {
		var section, periodFrom, periodTo, content string
		if err := nrows.Scan(&section, &periodFrom, &periodTo, &content); err != nil {
			return "", err
		}
		escaped := strings.ReplaceAll(content, `"`, `""`)
		csv.WriteString(fmt.Sprintf("%s,%s,%s,\"%s\"\n", section, periodFrom, periodTo, escaped))
	}

	return csv.String(), nil
}

func int64PtrStr(v *int64) string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%d", *v)
}
