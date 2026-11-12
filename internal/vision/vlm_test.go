package vision_test

import (
	"context"
	"image/color"
	"testing"

	"github.com/alexandrmotologa/skupilot/internal/domain"
	"github.com/alexandrmotologa/skupilot/internal/vision"
)

func TestHeuristicVLMExtraction(t *testing.T) {
	vlm := vision.NewHeuristicVLM()

	// Create green image representing an emerald dress
	emerald := color.RGBA{R: 30, G: 160, B: 110, A: 255}
	img := createSyntheticImage(1000, 1000, color.RGBA{R: 255, G: 255, B: 255, A: 255}, emerald)

	raw := domain.RawFeedItem{
		SKU:         "SLK-DR-003",
		Title:       "Womens emerald green silk dress",
		Description: "Elegant evening slip dress with cowl neck",
		Category:    "Apparel > Women > Dresses",
		Brand:       "Aura Studio",
	}

	attrs, err := vlm.ExtractAttributes(context.Background(), img, raw)
	if err != nil {
		t.Fatalf("unexpected error in heuristic extraction: %v", err)
	}

	if attrs.Material != "Pure Mulberry Silk" {
		t.Errorf("expected silk material, got: %s", attrs.Material)
	}
	if attrs.Neckline != "Cowl Neckline" {
		t.Errorf("expected Cowl Neckline, got: %s", attrs.Neckline)
	}
	if attrs.ColorFamily != "Emerald Green" && attrs.ColorFamily != "Green" {
		t.Errorf("expected Emerald Green or Green color family, got: %s", attrs.ColorFamily)
	}
	if attrs.Sleeve != "Sleeveless" {
		t.Errorf("expected Sleeveless dress, got: %s", attrs.Sleeve)
	}
}
