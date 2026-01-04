package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/Karashina/gcsim-unofficial-clone/internal/auth"
	"github.com/Karashina/gcsim-unofficial-clone/internal/db"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/optimization"
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
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next(w, r)
	})
}

func main() {
	rand.Seed(time.Now().UnixNano())

	// データベース初期化
	dbConfig := &db.Config{
		DBPath: getEnv("GCSIM_DB_PATH", "./data/gcsim.db"),
		Debug:  getEnv("GCSIM_DB_DEBUG", "false") == "true",
	}
	database, err := db.NewDatabase(dbConfig)
	if err != nil {
		log.Fatalf("[DB] Failed to initialize database: %v", err)
	}
	log.Println("[DB] Database initialized")

	// 管理者ユーザーの初期化
	adminUsername := getEnv("GCSIM_ADMIN_USERNAME", "admin")
	adminPassword := getEnv("GCSIM_ADMIN_PASSWORD", "admin")

	if err := database.InitializeAdminUser(adminUsername, adminPassword); err != nil {
		log.Fatalf("[DB] Failed to initialize admin user: %v", err)
	}

	// 認証設定の読み込み
	authConfig := auth.Config{
		JWTSecret:     []byte(getEnv("GCSIM_JWT_SECRET", "default-secret-key-change-in-production")),
		TokenDuration: 24 * time.Hour,
	}

	// レート制限の設定（1秒間に10リクエスト、バースト20）
	rateLimiter := auth.NewRateLimiter(10, 20)

	log.Printf("[Auth] JWT authentication enabled")
	log.Printf("[Auth] Admin user: %s", adminUsername)

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

	// 認証不要エンドポイント
	http.HandleFunc("/api/register", corsHandler(auth.RegisterHandler(database.UserRepository, authConfig)))
	http.HandleFunc("/api/login", corsHandler(auth.LoginHandlerWithDB(database.UserRepository, authConfig)))
	http.HandleFunc("/api/verify", corsHandler(verifyHandler(authConfig)))

	// 認証が必要なエンドポイント（レート制限付き）
	http.Handle("/api/simulate", rateLimitMiddleware(rateLimiter, auth.Middleware(authConfig)(http.HandlerFunc(corsHandler(simulateHandler)))))
	http.Handle("/api/optimize", rateLimitMiddleware(rateLimiter, auth.Middleware(authConfig)(http.HandlerFunc(corsHandler(optimizeHandler)))))
	http.HandleFunc("/api/result", corsHandler(resultHandler))

	// 管理者専用エンドポイント
	http.Handle("/api/admin/users", auth.Middleware(authConfig)(auth.AdminMiddleware(http.HandlerFunc(corsHandler(auth.ListUsersHandler(database.UserRepository))))))
	http.Handle("/api/admin/users/pending", auth.Middleware(authConfig)(auth.AdminMiddleware(http.HandlerFunc(corsHandler(auth.ListPendingUsersHandler(database.UserRepository))))))
	http.Handle("/api/admin/users/approve", auth.Middleware(authConfig)(auth.AdminMiddleware(http.HandlerFunc(corsHandler(auth.ApproveUserHandler(database.UserRepository))))))

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

// getEnv 環境変数を取得（デフォルト値付き）
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// maskPassword パスワードをマスク表示
func maskPassword(password string) string {
	if len(password) <= 2 {
		return "***"
	}
	return string(password[0]) + strings.Repeat("*", len(password)-2) + string(password[len(password)-1])
}

// rateLimitMiddleware レート制限ミドルウェア
func rateLimitMiddleware(limiter *auth.RateLimiter, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// クライアントIPを取得
		ip := r.RemoteAddr
		if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
			ip = strings.Split(forwardedFor, ",")[0]
		}

		if !limiter.Allow(ip) {
			log.Printf("[RateLimit] Rate limit exceeded for IP: %s", ip)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(map[string]string{
				"error":   "rate_limit_exceeded",
				"message": "リクエストが多すぎます。しばらく待ってから再試行してください。",
			})
			return
		}

		next.ServeHTTP(w, r)
	})
}

// verifyHandler トークン検証エンドポイント
func verifyHandler(authConfig auth.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodPost {
			http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
			return
		}

		// Authorization ヘッダーからトークン取得
		authHeader := r.Header.Get("Authorization")
		token := ""
		if authHeader != "" {
			parts := strings.Split(authHeader, " ")
			if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
				token = parts[1]
			}
		}

		if token == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"valid":   false,
				"message": "トークンがありません",
			})
			return
		}

		// トークン検証
		claims, err := auth.ValidateToken(token, authConfig)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"valid":   false,
				"message": "トークンが無効または期限切れです",
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"valid":    true,
			"user_id":  claims.UserID,
			"username": claims.Username,
			"role":     claims.Role,
			"message":  "トークンは有効です",
		})
	}
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

func optimizeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	bodyBytes, _ := io.ReadAll(r.Body)
	var payload struct {
		Config  string `json:"config"`
		Options string `json:"options"`
	}
	if err := json.Unmarshal(bodyBytes, &payload); err != nil {
		payload.Config = string(bodyBytes)
	}

	log.Printf("[Optimize] Received config (length: %d)", len(payload.Config))

	if payload.Config == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"error":   "empty configuration",
			"message": "Configuration is required",
		})
		return
	}

	// Run optimizer
	optimizedConfig, stats, err := runOptimizer(payload.Config, payload.Options)
	if err != nil {
		log.Printf("[Optimize] Error: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"error":   err.Error(),
			"message": "Optimization failed",
		})
		return
	}

	log.Printf("[Optimize] Optimization completed successfully")

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"optimized_config": optimizedConfig,
		"statistics":       stats,
	})
}

func runOptimizer(config string, additionalOptions string) (string, map[string]interface{}, error) {
	// Create temporary config file
	tmpFile, err := os.CreateTemp("", "gcsim-optim-*.txt")
	if err != nil {
		return "", nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	_, err = tmpFile.WriteString(config)
	if err != nil {
		tmpFile.Close()
		return "", nil, fmt.Errorf("failed to write config: %w", err)
	}
	tmpFile.Close()

	// Create output file for optimized config
	outFile, err := os.CreateTemp("", "gcsim-optim-out-*.txt")
	if err != nil {
		return "", nil, fmt.Errorf("failed to create output temp file: %w", err)
	}
	outPath := outFile.Name()
	outFile.Close()
	defer os.Remove(outPath)

	log.Printf("[Optimize] Running optimizer with config file: %s", tmpPath)

	// Capture stdout/stderr
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()
	os.Stdout = wOut
	os.Stderr = wErr

	// Run the optimizer with recovery
	var optimizerErr error
	func() {
		defer func() {
			if r := recover(); r != nil {
				optimizerErr = fmt.Errorf("optimizer panic: %v", r)
			}
		}()

		simopt := simulator.Options{
			ConfigPath:       tmpPath,
			ResultSaveToPath: outPath,
		}
		optimization.RunSubstatOptim(simopt, false, additionalOptions)
	}()

	// Restore stdout/stderr
	wOut.Close()
	wErr.Close()
	os.Stdout = oldStdout
	os.Stderr = oldStderr

	var bufOut, bufErr bytes.Buffer
	io.Copy(&bufOut, rOut)
	io.Copy(&bufErr, rErr)
	rOut.Close()
	rErr.Close()

	if optimizerErr != nil {
		return "", nil, optimizerErr
	}

	// Read optimized config
	optimizedBytes, err := os.ReadFile(outPath)
	if err != nil {
		// If output file doesn't exist, the optimizer might have printed to stdout
		if bufOut.Len() > 0 {
			return bufOut.String(), nil, nil
		}
		return "", nil, fmt.Errorf("failed to read optimized config: %w", err)
	}

	optimizedConfig := string(optimizedBytes)

	// Run simulation with optimized config to get stats
	simResult, err := runSimulation(optimizedConfig)
	if err != nil {
		// Return optimized config even if sim fails
		log.Printf("[Optimize] Simulation with optimized config failed: %v", err)
		return optimizedConfig, nil, nil
	}

	return optimizedConfig, simResult, nil
}
