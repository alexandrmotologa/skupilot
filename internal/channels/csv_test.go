package channels_test

import (
	"strings"
	"testing"

	"github.com/alexandrmotologa/skupilot/internal/channels"
	"github.com/alexandrmotologa/skupilot/internal/domain"
)

func TestGenerateEnrichedCSV(t *testing.T) {
	products := []domain.Product{
		{
			RawSKU:                "TEST-SKU-1",
			RawTitle:              "Linen Shirt",
			Brand:                 "Studio",
			ImageZoomReady:        true,
			ImageWhiteBGScore:     0.992,
			ImageComplianceStatus: domain.CompliancePass,
			ExtractedAttributes: domain.ExtractedAttributes{
				Material:    "100% Linen",
				ColorFamily: "White",
			},
			QualityScore: 95,
		},
	}

	csvBytes, err := channels.GenerateEnrichedCSV(products)
	if err != nil {
		t.Fatalf("failed generating enriched CSV: %v", err)
	}

	csvStr := string(csvBytes)
	if !strings.Contains(csvStr, "sku,raw_title,brand") {
		t.Errorf("missing header row in CSV")
	}
	if !strings.Contains(csvStr, "TEST-SKU-1,Linen Shirt,Studio") {
		t.Errorf("missing product row in CSV: %s", csvStr)
	}
	if !strings.Contains(csvStr, "100% Linen") {
		t.Errorf("missing extracted attribute in CSV: %s", csvStr)
	}
}
