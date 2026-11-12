package engine_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/alexandrmotologa/skupilot/internal/domain"
	"github.com/alexandrmotologa/skupilot/internal/engine"
	"github.com/alexandrmotologa/skupilot/internal/store"
	"github.com/alexandrmotologa/skupilot/internal/vision"
)

func TestEngineEnrichmentPipeline(t *testing.T) {
	tmpDir := t.TempDir()
	db, err := store.Open(filepath.Join(tmpDir, "test.db"))
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	batch, err := db.CreateBatch("Spring 2026 Test Batch", 2)
	if err != nil {
		t.Fatalf("failed to create batch: %v", err)
	}

	vlm := vision.NewHeuristicVLM()
	eng := engine.NewEngine(db, vlm)

	items := []domain.RawFeedItem{
		{
			SKU:         "TEST-SH-100",
			Title:       "Men Casual Linen Band Collar Shirt",
			Description: "Breathable lightweight summer button-up shirt in white",
			Category:    "Apparel > Men > Shirts",
			Price:       75.0,
			Brand:       "Nordic Weave",
		},
		{
			SKU:         "TEST-JK-200",
			Title:       "Heavyweight Flannel Trucker Jacket",
			Description: "Brushed flannel insulated jacket with metal buttons",
			Category:    "Apparel > Outerwear > Jackets",
			Price:       120.0,
			Brand:       "Forge & Co",
		},
	}

	progressCount := 0
	err = eng.ProcessFeed(context.Background(), batch.ID, items, func(current, total int, sku string, p *domain.Product) {
		progressCount++
	})
	if err != nil {
		t.Fatalf("process feed failed: %v", err)
	}

	if progressCount != 2 {
		t.Errorf("expected 2 progress callbacks, got %d", progressCount)
	}

	// Verify database persistence
	products, total, err := db.ListProducts(store.ProductFilter{BatchID: batch.ID})
	if err != nil {
		t.Fatalf("failed to query products: %v", err)
	}
	if total != 2 {
		t.Fatalf("expected 2 products saved, got %d", total)
	}

	for _, p := range products {
		if p.OptimizedTitle == "" {
			t.Errorf("expected non-empty optimized title for %s", p.RawSKU)
		}
		if len(p.BulletPoints) != 5 {
			t.Errorf("expected 5 bullet points for %s, got %d", p.RawSKU, len(p.BulletPoints))
		}
		if p.QualityScore <= 0 {
			t.Errorf("expected positive quality score, got %d", p.QualityScore)
		}
	}
}
