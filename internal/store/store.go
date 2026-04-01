package store

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

type DB struct{ conn *sql.DB }

func Open(dataDir string) (*DB, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("create data dir: %w", err)
	}
	conn, err := sql.Open("sqlite", filepath.Join(dataDir, "stampede.db"))
	if err != nil {
		return nil, err
	}
	conn.Exec("PRAGMA journal_mode=WAL")
	conn.Exec("PRAGMA busy_timeout=5000")
	conn.SetMaxOpenConns(4)
	db := &DB{conn: conn}
	if err := db.migrate(); err != nil {
		return nil, err
	}
	return db, nil
}

func (db *DB) Close() error { return db.conn.Close() }

func (db *DB) migrate() error {
	_, err := db.conn.Exec(`
CREATE TABLE IF NOT EXISTS tests (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    url TEXT NOT NULL,
    method TEXT DEFAULT 'GET',
    headers_json TEXT DEFAULT '{}',
    body TEXT DEFAULT '',
    concurrency INTEGER DEFAULT 10,
    duration_seconds INTEGER DEFAULT 30,
    created_at TEXT DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS runs (
    id TEXT PRIMARY KEY,
    test_id TEXT NOT NULL,
    status TEXT DEFAULT 'pending',
    total_requests INTEGER DEFAULT 0,
    success_count INTEGER DEFAULT 0,
    error_count INTEGER DEFAULT 0,
    rps REAL DEFAULT 0,
    avg_ms REAL DEFAULT 0,
    min_ms REAL DEFAULT 0,
    max_ms REAL DEFAULT 0,
    p50_ms REAL DEFAULT 0,
    p95_ms REAL DEFAULT 0,
    p99_ms REAL DEFAULT 0,
    status_codes TEXT DEFAULT '{}',
    errors_json TEXT DEFAULT '[]',
    started_at TEXT DEFAULT '',
    completed_at TEXT DEFAULT '',
    created_at TEXT DEFAULT (datetime('now'))
);
CREATE INDEX IF NOT EXISTS idx_runs_test ON runs(test_id);
`)
	return err
}

// --- Tests ---

type Test struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	URL             string `json:"url"`
	Method          string `json:"method"`
	HeadersJSON     string `json:"headers_json"`
	Body            string `json:"body"`
	Concurrency     int    `json:"concurrency"`
	DurationSeconds int    `json:"duration_seconds"`
	CreatedAt       string `json:"created_at"`
}

func (db *DB) CreateTest(name, url, method, headersJSON, body string, concurrency, durationSec int) (*Test, error) {
	id := "tst_" + genID(8)
	now := time.Now().UTC().Format(time.RFC3339)
	if method == "" {
		method = "GET"
	}
	if concurrency <= 0 {
		concurrency = 10
	}
	if durationSec <= 0 {
		durationSec = 30
	}
	if headersJSON == "" {
		headersJSON = "{}"
	}
	_, err := db.conn.Exec("INSERT INTO tests (id,name,url,method,headers_json,body,concurrency,duration_seconds,created_at) VALUES (?,?,?,?,?,?,?,?,?)",
		id, name, url, method, headersJSON, body, concurrency, durationSec, now)
	if err != nil {
		return nil, err
	}
	return &Test{ID: id, Name: name, URL: url, Method: method, HeadersJSON: headersJSON, Body: body,
		Concurrency: concurrency, DurationSeconds: durationSec, CreatedAt: now}, nil
}

func (db *DB) ListTests() ([]Test, error) {
	rows, err := db.conn.Query("SELECT id,name,url,method,headers_json,body,concurrency,duration_seconds,created_at FROM tests ORDER BY created_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Test
	for rows.Next() {
		var t Test
		rows.Scan(&t.ID, &t.Name, &t.URL, &t.Method, &t.HeadersJSON, &t.Body, &t.Concurrency, &t.DurationSeconds, &t.CreatedAt)
		out = append(out, t)
	}
	return out, rows.Err()
}

func (db *DB) GetTest(id string) (*Test, error) {
	var t Test
	err := db.conn.QueryRow("SELECT id,name,url,method,headers_json,body,concurrency,duration_seconds,created_at FROM tests WHERE id=?", id).
		Scan(&t.ID, &t.Name, &t.URL, &t.Method, &t.HeadersJSON, &t.Body, &t.Concurrency, &t.DurationSeconds, &t.CreatedAt)
	return &t, err
}

func (db *DB) DeleteTest(id string) error {
	db.conn.Exec("DELETE FROM runs WHERE test_id=?", id)
	_, err := db.conn.Exec("DELETE FROM tests WHERE id=?", id)
	return err
}

// --- Runs ---

type Run struct {
	ID            string  `json:"id"`
	TestID        string  `json:"test_id"`
	Status        string  `json:"status"`
	TotalRequests int     `json:"total_requests"`
	SuccessCount  int     `json:"success_count"`
	ErrorCount    int     `json:"error_count"`
	RPS           float64 `json:"rps"`
	AvgMs         float64 `json:"avg_ms"`
	MinMs         float64 `json:"min_ms"`
	MaxMs         float64 `json:"max_ms"`
	P50Ms         float64 `json:"p50_ms"`
	P95Ms         float64 `json:"p95_ms"`
	P99Ms         float64 `json:"p99_ms"`
	StatusCodes   string  `json:"status_codes"`
	ErrorsJSON    string  `json:"errors,omitempty"`
	StartedAt     string  `json:"started_at,omitempty"`
	CompletedAt   string  `json:"completed_at,omitempty"`
	CreatedAt     string  `json:"created_at"`
}

func (db *DB) CreateRun(testID string) (*Run, error) {
	id := "run_" + genID(8)
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := db.conn.Exec("INSERT INTO runs (id,test_id,status,started_at,created_at) VALUES (?,?,?,?,?)",
		id, testID, "running", now, now)
	if err != nil {
		return nil, err
	}
	return &Run{ID: id, TestID: testID, Status: "running", StartedAt: now, CreatedAt: now, StatusCodes: "{}", ErrorsJSON: "[]"}, nil
}

func (db *DB) CompleteRun(id string, total, success, errors int, rps, avgMs, minMs, maxMs, p50, p95, p99 float64, statusCodes, errorsJSON string) {
	now := time.Now().UTC().Format(time.RFC3339)
	db.conn.Exec(`UPDATE runs SET status='completed', total_requests=?, success_count=?, error_count=?,
		rps=?, avg_ms=?, min_ms=?, max_ms=?, p50_ms=?, p95_ms=?, p99_ms=?,
		status_codes=?, errors_json=?, completed_at=? WHERE id=?`,
		total, success, errors, rps, avgMs, minMs, maxMs, p50, p95, p99, statusCodes, errorsJSON, now, id)
}

func (db *DB) FailRun(id, errMsg string) {
	now := time.Now().UTC().Format(time.RFC3339)
	db.conn.Exec("UPDATE runs SET status='failed', errors_json=?, completed_at=? WHERE id=?",
		fmt.Sprintf(`["%s"]`, errMsg), now, id)
}

func (db *DB) GetRun(id string) (*Run, error) {
	var r Run
	err := db.conn.QueryRow(`SELECT id,test_id,status,total_requests,success_count,error_count,
		rps,avg_ms,min_ms,max_ms,p50_ms,p95_ms,p99_ms,status_codes,errors_json,started_at,completed_at,created_at
		FROM runs WHERE id=?`, id).
		Scan(&r.ID, &r.TestID, &r.Status, &r.TotalRequests, &r.SuccessCount, &r.ErrorCount,
			&r.RPS, &r.AvgMs, &r.MinMs, &r.MaxMs, &r.P50Ms, &r.P95Ms, &r.P99Ms,
			&r.StatusCodes, &r.ErrorsJSON, &r.StartedAt, &r.CompletedAt, &r.CreatedAt)
	return &r, err
}

func (db *DB) ListRuns(testID string) ([]Run, error) {
	rows, err := db.conn.Query(`SELECT id,test_id,status,total_requests,success_count,error_count,
		rps,avg_ms,min_ms,max_ms,p50_ms,p95_ms,p99_ms,status_codes,errors_json,started_at,completed_at,created_at
		FROM runs WHERE test_id=? ORDER BY created_at DESC LIMIT 50`, testID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Run
	for rows.Next() {
		var r Run
		rows.Scan(&r.ID, &r.TestID, &r.Status, &r.TotalRequests, &r.SuccessCount, &r.ErrorCount,
			&r.RPS, &r.AvgMs, &r.MinMs, &r.MaxMs, &r.P50Ms, &r.P95Ms, &r.P99Ms,
			&r.StatusCodes, &r.ErrorsJSON, &r.StartedAt, &r.CompletedAt, &r.CreatedAt)
		out = append(out, r)
	}
	return out, rows.Err()
}

func (db *DB) ActiveRuns() int {
	var count int
	db.conn.QueryRow("SELECT COUNT(*) FROM runs WHERE status='running'").Scan(&count)
	return count
}

// --- Stats ---

func (db *DB) Stats() map[string]any {
	var tests, runs, active int
	db.conn.QueryRow("SELECT COUNT(*) FROM tests").Scan(&tests)
	db.conn.QueryRow("SELECT COUNT(*) FROM runs").Scan(&runs)
	db.conn.QueryRow("SELECT COUNT(*) FROM runs WHERE status='running'").Scan(&active)
	return map[string]any{"tests": tests, "runs": runs, "active_runs": active}
}

func (db *DB) Cleanup(days int) (int64, error) {
	cutoff := time.Now().AddDate(0, 0, -days).Format("2006-01-02 15:04:05")
	res, err := db.conn.Exec("DELETE FROM runs WHERE created_at < ?", cutoff)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func genID(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return hex.EncodeToString(b)
}
