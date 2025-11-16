package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/model"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/simulator"
)

// helper: write minimal dummy result json
func writeDummyResultFile(t *testing.T, id string, content string) string {
	dir := "results"
	err := os.MkdirAll(dir, 0o755)
	if err != nil {
		t.Fatalf("mkdir results: %v", err)
	}
	path := filepath.Join(dir, id+".json")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write dummy result: %v", err)
	}
	return path
}

func TestSimulateSyncSuccess(t *testing.T) {
	// mock runSimulator to return a tiny SimulationResult
	orig := runSimulator
	defer func() { runSimulator = orig }()
	runSimulator = func(ctx context.Context, opts simulator.Options) (*model.SimulationResult, error) {
		res := &model.SimulationResult{}
		return res, nil
	}

	reqBody := `{"config":"dummy config"}`
	req := httptest.NewRequest(http.MethodPost, "/api/simulate?sync=true", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	simulateHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", rec.Code, rec.Body.String())
	}
	// should be valid json
	var out map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("response not json: %v", err)
	}
}

func TestSimulateAsyncFlow(t *testing.T) {
	// speed up by mocking runSimulator to sleep then return
	orig := runSimulator
	defer func() { runSimulator = orig }()
	runSimulator = func(ctx context.Context, opts simulator.Options) (*model.SimulationResult, error) {
		time.Sleep(50 * time.Millisecond)
		res := &model.SimulationResult{}
		return res, nil
	}

	// cleanup job store/results
	jobStore = &jobStoreType{store: make(map[string]*job)}
	_ = os.RemoveAll("results")

	reqBody := `{"config":"dummy config async"}`
	req := httptest.NewRequest(http.MethodPost, "/api/simulate", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	simulateHandler(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected 202 Accepted, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	jobID := resp["job_id"]
	if jobID == "" {
		t.Fatalf("missing job_id")
	}

	// poll for result
	for i := 0; i < 20; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/result?id="+jobID, nil)
		rec := httptest.NewRecorder()
		resultHandler(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("result handler returned %d", rec.Code)
		}
		var out map[string]interface{}
		if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
			t.Fatalf("invalid json: %v", err)
		}
		if out["status"] == "done" {
			// done; success
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("job did not complete in time")
}

func TestResultFileHandler(t *testing.T) {
	id := "testfile"
	content := `{"ok":true}`
	writeDummyResultFile(t, id, content)

	req := httptest.NewRequest(http.MethodGet, "/api/resultfile?key=results/testfile.json", nil)
	rec := httptest.NewRecorder()

	resultFileHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if strings.TrimSpace(rec.Body.String()) != content {
		t.Fatalf("unexpected content: %s", rec.Body.String())
	}
}
