package store
import ("database/sql";"fmt";"os";"path/filepath";"time";_ "modernc.org/sqlite")
type DB struct{db *sql.DB}
type LoadTest struct {
	ID string `json:"id"`
	Name string `json:"name"`
	TargetURL string `json:"target_url"`
	Concurrency int `json:"concurrency"`
	Duration int `json:"duration_seconds"`
	RequestCount int `json:"request_count"`
	AvgLatency int `json:"avg_latency_ms"`
	ErrorRate int `json:"error_rate"`
	Status string `json:"status"`
	CreatedAt string `json:"created_at"`
}
func Open(d string)(*DB,error){if err:=os.MkdirAll(d,0755);err!=nil{return nil,err};db,err:=sql.Open("sqlite",filepath.Join(d,"stampede.db")+"?_journal_mode=WAL&_busy_timeout=5000");if err!=nil{return nil,err}
db.Exec(`CREATE TABLE IF NOT EXISTS load_tests(id TEXT PRIMARY KEY,name TEXT NOT NULL,target_url TEXT DEFAULT '',concurrency INTEGER DEFAULT 10,duration_seconds INTEGER DEFAULT 30,request_count INTEGER DEFAULT 0,avg_latency_ms INTEGER DEFAULT 0,error_rate INTEGER DEFAULT 0,status TEXT DEFAULT 'pending',created_at TEXT DEFAULT(datetime('now')))`)
return &DB{db:db},nil}
func(d *DB)Close()error{return d.db.Close()}
func genID()string{return fmt.Sprintf("%d",time.Now().UnixNano())}
func now()string{return time.Now().UTC().Format(time.RFC3339)}
func(d *DB)Create(e *LoadTest)error{e.ID=genID();e.CreatedAt=now();_,err:=d.db.Exec(`INSERT INTO load_tests(id,name,target_url,concurrency,duration_seconds,request_count,avg_latency_ms,error_rate,status,created_at)VALUES(?,?,?,?,?,?,?,?,?,?)`,e.ID,e.Name,e.TargetURL,e.Concurrency,e.Duration,e.RequestCount,e.AvgLatency,e.ErrorRate,e.Status,e.CreatedAt);return err}
func(d *DB)Get(id string)*LoadTest{var e LoadTest;if d.db.QueryRow(`SELECT id,name,target_url,concurrency,duration_seconds,request_count,avg_latency_ms,error_rate,status,created_at FROM load_tests WHERE id=?`,id).Scan(&e.ID,&e.Name,&e.TargetURL,&e.Concurrency,&e.Duration,&e.RequestCount,&e.AvgLatency,&e.ErrorRate,&e.Status,&e.CreatedAt)!=nil{return nil};return &e}
func(d *DB)List()[]LoadTest{rows,_:=d.db.Query(`SELECT id,name,target_url,concurrency,duration_seconds,request_count,avg_latency_ms,error_rate,status,created_at FROM load_tests ORDER BY created_at DESC`);if rows==nil{return nil};defer rows.Close();var o []LoadTest;for rows.Next(){var e LoadTest;rows.Scan(&e.ID,&e.Name,&e.TargetURL,&e.Concurrency,&e.Duration,&e.RequestCount,&e.AvgLatency,&e.ErrorRate,&e.Status,&e.CreatedAt);o=append(o,e)};return o}
func(d *DB)Update(e *LoadTest)error{_,err:=d.db.Exec(`UPDATE load_tests SET name=?,target_url=?,concurrency=?,duration_seconds=?,request_count=?,avg_latency_ms=?,error_rate=?,status=? WHERE id=?`,e.Name,e.TargetURL,e.Concurrency,e.Duration,e.RequestCount,e.AvgLatency,e.ErrorRate,e.Status,e.ID);return err}
func(d *DB)Delete(id string)error{_,err:=d.db.Exec(`DELETE FROM load_tests WHERE id=?`,id);return err}
func(d *DB)Count()int{var n int;d.db.QueryRow(`SELECT COUNT(*) FROM load_tests`).Scan(&n);return n}

func(d *DB)Search(q string, filters map[string]string)[]LoadTest{
    where:="1=1"
    args:=[]any{}
    if q!=""{
        where+=" AND (name LIKE ?)"
        args=append(args,"%"+q+"%");
    }
    if v,ok:=filters["status"];ok&&v!=""{where+=" AND status=?";args=append(args,v)}
    rows,_:=d.db.Query(`SELECT id,name,target_url,concurrency,duration_seconds,request_count,avg_latency_ms,error_rate,status,created_at FROM load_tests WHERE `+where+` ORDER BY created_at DESC`,args...)
    if rows==nil{return nil};defer rows.Close()
    var o []LoadTest;for rows.Next(){var e LoadTest;rows.Scan(&e.ID,&e.Name,&e.TargetURL,&e.Concurrency,&e.Duration,&e.RequestCount,&e.AvgLatency,&e.ErrorRate,&e.Status,&e.CreatedAt);o=append(o,e)};return o
}

func(d *DB)Stats()map[string]any{
    m:=map[string]any{"total":d.Count()}
    rows,_:=d.db.Query(`SELECT status,COUNT(*) FROM load_tests GROUP BY status`)
    if rows!=nil{defer rows.Close();by:=map[string]int{};for rows.Next(){var s string;var c int;rows.Scan(&s,&c);by[s]=c};m["by_status"]=by}
    return m
}
