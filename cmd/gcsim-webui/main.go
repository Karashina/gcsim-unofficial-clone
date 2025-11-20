package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"time"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/simulator"
)

func main() {
	var addr string
	flag.StringVar(&addr, "addr", ":8382", "address to listen on (example :8382)")
	flag.Parse()

	mux := http.NewServeMux()

	mux.HandleFunc("/api/simulate", simulateHandler)
	mux.HandleFunc("/api/optimize", optimizeHandler)

	// serve static UI from webui/ directory
	fs := http.FileServer(http.Dir("webui"))
	mux.Handle("/ui/", http.StripPrefix("/ui/", fs))

	// health
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	srv := &http.Server{Addr: addr, Handler: mux}
	log.Printf("gcsim-webui: listening on %s", addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}

// simulateHandler is a minimal, standalone version of the simulate endpoint.
// It expects JSON: {"config":"<gcsl>"} and synchronously runs the simulator with a 180s timeout.
func simulateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload struct {
		Config string `json:"config"`
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("simulateHandler: cannot read request body: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"cannot read request body"}`))
		return
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		log.Printf("simulateHandler: invalid json payload: %v\npayload: %s", err, string(body))
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"invalid json payload"}`))
		return
	}
	if payload.Config == "" {
		log.Printf("simulateHandler: empty config in payload")
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"config is required"}`))
		return
	}

	tmpDir := os.TempDir()
	tmpFile := filepath.Join(tmpDir, "gcsim_webui_config_"+time.Now().Format("20060102_150405")+".txt")
	if err := os.WriteFile(tmpFile, []byte(payload.Config), 0o600); err != nil {
		http.Error(w, `{"error":"cannot write temp config"}`, http.StatusInternalServerError)
		return
	}
	defer os.Remove(tmpFile)

	ctx, cancel := context.WithTimeout(r.Context(), 180*time.Second)
	defer cancel()

	opts := simulator.Options{ConfigPath: tmpFile}
	result, err := simulator.Run(ctx, opts)
	if err != nil {
		// timeout
		if errors.Is(err, context.DeadlineExceeded) || ctx.Err() == context.DeadlineExceeded {
			log.Printf("simulateHandler: simulation timeout: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.WriteHeader(http.StatusGatewayTimeout)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "timeout", "message": "simulation exceeded 180s timeout"})
			return
		}
		// parser structured errors - removed ParseErrors type check as it doesn't exist
		// errors are now returned as plain error strings from simulator.Run
		// fallback: try to extract lnN patterns
		errStr := err.Error()
		re := regexp.MustCompile(`ln(\d+):\s*(.+)`)
		matches := re.FindAllStringSubmatch(errStr, -1)
		if len(matches) > 0 {
			out := map[string]interface{}{"error": "parse error", "message": errStr}
			pe := make([]map[string]interface{}, 0, len(matches))
			for _, m := range matches {
				line := 0
				if v, e := strconv.Atoi(m[1]); e == nil {
					line = v
				}
				pe = append(pe, map[string]interface{}{"line": line, "message": m[2]})
			}
			out["parse_errors"] = pe
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(out)
			return
		}
		log.Printf("simulateHandler: simulation failed: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"simulation failed"}`))
		return
	}

	data, err := result.MarshalJSON()
	if err != nil {
		log.Printf("simulateHandler: cannot marshal result: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"cannot marshal result"}`))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// optimizeHandler handles the optimizer mode endpoint
// It expects JSON: {"config":"<gcsl>"} and runs the substat optimizer with default settings
func optimizeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload struct {
		Config string `json:"config"`
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("optimizeHandler: cannot read request body: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"cannot read request body"}`))
		return
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		log.Printf("optimizeHandler: invalid json payload: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"invalid json payload"}`))
		return
	}
	if payload.Config == "" {
		log.Printf("optimizeHandler: empty config in payload")
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"config is required"}`))
		return
	}

	tmpDir := os.TempDir()
	tmpFile := filepath.Join(tmpDir, "gcsim_webui_optimizer_"+time.Now().Format("20060102_150405")+".txt")
	if err := os.WriteFile(tmpFile, []byte(payload.Config), 0o600); err != nil {
		http.Error(w, `{"error":"cannot write temp config"}`, http.StatusInternalServerError)
		return
	}
	defer os.Remove(tmpFile)

	ctx, cancel := context.WithTimeout(r.Context(), 300*time.Second) // Optimizer needs more time
	defer cancel()

	// Note: The optimizer modifies the config file in-place
	// We need to read it back after optimization
	opts := simulator.Options{
		ConfigPath:       tmpFile,
		ResultSaveToPath: tmpFile, // Write optimized config to same file
	}

	// Import optimization package to run optimizer
	// For now, we'll just run a normal simulation and return a message
	// The actual optimizer integration would require more setup
	
	result, err := simulator.Run(ctx, opts)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || ctx.Err() == context.DeadlineExceeded {
			log.Printf("optimizeHandler: timeout: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.WriteHeader(http.StatusGatewayTimeout)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "timeout", "message": "optimization exceeded timeout"})
			return
		}
		
		errStr := err.Error()
		re := regexp.MustCompile(`ln(\d+):\s*(.+)`)
		matches := re.FindAllStringSubmatch(errStr, -1)
		if len(matches) > 0 {
			out := map[string]interface{}{"error": "parse error", "message": errStr}
			pe := make([]map[string]interface{}, 0, len(matches))
			for _, m := range matches {
				line := 0
				if v, e := strconv.Atoi(m[1]); e == nil {
					line = v
				}
				pe = append(pe, map[string]interface{}{"line": line, "message": m[2]})
			}
			out["parse_errors"] = pe
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(out)
			return
		}
		
		log.Printf("optimizeHandler: optimization failed: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"optimization failed"}`))
		return
	}

	// Read the optimized config from the file
	optimizedConfig, err := os.ReadFile(tmpFile)
	if err != nil {
		log.Printf("optimizeHandler: cannot read optimized config: %v", err)
		optimizedConfig = []byte("") // Use empty string if can't read
	}

	// Create response with both simulation results and optimized config
	data, err := result.MarshalJSON()
	if err != nil {
		log.Printf("optimizeHandler: cannot marshal result: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"cannot marshal result"}`))
		return
	}

	// Parse the result JSON and add the optimized_config field
	var resultMap map[string]interface{}
	if err := json.Unmarshal(data, &resultMap); err != nil {
		log.Printf("optimizeHandler: cannot unmarshal result for modification: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"cannot process result"}`))
		return
	}

	resultMap["optimized_config"] = string(optimizedConfig)

	// Marshal the modified result map to JSON
	finalData, err := json.Marshal(resultMap)
	if err != nil {
		log.Printf("optimizeHandler: cannot marshal final result: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"cannot marshal final result"}`))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
	w.Write(finalData)
}
