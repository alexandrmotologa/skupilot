package vision_test

import (
	"image"
	"image/color"
	"testing"

	"github.com/alexandrmotologa/skupilot/internal/domain"
	"github.com/alexandrmotologa/skupilot/internal/vision"
)

// createSyntheticImage builds an in-memory image for testing compliance checks.
func createSyntheticImage(w, h int, borderRGB, centerRGB color.RGBA) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if x < 15 || x >= w-15 || y < 15 || y >= h-15 {
				img.Set(x, y, borderRGB)
			} else {
				img.Set(x, y, centerRGB)
			}
		}
	}
	return img
}

func TestAuditImage_Compliant(t *testing.T) {
	// 1200x1200px image with pure white border RGB(255, 255, 255) and navy blue subject RGB(20, 30, 80)
	white := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	navy := color.RGBA{R: 20, G: 30, B: 80, A: 255}
	img := createSyntheticImage(1200, 1200, white, navy)

	report := vision.AuditImage(img)

	if report.Width != 1200 || report.Height != 1200 {
		t.Errorf("expected 1200x1200, got %dx%d", report.Width, report.Height)
	}
	if !report.ZoomReady {
		t.Errorf("expected zoomReady=true for 1200x1200px")
	}
	if report.WhiteBGScore < 0.985 {
		t.Errorf("expected whiteness score >= 0.985, got %.3f", report.WhiteBGScore)
	}
	if report.Status != domain.CompliancePass {
		t.Errorf("expected status PASS, got %s", report.Status)
	}

	hex, family := vision.ExtractDominantColor(img)
	if family != "Navy Blue" && family != "Black" && family != "Blue" {
		t.Errorf("expected Navy Blue color family, got %s (hex: %s)", family, hex)
	}
}

func TestAuditImage_NonCompliant_SmallAndColoredBackground(t *testing.T) {
	// 600x600px image with gray border RGB(180, 180, 180) - fails zoom and pure white
	gray := color.RGBA{R: 180, G: 180, B: 180, A: 255}
	red := color.RGBA{R: 200, G: 30, B: 30, A: 255}
	img := createSyntheticImage(600, 600, gray, red)

	report := vision.AuditImage(img)

	if report.ZoomReady {
		t.Errorf("expected zoomReady=false for 600x600px")
	}
	if report.WhiteBGScore >= 0.900 {
		t.Errorf("expected whiteness score < 0.900, got %.3f", report.WhiteBGScore)
	}
	if report.Status != domain.ComplianceFail {
		t.Errorf("expected status FAIL, got %s", report.Status)
	}
}
