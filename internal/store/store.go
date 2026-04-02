package store
import ("database/sql";"fmt";"os";"path/filepath";"time";_ "modernc.org/sqlite")
type DB struct{db *sql.DB}
type Migration struct{
	ID string `json:"id"`
	Name string `json:"name"`
	Version string `json:"version"`
	SQL string `json:"sql_up"`
	Status string `json:"status"`
	AppliedAt string `json:"applied_at"`
	CreatedAt string `json:"created_at"`
}
func Open(d string)(*DB,error){if err:=os.MkdirAll(d,0755);err!=nil{return nil,err};db,err:=sql.Open("sqlite",filepath.Join(d,"stampede.db")+"?_journal_mode=WAL&_busy_timeout=5000");if err!=nil{return nil,err}
db.Exec(`CREATE TABLE IF NOT EXISTS migrations(id TEXT PRIMARY KEY,name TEXT NOT NULL,version TEXT DEFAULT '',sql_up TEXT DEFAULT '',status TEXT DEFAULT 'pending',applied_at TEXT DEFAULT '',created_at TEXT DEFAULT(datetime('now')))`)
return &DB{db:db},nil}
func(d *DB)Close()error{return d.db.Close()}
func genID()string{return fmt.Sprintf("%d",time.Now().UnixNano())}
func now()string{return time.Now().UTC().Format(time.RFC3339)}
func(d *DB)Create(e *Migration)error{e.ID=genID();e.CreatedAt=now();_,err:=d.db.Exec(`INSERT INTO migrations(id,name,version,sql_up,status,applied_at,created_at)VALUES(?,?,?,?,?,?,?)`,e.ID,e.Name,e.Version,e.SQL,e.Status,e.AppliedAt,e.CreatedAt);return err}
func(d *DB)Get(id string)*Migration{var e Migration;if d.db.QueryRow(`SELECT id,name,version,sql_up,status,applied_at,created_at FROM migrations WHERE id=?`,id).Scan(&e.ID,&e.Name,&e.Version,&e.SQL,&e.Status,&e.AppliedAt,&e.CreatedAt)!=nil{return nil};return &e}
func(d *DB)List()[]Migration{rows,_:=d.db.Query(`SELECT id,name,version,sql_up,status,applied_at,created_at FROM migrations ORDER BY created_at DESC`);if rows==nil{return nil};defer rows.Close();var o []Migration;for rows.Next(){var e Migration;rows.Scan(&e.ID,&e.Name,&e.Version,&e.SQL,&e.Status,&e.AppliedAt,&e.CreatedAt);o=append(o,e)};return o}
func(d *DB)Delete(id string)error{_,err:=d.db.Exec(`DELETE FROM migrations WHERE id=?`,id);return err}
func(d *DB)Count()int{var n int;d.db.QueryRow(`SELECT COUNT(*) FROM migrations`).Scan(&n);return n}
