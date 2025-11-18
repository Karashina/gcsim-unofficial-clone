package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/model"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/simulator"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	redis "github.com/redis/go-redis/v9"
)

// job related types and in-memory store
type jobStatus string

const (
	statusPending jobStatus = "pending"
	statusRunning jobStatus = "running"
	statusDone    jobStatus = "done"
	statusError   jobStatus = "error"
)

type job struct {
	ID        string    `json:"id"`
	Status    jobStatus `json:"status"`
	ResultKey string    `json:"result_key,omitempty"`
	Error     string    `json:"error,omitempty"`
	Created   time.Time `json:"created_at"`
}

type jobStoreType struct {
	mu    sync.RWMutex
	store map[string]*job
}

var jobStore = &jobStoreType{store: make(map[string]*job)}

var (
	redisClient *redis.Client
	workerSem   chan struct{}
	maxWorkers  int
	jobTTL      time.Duration
	s3Client    *minio.Client
	s3Bucket    string
)

func (s *jobStoreType) set(id string, j *job) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if j.Created.IsZero() {
		j.Created = time.Now()
	}
	s.store[id] = j
	// persist to redis if available
	_ = saveJobToRedis(j)
}

func (s *jobStoreType) get(id string) *job {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if j, ok := s.store[id]; ok {
		return j
	}
	return nil
}

func (s *jobStoreType) setStatus(id string, status jobStatus) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if j, ok := s.store[id]; ok {
		j.Status = status
		_ = saveJobToRedis(j)
	}
}

func (s *jobStoreType) setError(id, errMsg string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if j, ok := s.store[id]; ok {
		j.Status = statusError
		j.Error = errMsg
		_ = saveJobToRedis(j)
	}
}

func (s *jobStoreType) setResult(id string, data []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if j, ok := s.store[id]; ok {
		// store result as file on disk (results/<id>.json)
		key := fmt.Sprintf("results/%s.json", id)
		if err := uploadToS3(key, data); err == nil {
			j.ResultKey = key
		} else {
			// fallback: store in redis result key
			if redisClient != nil {
				_ = redisClient.Set(context.Background(), "gcsim:job:result:"+id, data, jobTTL).Err()
			}
		}
		_ = saveJobToRedis(j)
	}
}

// uploadToS3 stores bytes to local results directory under given key
// Note: simplified to local storage only to avoid external dependencies in CI/tests
func uploadToS3(key string, data []byte) error {
	// if configured, try to upload to MinIO/S3
	if s3Client != nil && s3Bucket != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		// PutObject expects an io.Reader
		_, err := s3Client.PutObject(ctx, s3Bucket, key, bytes.NewReader(data), int64(len(data)), minio.PutObjectOptions{ContentType: "application/json"})
		if err == nil {
			return nil
		}
		log.Printf("warning: failed to upload to s3: %v; falling back to local file", err)
	}

	// fallback: store locally
	dir := "results"
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	path := filepath.Join(dir, filepath.Base(key))
	return os.WriteFile(path, data, 0o644)
}

// presignedURLForKey returns an internal API URL to fetch the result
func presignedURLForKey(key string) (string, error) {
	// we return API endpoint to fetch result
	v := url.Values{}
	v.Set("key", key)
	return "/api/resultfile?" + v.Encode(), nil
}

// runSimulator variable allows tests to inject a fake runner
var runSimulator = func(ctx context.Context, opts simulator.Options) (*model.SimulationResult, error) {
	return simulator.Run(ctx, opts)
}

func (s *jobStoreType) delete(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.store, id)
	if redisClient != nil {
		_ = redisClient.Del(context.Background(), "gcsim:job:"+id).Err()
	}
}

// genJobID generates a short unique id using crypto/rand and timestamp
func genJobID() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		// fallback
		return fmt.Sprintf("job-%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b) + fmt.Sprintf("-%x", time.Now().UnixNano())
}

// saveJobToRedis persists a job to redis if available
func saveJobToRedis(j *job) error {
	if redisClient == nil {
		return nil
	}
	ctx := context.Background()
	b, err := json.Marshal(j)
	if err != nil {
		return err
	}
	return redisClient.Set(ctx, "gcsim:job:"+j.ID, b, jobTTL).Err()
}

// loadJobsFromRedis loads persisted jobs into in-memory store at startup
func loadJobsFromRedis() {
	if redisClient == nil {
		return
	}
	ctx := context.Background()
	keys, err := redisClient.Keys(ctx, "gcsim:job:*").Result()
	if err != nil {
		log.Printf("redis keys error: %v", err)
		return
	}
	for _, k := range keys {
		v, err := redisClient.Get(ctx, k).Bytes()
		if err != nil {
			continue
		}
		var j job
		if err := json.Unmarshal(v, &j); err != nil {
			continue
		}
		jobStore.set(j.ID, &j)
	}
}

// initResources initializes worker pool and redis client; called from serve
func initResources() {
	// max workers from env or default to NumCPU
	maxWorkers = runtime.NumCPU()
	if v := os.Getenv("GCSIM_MAX_WORKERS"); v != "" {
		if i, err := strconv.Atoi(v); err == nil && i > 0 {
			maxWorkers = i
		}
	}
	workerSem = make(chan struct{}, maxWorkers)

	// job TTL
	jobTTL = 24 * time.Hour
	if v := os.Getenv("GCSIM_JOB_TTL_HOURS"); v != "" {
		if h, err := strconv.Atoi(v); err == nil && h > 0 {
			jobTTL = time.Duration(h) * time.Hour
		}
	}

	// init redis
	addr := os.Getenv("GCSIM_REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}
	redisClient = redis.NewClient(&redis.Options{Addr: addr})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Printf("warning: cannot connect to redis at %s: %v; continuing without redis persistence", addr, err)
		redisClient = nil
	} else {
		loadJobsFromRedis()
	}

	// init minio/s3 client if configured
	minioEndpoint := os.Getenv("GCSIM_MINIO_ENDPOINT")
	if minioEndpoint != "" {
		accessKey := os.Getenv("GCSIM_MINIO_ACCESS_KEY")
		secretKey := os.Getenv("GCSIM_MINIO_SECRET_KEY")
		useSSL := false
		if os.Getenv("GCSIM_MINIO_USE_SSL") == "true" {
			useSSL = true
		}
		// initialize client
		client, err := minio.New(minioEndpoint, &minio.Options{
			Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
			Secure: useSSL,
		})
		if err != nil {
			log.Printf("warning: could not initialize minio client: %v; continuing without s3 persistence", err)
		} else {
			s3Client = client
			s3Bucket = os.Getenv("GCSIM_MINIO_BUCKET")
			// ensure bucket exists
			ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel2()
			if s3Bucket != "" {
				if err := s3Client.MakeBucket(ctx2, s3Bucket, minio.MakeBucketOptions{}); err != nil {
					// ignore error; could already exist or permission issue
					log.Printf("warning: MakeBucket returned: %v", err)
				}
			}
		}
	}

	// start cleanup goroutine
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			cutoff := time.Now().Add(-jobTTL)
			jobStore.mu.Lock()
			for id, j := range jobStore.store {
				if j.Created.Before(cutoff) {
					delete(jobStore.store, id)
					if redisClient != nil {
						_ = redisClient.Del(context.Background(), "gcsim:job:"+id).Err()
					}
				}
			}
			jobStore.mu.Unlock()
		}
	}()
}

func serve(
	connectionsClosed chan struct{},
	resultPath string,
	hash string,
	samplePath string,
	keepAlive bool,
) {
	server := &http.Server{Addr: address}
	done := make(chan bool)

	// initialize resources (worker pool, redis)
	initResources()

	http.HandleFunc("/data", func(resp http.ResponseWriter, req *http.Request) {
		success := handleResults(resp, req, resultPath, hash)
		if success && !keepAlive {
			done <- true
		}
	})

	http.HandleFunc("/sample", func(resp http.ResponseWriter, req *http.Request) {
		success := handleSample(resp, req, samplePath)
		if success && !keepAlive {
			done <- true
		}
	})

	// register handlers
	http.HandleFunc("/api/simulate", simulateHandler)

	http.HandleFunc("/api/result", resultHandler)
	http.HandleFunc("/api/resultfile", resultFileHandler)

	// serve web UI: prefer built files in webui/dist, otherwise serve webui/ for dev
	var uiDir string
	if _, err := os.Stat("webui/dist"); err == nil {
		uiDir = "webui/dist"
	} else {
		uiDir = "webui"
	}
	fs := http.FileServer(http.Dir(uiDir))
	http.Handle("/ui/", http.StripPrefix("/ui/", fs))
	// redirect root to UI
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// if the request is already for the UI root, serve it
		if r.URL.Path == "/" {
			http.Redirect(w, r, "/ui/", http.StatusFound)
			return
		}
		// otherwise, pass through to default mux (could be 404)
		http.NotFound(w, r)
	})

	go interruptShutdown(server, done, connectionsClosed)
	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("HTTP server ListenAndSever Error: %v", err)
		}
	}()
}

func handleResults(resp http.ResponseWriter, req *http.Request, path, hash string) bool {
	if req.Method == http.MethodOptions {
		log.Println("OPTIONS request received, responding...")
		optionsResponse(resp)
		return false
	}

	if req.Method != http.MethodGet {
		log.Printf("Invalid request method: %v\n", req.Method)
		resp.WriteHeader(http.StatusForbidden)
		return false
	}

	compressed, err := os.ReadFile(path)
	if err != nil {
		log.Printf("error reading gz data: %v\n", err)
		resp.WriteHeader(http.StatusInternalServerError)
		return false
	}

	log.Println("Received results request, sending response...")
	resp.Header().Set("Content-Type", "application/json")
	resp.Header().Set("Content-Encoding", "deflate")
	resp.Header().Set("Access-Control-Allow-Origin", "*")
	resp.Header().Set("Access-Control-Expose-Headers", "X-GCSIM-SHARE-AUTH")
	resp.Header().Set("X-GCSIM-SHARE-AUTH", hash)
	resp.WriteHeader(http.StatusOK)
	resp.Write(compressed)

	if f, ok := resp.(http.Flusher); ok {
		f.Flush()
	}
	return true
}

func handleSample(resp http.ResponseWriter, req *http.Request, path string) bool {
	if req.Method == http.MethodOptions {
		log.Println("OPTIONS request received, responding...")
		optionsResponse(resp)
		return false
	}

	if req.Method != http.MethodGet {
		log.Printf("Invalid request method: %v\n", req.Method)
		resp.WriteHeader(http.StatusForbidden)
		return false
	}

	compressed, err := os.ReadFile(path)
	if err != nil {
		log.Printf("error reading gz data: %v\n", err)
		resp.WriteHeader(http.StatusInternalServerError)
		return false
	}

	log.Println("Received sample request, sending response...")
	resp.Header().Set("Content-Type", "application/json")
	resp.Header().Set("Content-Encoding", "deflate")
	resp.Header().Set("Access-Control-Allow-Origin", "*")
	resp.WriteHeader(http.StatusOK)
	resp.Write(compressed)

	if f, ok := resp.(http.Flusher); ok {
		f.Flush()
	}
	return true
}

func optionsResponse(resp http.ResponseWriter) {
	resp.Header().Set("Access-Control-Allow-Origin", "*")
	resp.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	resp.Header().Set(
		"Access-Control-Allow-Headers",
		"Accept, Access-Control-Allow-Origin, Content-Type, "+
			"Content-Length, Accept-Encoding, Authorization")
	resp.WriteHeader(http.StatusNoContent)
}

// simulateHandler handles POST /api/simulate (synchronous only, 180s timeout)
func simulateHandler(resp http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodOptions {
		optionsResponse(resp)
		return
	}
	if req.Method != http.MethodPost {
		resp.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		log.Printf("error reading request body: %v", err)
		http.Error(resp, `{"error":"cannot read request body"}`, http.StatusBadRequest)
		return
	}
	var payload struct {
		Config  string                 `json:"config"`
		Options map[string]interface{} `json:"options,omitempty"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		log.Printf("invalid json payload: %v", err)
		http.Error(resp, `{"error":"invalid json payload"}`, http.StatusBadRequest)
		return
	}
	if payload.Config == "" {
		http.Error(resp, `{"error":"config is required"}`, http.StatusBadRequest)
		return
	}

	tmpDir := os.TempDir()
	tmpFile := filepath.Join(tmpDir, "gcsim_webui_config_"+time.Now().Format("20060102_150405")+".txt")
	if err := os.WriteFile(tmpFile, []byte(payload.Config), 0o600); err != nil {
		log.Printf("error writing temp config: %v", err)
		http.Error(resp, `{"error":"cannot write temp config"}`, http.StatusInternalServerError)
		return
	}
	defer os.Remove(tmpFile)

	// 180s timeout enforced
	ctx, cancel := context.WithTimeout(req.Context(), 180*time.Second)
	defer cancel()

	opts := simulator.Options{ConfigPath: tmpFile}
	// acquire worker
	workerSem <- struct{}{}
	result, err := runSimulator(ctx, opts)
	<-workerSem
	if err != nil {
		log.Printf("simulator.Run error: %v", err)
		if errors.Is(err, context.DeadlineExceeded) || ctx.Err() == context.DeadlineExceeded {
			resp.Header().Set("Content-Type", "application/json")
			resp.Header().Set("Access-Control-Allow-Origin", "*")
			resp.WriteHeader(http.StatusGatewayTimeout)
			_ = json.NewEncoder(resp).Encode(map[string]string{"error": "timeout", "message": "simulation exceeded 180s timeout"})
			return
		}
		// If error exposes parse details via an Errors() method, include them in response.
		// Use a structural type assertion to avoid depending on a concrete parserpkg type name.
		if pe, ok := err.(interface {
			Error() string
			Errors() interface{}
		}); ok {
			out := map[string]interface{}{
				"error":        "parse error",
				"message":      pe.Error(),
				"parse_errors": pe.Errors(),
			}
			resp.Header().Set("Content-Type", "application/json")
			resp.Header().Set("Access-Control-Allow-Origin", "*")
			resp.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(resp).Encode(out)
			return
		}
		// fallback: try to extract lnN patterns
		errStr := err.Error()
		re := regexp.MustCompile(`ln(\d+):\s*(.+)`)
		matches := re.FindAllStringSubmatch(errStr, -1)
		if len(matches) > 0 {
			out := map[string]interface{}{
				"error":   "parse error",
				"message": errStr,
			}
			pe := make([]map[string]interface{}, 0, len(matches))
			for _, m := range matches {
				line := 0
				if v, e := strconv.Atoi(m[1]); e == nil {
					line = v
				}
				pe = append(pe, map[string]interface{}{"line": line, "message": m[2]})
			}
			out["parse_errors"] = pe
			resp.Header().Set("Content-Type", "application/json")
			resp.Header().Set("Access-Control-Allow-Origin", "*")
			resp.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(resp).Encode(out)
			return
		}
		http.Error(resp, `{"error":"simulation failed: `+errStr+`"}`, http.StatusInternalServerError)
		return
	}
	data, err := result.MarshalJSON()
	if err != nil {
		log.Printf("error marshalling result: %v", err)
		http.Error(resp, `{"error":"cannot marshal result"}`, http.StatusInternalServerError)
		return
	}
	resp.Header().Set("Content-Type", "application/json")
	resp.Header().Set("Access-Control-Allow-Origin", "*")
	resp.WriteHeader(http.StatusOK)
	resp.Write(data)
}

// resultHandler handles GET /api/result?id=<jobid>
func resultHandler(resp http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodOptions {
		optionsResponse(resp)
		return
	}
	if req.Method != http.MethodGet {
		resp.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	id := req.URL.Query().Get("id")
	if id == "" {
		http.Error(resp, `{"error":"id is required"}`, http.StatusBadRequest)
		return
	}
	j := jobStore.get(id)
	if j == nil {
		http.Error(resp, `{"error":"job not found"}`, http.StatusNotFound)
		return
	}
	resp.Header().Set("Content-Type", "application/json")
	resp.Header().Set("Access-Control-Allow-Origin", "*")
	out := map[string]interface{}{"id": j.ID, "status": j.Status}
	if j.Status == statusDone {
		if j.ResultKey != "" {
			if u, err := presignedURLForKey(j.ResultKey); err == nil {
				out["result_url"] = u
			}
		} else if redisClient != nil {
			if data, err := redisClient.Get(context.Background(), "gcsim:job:result:"+j.ID).Bytes(); err == nil {
				out["result"] = json.RawMessage(data)
			}
		}
	}
	if j.Status == statusError {
		out["error"] = j.Error
	}
	_ = json.NewEncoder(resp).Encode(out)
}

// resultFileHandler returns local result files
func resultFileHandler(resp http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodOptions {
		optionsResponse(resp)
		return
	}
	if req.Method != http.MethodGet {
		resp.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	key := req.URL.Query().Get("key")
	if key == "" {
		http.Error(resp, `{"error":"key required"}`, http.StatusBadRequest)
		return
	}
	// restrict to results dir
	path := filepath.Join("results", filepath.Base(key))
	b, err := os.ReadFile(path)
	if err != nil {
		http.Error(resp, `{"error":"file not found"}`, http.StatusNotFound)
		return
	}
	resp.Header().Set("Content-Type", "application/json")
	resp.Header().Set("Access-Control-Allow-Origin", "*")
	resp.WriteHeader(http.StatusOK)
	resp.Write(b)
}

func interruptShutdown(server *http.Server, done chan bool, connectionsClosed chan struct{}) {
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)

	select {
	case <-done:
	case <-sigint:
	}

	if err := server.Shutdown(context.Background()); err != nil {
		log.Printf("HTTP server Shutdown Error: %v", err)
	}
	close(connectionsClosed)
}
