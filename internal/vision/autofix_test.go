package vision_test

import (
	"image/color"
	"testing"

	"github.com/alexandrmotologa/skupilot/internal/domain"
	"github.com/alexandrmotologa/skupilot/internal/vision"
)

func TestAutoWhitenAndPad(t *testing.T) {
	// Start with a non-compliant 600x400 image with light gray border RGB(220, 220, 220) and blue subject
	gray := color.RGBA{R: 220, G: 220, B: 220, A: 255}
	blue := color.RGBA{R: 30, G: 60, B: 180, A: 255}
	original := createSyntheticImage(600, 400, gray, blue)

	// Pre-check: original must fail compliance (width < 1000 and non-white background)
	preReport := vision.AuditImage(original)
	if preReport.Status == domain.CompliancePass {
		t.Fatalf("expected original to fail compliance")
	}

	// Apply Auto-Whiten and Pad with minimum 1000x1000 square requirement
	fixedImg, fixedReport := vision.AutoWhitenAndPad(original, 1000)

	bounds := fixedImg.Bounds()
	if bounds.Dx() < 1000 || bounds.Dy() < 1000 {
		t.Errorf("expected fixed image dimensions >= 1000x1000, got %dx%d", bounds.Dx(), bounds.Dy())
	}
	if bounds.Dx() != bounds.Dy() {
		t.Errorf("expected square 1:1 aspect ratio, got %dx%d", bounds.Dx(), bounds.Dy())
	}
	if !fixedReport.ZoomReady {
		t.Errorf("expected zoomReady=true on fixed image")
	}
	if fixedReport.WhiteBGScore < 0.985 {
		t.Errorf("expected whiteness score >= 0.985, got %.3f", fixedReport.WhiteBGScore)
	}
	if fixedReport.Status != domain.CompliancePass {
		t.Errorf("expected fixed image status PASS, got %s (diagnostics: %v)", fixedReport.Status, fixedReport.Diagnostics)
	}
}

func TestGenerateDefectHeatmap(t *testing.T) {
	gray := color.RGBA{R: 200, G: 200, B: 200, A: 255}
	blue := color.RGBA{R: 30, G: 60, B: 180, A: 255}
	original := createSyntheticImage(500, 500, gray, blue)

	heatmap := vision.GenerateDefectHeatmap(original)
	if heatmap == nil {
		t.Fatalf("expected non-nil defect heatmap")
	}

	// Corner pixel should have red overlay (R=239)
	r, _, _, a := heatmap.At(5, 5).RGBA()
	r8 := uint8(r >> 8)
	a8 := uint8(a >> 8)
	if r8 < 200 || a8 == 0 {
		t.Errorf("expected red highlight on non-white border pixel, got r=%d a=%d", r8, a8)
	}
}
