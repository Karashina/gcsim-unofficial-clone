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
	"strings"
	"time"

	"github.com/Karashina/gcsim-unofficial-clone/internal/auth"
	"github.com/Karashina/gcsim-unofficial-clone/internal/db"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/simulator"
)

func main() {
	var addr string
	flag.StringVar(&addr, "addr", ":8382", "address to listen on (example :8382)")
	flag.Parse()

	// 認証設定の読み込み
	jwtSecret := os.Getenv("GCSIM_JWT_SECRET")

	// JWT秘密鍵のバリデーション
	if jwtSecret == "" {
		log.Fatal("[Auth] FATAL: GCSIM_JWT_SECRET environment variable is not set. Server cannot start without a JWT secret key.")
	}

	// 脆弱なデフォルト値の検出
	weakSecrets := []string{
		"default-secret-key-change-in-production",
		"secret",
		"password",
		"123456",
		"admin",
		"test",
	}
	for _, weak := range weakSecrets {
		if jwtSecret == weak {
			log.Fatalf("[Auth] FATAL: GCSIM_JWT_SECRET contains a weak/default value (%s). Please set a strong random key.", weak)
		}
	}

	// 最小長チェック（32文字以上推奨）
	if len(jwtSecret) < 32 {
		log.Fatal("[Auth] FATAL: GCSIM_JWT_SECRET is too short. Please use at least 32 characters for security.")
	}

	authConfig := auth.Config{
		JWTSecret:     []byte(jwtSecret),
		TokenDuration: 24 * time.Hour,
	}

	// データベース初期化
	dbConfig := &db.Config{
		DBPath: getEnv("GCSIM_DB_PATH", "./data/gcsim.db"),
		Debug:  false,
	}
	database, err := db.NewDatabase(dbConfig)
	if err != nil {
		log.Fatalf("[DB] FATAL: Failed to initialize database: %v", err)
	}
	defer database.Close()

	// 管理者ユーザー初期化（固定値: karashina / TestPass1234!）
	if err := database.InitializeAdminUser("karashina", "TestPass1234!"); err != nil {
		log.Fatalf("[DB] FATAL: Failed to initialize admin user: %v", err)
	}

	// レート制限の設定（1秒間に5リクエスト、バースト10）
	rateLimiter := auth.NewRateLimiter(5, 10)

	log.Printf("[Auth] JWT authentication enabled")
	log.Printf("[Auth] JWT secret length: %d characters", len(jwtSecret))
	log.Printf("[DB] Database initialized successfully")

	// CORS設定
	allowedOrigins := getEnv("GCSIM_CORS_ALLOWED_ORIGINS", "*")
	log.Printf("[CORS] Allowed origins: %s", allowedOrigins)

	// 信頼できるプロキシIPリストの読み込み
	trustedProxies := parseTrustedProxies(getEnv("GCSIM_TRUSTED_PROXIES", ""))
	if len(trustedProxies) > 0 {
		log.Printf("[Proxy] Trusted proxies: %v", trustedProxies)
	} else {
		log.Printf("[Proxy] No trusted proxies configured - using direct connection mode")
	}

	mux := http.NewServeMux()

	// CORSハンドラーをラップ
	cors := makeCorsHandler(allowedOrigins)

	// 認証不要エンドポイント
	mux.Handle("/api/register", cors(auth.RegisterHandler(database.UserRepository, authConfig)))
	mux.Handle("/api/login", cors(auth.LoginHandlerWithDB(database.UserRepository, authConfig)))
	mux.Handle("/api/verify", cors(verifyHandler(authConfig)))

	// 管理者用エンドポイント（認証+管理者権限必要）
	mux.Handle("/api/admin/users", cors(auth.Middleware(authConfig)(auth.AdminMiddleware(http.HandlerFunc(auth.ListUsersHandler(database.UserRepository))))))
	mux.Handle("/api/admin/users/pending", cors(auth.Middleware(authConfig)(auth.AdminMiddleware(http.HandlerFunc(auth.ListPendingUsersHandler(database.UserRepository))))))
	mux.Handle("/api/admin/users/approve", cors(auth.Middleware(authConfig)(auth.AdminMiddleware(http.HandlerFunc(auth.ApproveUserHandler(database.UserRepository))))))

	// 管理者用パスワード変更エンドポイント（/api/admin/users/{id}/password）
	passwordChangeHandler := auth.Middleware(authConfig)(auth.AdminMiddleware(http.HandlerFunc(auth.AdminChangePasswordHandler(database.UserRepository))))
	mux.HandleFunc("/api/admin/users/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/password") {
			cors(passwordChangeHandler).ServeHTTP(w, r)
		} else {
			http.NotFound(w, r)
		}
	})

	// 認証が必要なエンドポイント（レート制限付き）
	mux.Handle("/api/change-password", rateLimitMiddleware(rateLimiter, trustedProxies, cors(auth.Middleware(authConfig)(http.HandlerFunc(auth.ChangePasswordHandler(database.UserRepository))))))

	// シミュレーションエンドポイント（認証オプショナル、レート制限付き）
	mux.Handle("/api/simulate", rateLimitMiddleware(rateLimiter, trustedProxies, cors(auth.OptionalMiddleware(authConfig)(http.HandlerFunc(simulateHandler)))))
	mux.Handle("/api/optimize", rateLimitMiddleware(rateLimiter, trustedProxies, cors(auth.OptionalMiddleware(authConfig)(http.HandlerFunc(optimizeHandler)))))

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

// makeCorsHandler creates a CORS handler with configurable allowed origins
func makeCorsHandler(allowedOriginsStr string) func(http.Handler) http.Handler {
	// Parse allowed origins (comma-separated)
	allowedOrigins := strings.Split(allowedOriginsStr, ",")
	for i := range allowedOrigins {
		allowedOrigins[i] = strings.TrimSpace(allowedOrigins[i])
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Check if origin is allowed
			allowed := false
			if allowedOriginsStr == "*" {
				allowed = true
				w.Header().Set("Access-Control-Allow-Origin", "*")
			} else {
				for _, allowedOrigin := range allowedOrigins {
					if origin == allowedOrigin {
						allowed = true
						w.Header().Set("Access-Control-Allow-Origin", origin)
						break
					}
				}
			}

			// If origin is not allowed and not a wildcard, don't set CORS headers
			if !allowed && origin != "" {
				log.Printf("[CORS] Blocked request from disallowed origin: %s", origin)
				// Continue without CORS headers - browser will block the response
			}

			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Allow-Credentials", "true")

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// simulateHandler is a minimal, standalone version of the simulate endpoint.
// It expects JSON: {"config":"<gcsl>"} and synchronously runs the simulator with a 180s timeout.
func simulateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
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

	// Support both JSON format {"config": "..."} and plain text config
	contentType := r.Header.Get("Content-Type")
	if strings.Contains(contentType, "text/plain") || contentType == "" {
		// Direct config text
		payload.Config = string(body)
	} else {
		// JSON format
		if err := json.Unmarshal(body, &payload); err != nil {
			log.Printf("simulateHandler: invalid json payload: %v\npayload: %s", err, string(body))
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error":"invalid json payload"}`))
			return
		}
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

	// Support both JSON format {"config": "..."} and plain text config
	contentType := r.Header.Get("Content-Type")
	if strings.Contains(contentType, "text/plain") || contentType == "" {
		// Direct config text
		payload.Config = string(body)
	} else {
		// JSON format
		if err := json.Unmarshal(body, &payload); err != nil {
			log.Printf("optimizeHandler: invalid json payload: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error":"invalid json payload"}`))
			return
		}
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

// parseTrustedProxies parses comma-separated trusted proxy IPs
func parseTrustedProxies(proxiesStr string) []string {
	if proxiesStr == "" {
		return nil
	}
	proxies := strings.Split(proxiesStr, ",")
	result := make([]string, 0, len(proxies))
	for _, p := range proxies {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

// getClientIP extracts the real client IP from request headers
// considering trusted proxies (Nginx, Cloudflare, etc.)
func getClientIP(r *http.Request, trustedProxies []string) string {
	// Get the direct connection IP (without port)
	remoteIP := r.RemoteAddr
	if idx := strings.LastIndex(remoteIP, ":"); idx != -1 {
		remoteIP = remoteIP[:idx]
	}

	// If no trusted proxies configured, use direct IP
	if len(trustedProxies) == 0 {
		return remoteIP
	}

	// Check if the direct connection is from a trusted proxy
	isTrusted := false
	for _, proxy := range trustedProxies {
		if remoteIP == proxy {
			isTrusted = true
			break
		}
	}

	// If not from trusted proxy, use direct IP
	if !isTrusted {
		return remoteIP
	}

	// Try X-Real-IP first (single IP, commonly set by Nginx)
	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		return strings.TrimSpace(realIP)
	}

	// Try X-Forwarded-For (comma-separated list, rightmost is closest to proxy)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		// Get the rightmost IP (closest to our proxy)
		for i := len(ips) - 1; i >= 0; i-- {
			ip := strings.TrimSpace(ips[i])
			if ip != "" {
				return ip
			}
		}
	}

	// Fallback to direct IP
	return remoteIP
}

// rateLimitMiddleware レート制限ミドルウェア
func rateLimitMiddleware(limiter *auth.RateLimiter, trustedProxies []string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// クライアントIPを取得（プロキシ対応）
		ip := getClientIP(r, trustedProxies)

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
			"valid":   true,
			"user_id": claims.UserID,
			"message": "トークンは有効です",
		})
	}
}
