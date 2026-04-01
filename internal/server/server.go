package server

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/stockyard-dev/stockyard-stampede/internal/store"
)

type Server struct {
	db     *store.DB
	mux    *http.ServeMux
	port   int
	limits Limits

	mu         sync.Mutex
	liveStats  map[string]*LiveStats // run_id → stats
}

type LiveStats struct {
	Total    atomic.Int64
	Success  atomic.Int64
	Errors   atomic.Int64
	Running  atomic.Bool
	StartedAt time.Time
}

func New(db *store.DB, port int, limits Limits) *Server {
	s := &Server{db: db, mux: http.NewServeMux(), port: port, limits: limits, liveStats: make(map[string]*LiveStats)}
	s.routes()
	return s
}

func (s *Server) routes() {
	s.mux.HandleFunc("POST /api/tests", s.handleCreateTest)
	s.mux.HandleFunc("GET /api/tests", s.handleListTests)
	s.mux.HandleFunc("GET /api/tests/{id}", s.handleGetTest)
	s.mux.HandleFunc("DELETE /api/tests/{id}", s.handleDeleteTest)

	s.mux.HandleFunc("POST /api/tests/{id}/run", s.handleRunTest)
	s.mux.HandleFunc("GET /api/runs/{id}", s.handleGetRun)
	s.mux.HandleFunc("GET /api/tests/{id}/runs", s.handleListRuns)
	s.mux.HandleFunc("GET /api/runs/{id}/live", s.handleLiveStats)

	s.mux.HandleFunc("GET /api/status", s.handleStatus)
	s.mux.HandleFunc("GET /health", s.handleHealth)
	s.mux.HandleFunc("GET /ui", s.handleUI)
	s.mux.HandleFunc("GET /api/version", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, map[string]any{"product": "stockyard-stampede", "version": "0.1.0"})
	})
}

func (s *Server) Start() error {
	addr := fmt.Sprintf(":%d", s.port)
	log.Printf("[stampede] listening on %s", addr)
	return http.ListenAndServe(addr, s.mux)
}

// --- Load test engine ---

func (s *Server) runLoadTest(test *store.Test, run *store.Run) {
	live := &LiveStats{StartedAt: time.Now()}
	live.Running.Store(true)
	s.mu.Lock()
	s.liveStats[run.ID] = live
	s.mu.Unlock()

	defer func() {
		live.Running.Store(false)
		time.Sleep(30 * time.Second) // keep live stats available briefly
		s.mu.Lock()
		delete(s.liveStats, run.ID)
		s.mu.Unlock()
	}()

	concurrency := test.Concurrency
	if s.limits.MaxConcurrency > 0 && concurrency > s.limits.MaxConcurrency {
		concurrency = s.limits.MaxConcurrency
	}
	duration := time.Duration(test.DurationSeconds) * time.Second
	if s.limits.MaxDuration > 0 && test.DurationSeconds > s.limits.MaxDuration {
		duration = time.Duration(s.limits.MaxDuration) * time.Second
	}

	var headers map[string]string
	json.Unmarshal([]byte(test.HeadersJSON), &headers)

	var mu sync.Mutex
	var latencies []float64
	statusCodes := make(map[int]int)
	var errMsgs []string
	var successCount, errorCount int

	client := &http.Client{Timeout: 10 * time.Second}
	deadline := time.After(duration)
	stop := make(chan struct{})

	var wg sync.WaitGroup
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stop:
					return
				default:
				}

				var bodyReader io.Reader
				if test.Body != "" {
					bodyReader = strings.NewReader(test.Body)
				}

				req, err := http.NewRequest(test.Method, test.URL, bodyReader)
				if err != nil {
					live.Errors.Add(1)
					mu.Lock()
					errorCount++
					if len(errMsgs) < 10 {
						errMsgs = append(errMsgs, err.Error())
					}
					mu.Unlock()
					continue
				}
				for k, v := range headers {
					req.Header.Set(k, v)
				}
				if test.Body != "" {
					req.Header.Set("Content-Type", "application/json")
				}

				start := time.Now()
				resp, err := client.Do(req)
				latency := float64(time.Since(start).Milliseconds())
				live.Total.Add(1)

				if err != nil {
					live.Errors.Add(1)
					mu.Lock()
					errorCount++
					if len(errMsgs) < 10 {
						errMsgs = append(errMsgs, err.Error())
					}
					mu.Unlock()
					continue
				}
				io.Copy(io.Discard, resp.Body)
				resp.Body.Close()

				mu.Lock()
				latencies = append(latencies, latency)
				statusCodes[resp.StatusCode]++
				if resp.StatusCode >= 200 && resp.StatusCode < 400 {
					successCount++
					live.Success.Add(1)
				} else {
					errorCount++
					live.Errors.Add(1)
				}
				mu.Unlock()
			}
		}()
	}

	<-deadline
	close(stop)
	wg.Wait()

	// Calculate results
	total := successCount + errorCount
	elapsed := time.Since(live.StartedAt).Seconds()
	rps := 0.0
	if elapsed > 0 {
		rps = float64(total) / elapsed
	}

	var avgMs, minMs, maxMs, p50, p95, p99 float64
	if len(latencies) > 0 {
		sort.Float64s(latencies)
		minMs = latencies[0]
		maxMs = latencies[len(latencies)-1]
		sum := 0.0
		for _, l := range latencies {
			sum += l
		}
		avgMs = sum / float64(len(latencies))
		p50 = percentile(latencies, 50)
		p95 = percentile(latencies, 95)
		p99 = percentile(latencies, 99)
	}

	scJSON, _ := json.Marshal(statusCodes)
	errJSON, _ := json.Marshal(errMsgs)

	s.db.CompleteRun(run.ID, total, successCount, errorCount,
		math.Round(rps*100)/100,
		math.Round(avgMs*100)/100,
		math.Round(minMs*100)/100,
		math.Round(maxMs*100)/100,
		math.Round(p50*100)/100,
		math.Round(p95*100)/100,
		math.Round(p99*100)/100,
		string(scJSON), string(errJSON))

	log.Printf("[stampede] run %s complete: %d requests, %.1f rps, p50=%.0fms p99=%.0fms, %d errors",
		run.ID, total, rps, p50, p99, errorCount)
}

func percentile(sorted []float64, pct float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	idx := int(math.Ceil(pct/100*float64(len(sorted)))) - 1
	if idx < 0 {
		idx = 0
	}
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}
	return sorted[idx]
}

// --- Handlers ---

func (s *Server) handleCreateTest(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name            string `json:"name"`
		URL             string `json:"url"`
		Method          string `json:"method"`
		Headers         map[string]string `json:"headers"`
		Body            string `json:"body"`
		Concurrency     int    `json:"concurrency"`
		DurationSeconds int    `json:"duration_seconds"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, 400, map[string]string{"error": "invalid JSON"})
		return
	}
	if req.URL == "" {
		writeJSON(w, 400, map[string]string{"error": "url is required"})
		return
	}
	if req.Name == "" {
		req.Name = req.URL
	}
	if s.limits.MaxTests > 0 {
		tests, _ := s.db.ListTests()
		if LimitReached(s.limits.MaxTests, len(tests)) {
			writeJSON(w, 402, map[string]string{"error": fmt.Sprintf("free tier limit: %d tests — upgrade to Pro", s.limits.MaxTests), "upgrade": "https://stockyard.dev/stampede/"})
			return
		}
	}
	if req.Concurrency > s.limits.MaxConcurrency && s.limits.MaxConcurrency > 0 {
		req.Concurrency = s.limits.MaxConcurrency
	}
	if req.DurationSeconds > s.limits.MaxDuration && s.limits.MaxDuration > 0 {
		req.DurationSeconds = s.limits.MaxDuration
	}
	headersJSON := "{}"
	if len(req.Headers) > 0 {
		b, _ := json.Marshal(req.Headers)
		headersJSON = string(b)
	}
	test, err := s.db.CreateTest(req.Name, req.URL, req.Method, headersJSON, req.Body, req.Concurrency, req.DurationSeconds)
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, 201, map[string]any{"test": test})
}

func (s *Server) handleListTests(w http.ResponseWriter, r *http.Request) {
	tests, err := s.db.ListTests()
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	if tests == nil {
		tests = []store.Test{}
	}
	writeJSON(w, 200, map[string]any{"tests": tests, "count": len(tests)})
}

func (s *Server) handleGetTest(w http.ResponseWriter, r *http.Request) {
	t, err := s.db.GetTest(r.PathValue("id"))
	if err != nil {
		writeJSON(w, 404, map[string]string{"error": "test not found"})
		return
	}
	writeJSON(w, 200, map[string]any{"test": t})
}

func (s *Server) handleDeleteTest(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if _, err := s.db.GetTest(id); err != nil {
		writeJSON(w, 404, map[string]string{"error": "test not found"})
		return
	}
	s.db.DeleteTest(id)
	writeJSON(w, 200, map[string]string{"status": "deleted"})
}

func (s *Server) handleRunTest(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	test, err := s.db.GetTest(id)
	if err != nil {
		writeJSON(w, 404, map[string]string{"error": "test not found"})
		return
	}

	// Limit concurrent runs
	active := s.db.ActiveRuns()
	if active > 0 {
		writeJSON(w, 429, map[string]string{"error": "a test is already running — wait for it to finish"})
		return
	}

	run, err := s.db.CreateRun(id)
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}

	go s.runLoadTest(test, run)

	log.Printf("[stampede] started run %s: %s %s (%d workers, %ds)", run.ID, test.Method, test.URL, test.Concurrency, test.DurationSeconds)
	writeJSON(w, 200, map[string]any{"run": run, "message": fmt.Sprintf("running %d workers for %ds", test.Concurrency, test.DurationSeconds)})
}

func (s *Server) handleGetRun(w http.ResponseWriter, r *http.Request) {
	run, err := s.db.GetRun(r.PathValue("id"))
	if err != nil {
		writeJSON(w, 404, map[string]string{"error": "run not found"})
		return
	}
	writeJSON(w, 200, map[string]any{"run": run})
}

func (s *Server) handleListRuns(w http.ResponseWriter, r *http.Request) {
	runs, err := s.db.ListRuns(r.PathValue("id"))
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	if runs == nil {
		runs = []store.Run{}
	}
	writeJSON(w, 200, map[string]any{"runs": runs, "count": len(runs)})
}

func (s *Server) handleLiveStats(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	s.mu.Lock()
	live, ok := s.liveStats[id]
	s.mu.Unlock()
	if !ok {
		writeJSON(w, 404, map[string]string{"error": "no live stats for this run"})
		return
	}
	elapsed := time.Since(live.StartedAt).Seconds()
	total := live.Total.Load()
	rps := 0.0
	if elapsed > 0 {
		rps = float64(total) / elapsed
	}
	writeJSON(w, 200, map[string]any{
		"running":  live.Running.Load(),
		"total":    total,
		"success":  live.Success.Load(),
		"errors":   live.Errors.Load(),
		"elapsed":  fmt.Sprintf("%.1f", elapsed),
		"rps":      fmt.Sprintf("%.1f", rps),
	})
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, s.db.Stats())
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, map[string]string{"status": "ok"})
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}
