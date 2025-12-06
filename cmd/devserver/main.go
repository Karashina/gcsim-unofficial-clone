package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/simulator"
)

type jobStatus string

const (
	statusPending jobStatus = "pending"
	statusDone    jobStatus = "done"
	statusError   jobStatus = "error"
)

type job struct {
	ID     string      `json:"id"`
	Status jobStatus   `json:"status"`
	Result interface{} `json:"result,omitempty"`
}

var (
	jobs   = map[string]*job{}
	jobsMu sync.Mutex
)

// corsMiddleware adds CORS headers for development
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// corsHandler wraps a HandlerFunc with CORS headers
func corsHandler(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next(w, r)
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())
	// static UI dir
	uiDir := "webui/dist"
	if _, err := os.Stat(uiDir); os.IsNotExist(err) {
		uiDir = "webui"
	}

	log.Printf("Using UI directory: %s", uiDir)

	fs := http.FileServer(http.Dir(uiDir))
	http.Handle("/ui/", corsMiddleware(http.StripPrefix("/ui/", fs)))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/ui/", http.StatusFound)
	})

	http.HandleFunc("/api/simulate", corsHandler(simulateHandler))
	http.HandleFunc("/api/result", corsHandler(resultHandler))

	addr := ":8381"
	if port := os.Getenv("DEVSERVER_PORT"); port != "" {
		addr = ":" + port
	}
	srv := &http.Server{Addr: addr}
	go func() {
		log.Printf("devserver listening on %s", addr)
		log.Printf("Open your browser at http://localhost%s/ui/", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	// wait for interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
}

func simulateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// read body bytes so we can support either JSON { config: "..." } or raw text
	bodyBytes, _ := io.ReadAll(r.Body)
	var payload struct {
		Config string `json:"config"`
	}
	if err := json.Unmarshal(bodyBytes, &payload); err != nil {
		// fallback: treat entire body as raw config text
		payload.Config = string(bodyBytes)
	}

	log.Printf("[Simulate] Received config (length: %d)", len(payload.Config))

	// Run actual simulation
	result, err := runSimulation(payload.Config)
	if err != nil {
		log.Printf("[Simulate] Error: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"error":   err.Error(),
			"message": "Simulation failed",
		})
		return
	}

	q := r.URL.Query()
	if q.Get("sync") == "true" {
		// return result directly
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(result)
		return
	}

	// async: create job and immediately mark done
	id := fmt.Sprintf("job-%d", rand.Intn(1_000_000))
	j := &job{ID: id, Status: statusDone, Result: result}
	jobsMu.Lock()
	jobs[id] = j
	jobsMu.Unlock()

	log.Printf("[Simulate] Created job: %s", id)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(w).Encode(map[string]string{"job_id": id})
}

func resultHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, `{"error":"id required"}`, http.StatusBadRequest)
		return
	}
	jobsMu.Lock()
	j, ok := jobs[id]
	jobsMu.Unlock()
	if !ok {
		http.Error(w, `{"error":"job not found"}`, http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(j)
}

func runSimulation(config string) (map[string]interface{}, error) {
	if config == "" {
		return nil, fmt.Errorf("empty configuration")
	}

	// Create temporary config file
	tmpFile, err := os.CreateTemp("", "gcsim-config-*.txt")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(config)
	if err != nil {
		tmpFile.Close()
		return nil, fmt.Errorf("failed to write config: %w", err)
	}
	tmpFile.Close()

	log.Printf("[Simulate] Running simulation with config file: %s", tmpFile.Name())

	// Run simulation
	ctx := context.Background()
	opts := simulator.Options{
		ConfigPath: tmpFile.Name(),
	}

	result, err := simulator.Run(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("simulation failed: %w", err)
	}

	log.Printf("[Simulate] Simulation completed successfully")

	// Convert result to map for JSON encoding
	// The simulator.Result is already compatible with JSON marshaling
	resultBytes, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	var resultMap map[string]interface{}
	err = json.Unmarshal(resultBytes, &resultMap)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal result: %w", err)
	}

	return resultMap, nil
}

func sampleResult() map[string]interface{} {
	// minimal fields matching server/api.md
	return map[string]interface{}{
		"summary": map[string]interface{}{"dps": 12345.6, "duration": 60},
		"characters": []map[string]interface{}{
			{"name": "Diluc", "level": 90, "hp": 12000, "atk": 900, "def": 300, "dps": 4567.8, "equipment": []string{"Wolf's Gravestone", "Gladiator"}},
		},
		"dps_samples": []int{4000, 4100, 4200, 4300, 4400, 4500},
		"timeline":    []map[string]interface{}{{"t": 0, "energy": 40}, {"t": 1, "energy": 60}, {"t": 2, "energy": 80}},
	}
}
