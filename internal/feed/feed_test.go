package feed_test

import (
	"os"
	"strings"
	"testing"

	"github.com/alexandrmotologa/skupilot/internal/feed"
)

func TestParseCSV(t *testing.T) {
	csvData := `sku,title,description,category,price,currency,brand,image_url
LIN-SH-001,Mens linen shirt white,Casual button up summer shirt made with linen fabric,Apparel > Men > Shirts,68.00,USD,Nordic Weave,https://example.com/shirt.jpg
OVR-HD-002,Heavyweight fleece hoodie black,Cozy relaxed pullover hooded sweatshirt,Apparel > Hoodies,84.00,USD,AeroThread,https://example.com/hoodie.jpg
`
	items, err := feed.ParseCSV(strings.NewReader(csvData))
	if err != nil {
		t.Fatalf("failed to parse CSV: %v", err)
	}

	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}

	first := items[0]
	if first.SKU != "LIN-SH-001" || first.Price != 68.0 || first.Brand != "Nordic Weave" {
		t.Errorf("unexpected first item: %+v", first)
	}
}

func TestParseSampleApparelCSV(t *testing.T) {
	file, err := os.Open("../../examples/sample_apparel_feed.csv")
	if err != nil {
		t.Fatalf("failed to open sample CSV: %v", err)
	}
	defer file.Close()

	items, err := feed.ParseCSV(file)
	if err != nil {
		t.Fatalf("failed to parse sample CSV: %v", err)
	}
	if len(items) != 8 {
		t.Errorf("expected 8 items, got %d", len(items))
	}
}

func TestParseSampleFootwearJSON(t *testing.T) {
	file, err := os.Open("../../examples/sample_footwear_feed.json")
	if err != nil {
		t.Fatalf("failed to open sample JSON: %v", err)
	}
	defer file.Close()

	items, err := feed.ParseJSON(file)
	if err != nil {
		t.Fatalf("failed to parse sample JSON: %v", err)
	}
	if len(items) != 3 {
		t.Errorf("expected 3 items, got %d", len(items))
	}
	if items[0].SKU != "CHL-BT-101" || items[0].Price != 180.0 {
		t.Errorf("unexpected first footwear item: %+v", items[0])
	}
}
