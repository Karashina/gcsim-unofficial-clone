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
	"sync"
	"time"
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

func main() {
	rand.Seed(time.Now().UnixNano())
	// static UI dir
	uiDir := "webui/dist"
	if _, err := os.Stat(uiDir); os.IsNotExist(err) {
		uiDir = "webui"
	}

	fs := http.FileServer(http.Dir(uiDir))
	http.Handle("/ui/", http.StripPrefix("/ui/", fs))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/ui/", http.StatusFound)
	})

	http.HandleFunc("/api/simulate", simulateHandler)
	http.HandleFunc("/api/result", resultHandler)

	addr := ":8381"
	srv := &http.Server{Addr: addr}
	go func() {
		log.Printf("devserver listening on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	// wait for interrupt
	c := make(chan os.Signal, 1)
	<-c
	_ = srv.Shutdown(context.Background())
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
	q := r.URL.Query()
	if q.Get("sync") == "true" {
		// return a fake SimulationResult-like object
		res := sampleResult()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(res)
		return
	}
	// async: create job and immediately mark done (for dev)
	id := fmt.Sprintf("job-%d", rand.Intn(1_000_000))
	res := sampleResult()
	j := &job{ID: id, Status: statusDone, Result: res}
	jobsMu.Lock()
	jobs[id] = j
	jobsMu.Unlock()
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
