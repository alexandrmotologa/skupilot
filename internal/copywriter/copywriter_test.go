package copywriter_test

import (
	"strings"
	"testing"

	"github.com/alexandrmotologa/skupilot/internal/copywriter"
	"github.com/alexandrmotologa/skupilot/internal/domain"
)

func TestA9TitleGeneration(t *testing.T) {
	raw := domain.RawFeedItem{
		SKU:      "LIN-SH-001",
		Title:    "Mens linen shirt white",
		Category: "Apparel > Men > Shirts",
		Brand:    "Nordic Weave",
	}
	attr := domain.ExtractedAttributes{
		Material:    "100% Breathable Linen",
		Collar:      "Mandarin Band Collar",
		Fit:         "Tailored Slim Fit",
		ColorFamily: "Crisp White",
	}

	title := copywriter.GenerateA9Title(raw, attr)

	if !strings.HasPrefix(title, "Nordic Weave") {
		t.Errorf("expected title to start with brand, got: %s", title)
	}
	if len(title) < 100 || len(title) > 200 {
		t.Errorf("title length %d not within optimal bounds (100-200): %s", len(title), title)
	}
	if strings.Contains(strings.ToLower(title), "best seller") {
		t.Errorf("title contains banned words: %s", title)
	}
}

func TestBulletsGeneration(t *testing.T) {
	raw := domain.RawFeedItem{
		SKU:   "OVR-HD-002",
		Title: "Heavyweight fleece hoodie black",
		Brand: "AeroThread",
	}
	attr := domain.ExtractedAttributes{
		Material:    "Heavyweight Cotton Fleece",
		Fit:         "Relaxed Oversized Fit",
		Closure:     "Pullover Overhead",
		Care:        "Machine wash cold gentle",
		ColorFamily: "Onyx Black",
	}

	bullets := copywriter.GenerateBullets(raw, attr)
	if len(bullets) != 5 {
		t.Fatalf("expected exactly 5 bullets, got %d", len(bullets))
	}
	for i, b := range bullets {
		if !strings.Contains(b, ":") {
			t.Errorf("bullet %d missing headline colon structure: %s", i, b)
		}
	}
}

func TestQualityScorecard(t *testing.T) {
	p := &domain.Product{
		Brand:             "Nordic Weave",
		ImageWhiteBGScore: 0.995,
		ImageZoomReady:    true,
		ExtractedAttributes: domain.ExtractedAttributes{
			Material:    "100% Linen",
			Fit:         "Slim",
			Collar:      "Mandarin",
			ColorFamily: "White",
			Closure:     "Buttons",
		},
		OptimizedTitle: "Nordic Weave Men's 100% Linen Band Collar Summer Shirt - Tailored Casual Slim Fit in Pure White",
		BulletPoints: []string{
			"PREMIUM LINEN: Breathable woven natural fibers.",
			"REFINED FIT: Flattering silhouette and drape.",
			"PRECISION HARDWARE: Cross-stitched natural buttons.",
			"EASY CARE: Machine washable and durable.",
			"VERSATILE STYLING: Everyday work and leisure.",
		},
	}

	scorecard := copywriter.CalculateQualityScorecard(p)
	if scorecard.TotalScore < 90 {
		t.Errorf("expected high quality score for fully compliant listing, got %d (rec: %v)", scorecard.TotalScore, scorecard.Recommendations)
	}
}
