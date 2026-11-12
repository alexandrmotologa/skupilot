package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/alexandrmotologa/skupilot/internal/api"
	"github.com/alexandrmotologa/skupilot/internal/domain"
	"github.com/alexandrmotologa/skupilot/internal/engine"
	"github.com/alexandrmotologa/skupilot/internal/store"
	"github.com/alexandrmotologa/skupilot/internal/vision"
)

func setupTestServer(t *testing.T) (*httptest.Server, *store.DB) {
	tmpDir := t.TempDir()
	db, err := store.Open(filepath.Join(tmpDir, "test.db"))
	if err != nil {
		t.Fatalf("failed opening test db: %v", err)
	}

	vlm := vision.NewHeuristicVLM()
	eng := engine.NewEngine(db, vlm)
	srv := api.NewServer(db, eng, nil)

	return httptest.NewServer(srv), db
}

func TestHealthEndpoint(t *testing.T) {
	ts, db := setupTestServer(t)
	defer ts.Close()
	defer db.Close()

	resp, err := http.Get(ts.URL + "/api/health")
	if err != nil {
		t.Fatalf("health request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", resp.StatusCode)
	}

	var data map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		t.Fatalf("failed decoding health json: %v", err)
	}
	if data["status"] != "ok" {
		t.Errorf("expected status ok, got %v", data["status"])
	}
}

func TestBatchAndProductsEndpoints(t *testing.T) {
	ts, db := setupTestServer(t)
	defer ts.Close()
	defer db.Close()

	// Seed test batch
	batch, _ := db.CreateBatch("Summer Catalog 2026", 1)
	p := &domain.Product{
		BatchID:               batch.ID,
		RawSKU:                "SUM-001",
		RawTitle:              "Linen Shirt",
		OptimizedTitle:        "Nordic Weave Men's 100% Linen Band Collar Summer Shirt",
		QualityScore:          95,
		PublicationStatus:     domain.StatusPending,
		ImageComplianceStatus: domain.CompliancePass,
	}
	_ = db.InsertProduct(p)

	// Test GET /api/batches
	resp, err := http.Get(ts.URL + "/api/batches")
	if err != nil {
		t.Fatalf("failed listing batches: %v", err)
	}
	defer resp.Body.Close()

	var batches []domain.CatalogBatch
	if err := json.NewDecoder(resp.Body).Decode(&batches); err != nil {
		t.Fatalf("failed decoding batches: %v", err)
	}
	if len(batches) != 1 || batches[0].ID != batch.ID {
		t.Errorf("expected 1 batch with id %s, got %+v", batch.ID, batches)
	}

	// Test GET /api/batches/{id}/products
	pResp, err := http.Get(ts.URL + "/api/batches/" + batch.ID + "/products")
	if err != nil {
		t.Fatalf("failed listing products: %v", err)
	}
	defer pResp.Body.Close()

	var pData struct {
		Total int              `json:"total"`
		Items []domain.Product `json:"items"`
	}
	if err := json.NewDecoder(pResp.Body).Decode(&pData); err != nil {
		t.Fatalf("failed decoding products: %v", err)
	}
	if pData.Total != 1 || len(pData.Items) != 1 {
		t.Errorf("expected 1 product, got total=%d, len=%d", pData.Total, len(pData.Items))
	}

	// Test PATCH /api/products/{id} to update publication status
	patchBody := map[string]interface{}{
		"publication_status": "APPROVED",
	}
	patchBytes, _ := json.Marshal(patchBody)
	req, _ := http.NewRequest(http.MethodPatch, ts.URL+"/api/products/"+p.ID, bytes.NewReader(patchBytes))
	req.Header.Set("Content-Type", "application/json")

	patchResp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("failed patching product: %v", err)
	}
	defer patchResp.Body.Close()

	if patchResp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 OK for patch, got %d", patchResp.StatusCode)
	}

	// Test GET /api/batches/{id}/export/amazon
	amzResp, err := http.Get(ts.URL + "/api/batches/" + batch.ID + "/export/amazon")
	if err != nil {
		t.Fatalf("failed exporting amazon tsv: %v", err)
	}
	defer amzResp.Body.Close()

	if amzResp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 OK for amazon export, got %d", amzResp.StatusCode)
	}
	if ct := amzResp.Header.Get("Content-Type"); ct != "text/tab-separated-values; charset=utf-8" {
		t.Errorf("expected TSV content type, got: %s", ct)
	}
}
