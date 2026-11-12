package store_test

import (
	"path/filepath"
	"testing"

	"github.com/alexandrmotologa/skupilot/internal/domain"
	"github.com/alexandrmotologa/skupilot/internal/store"
)

func TestStoreOperations(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_catalog.db")

	db, err := store.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	// 1. Create Batch
	batch, err := db.CreateBatch("Spring 2026 Test Feed", 10)
	if err != nil {
		t.Fatalf("failed to create batch: %v", err)
	}
	if batch.ID == "" || batch.TotalSKUs != 10 {
		t.Fatalf("unexpected batch fields: %+v", batch)
	}

	// 2. Insert Products
	p1 := &domain.Product{
		BatchID:  batch.ID,
		RawSKU:   "TEST-SKU-001",
		RawTitle: "Pure Linen White Summer Shirt",
		RawPrice: 65.0,
		Currency: "USD",
		Brand:    "Nordic Studio",
		ImageURL: "https://example.com/shirt.jpg",
		ExtractedAttributes: domain.ExtractedAttributes{
			Material:    "100% Breathable Linen",
			Collar:      "Mandarin Band Collar",
			Fit:         "Slim Fit",
			ColorHex:    "#FFFFFF",
			ColorFamily: "White",
		},
		OptimizedTitle: "Nordic Studio Men's 100% Linen Band Collar Summer Shirt - White",
		BulletPoints: []string{
			"100% Breathable Linen: Keeps you cool.",
			"Mandarin Collar: Clean minimalist aesthetic.",
		},
		SEODescriptionHTML:    "<p>Summer linen shirt</p>",
		GoogleTaxonomyID:      212,
		GoogleTaxonomyPath:    "Apparel > Shirts",
		QualityScore:          95,
		ImageComplianceStatus: domain.CompliancePass,
		PublicationStatus:     domain.StatusPending,
	}

	if err := db.InsertProduct(p1); err != nil {
		t.Fatalf("failed to insert product: %v", err)
	}

	// 3. Retrieve Product
	fetched, err := db.GetProduct(p1.ID)
	if err != nil {
		t.Fatalf("failed to get product: %v", err)
	}
	if fetched == nil {
		t.Fatalf("expected product to be found")
	}
	if fetched.RawSKU != "TEST-SKU-001" || fetched.ExtractedAttributes.Material != "100% Breathable Linen" {
		t.Errorf("mismatch in fetched product attributes: %+v", fetched)
	}

	// 4. Update Product
	fetched.PublicationStatus = domain.StatusApproved
	fetched.QualityScore = 98
	if err := db.UpdateProduct(fetched); err != nil {
		t.Fatalf("failed to update product: %v", err)
	}

	updated, _ := db.GetProduct(p1.ID)
	if updated.PublicationStatus != domain.StatusApproved || updated.QualityScore != 98 {
		t.Errorf("update failed to reflect: %+v", updated)
	}

	// 5. Query Products with Filter
	items, total, err := db.ListProducts(store.ProductFilter{
		BatchID:           batch.ID,
		PublicationStatus: string(domain.StatusApproved),
	})
	if err != nil {
		t.Fatalf("failed to list products: %v", err)
	}
	if total != 1 || len(items) != 1 {
		t.Fatalf("expected 1 approved item, got total=%d len=%d", total, len(items))
	}
}
