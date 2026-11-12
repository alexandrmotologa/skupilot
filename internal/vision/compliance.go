package vision

import (
	"context"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"net/http"
	"time"

	"github.com/alexandrmotologa/skupilot/internal/domain"
)

// AuditImage performs comprehensive mathematical compliance checks on an image.
func AuditImage(img image.Image) domain.ComplianceReport {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	aspectRatio := 1.0
	if height > 0 {
		aspectRatio = math.Round((float64(width)/float64(height))*100) / 100
	}

	zoomReady := (width >= 1000) && (height >= 1000)

	// Background Whiteness Metric
	// Sample outer perimeter (10px border band)
	borderThickness := 10
	if borderThickness > width/10 {
		borderThickness = width / 10
	}
	if borderThickness > height/10 {
		borderThickness = height / 10
	}
	if borderThickness < 1 {
		borderThickness = 1
	}

	whitePixels := 0
	totalPerimeterPixels := 0
	borderContrastVariance := 0.0

	// We sample along top, bottom, left, and right borders
	var borderLuminances []float64

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			isBorder := (x < bounds.Min.X+borderThickness) ||
				(x >= bounds.Max.X-borderThickness) ||
				(y < bounds.Min.Y+borderThickness) ||
				(y >= bounds.Max.Y-borderThickness)

			if isBorder {
				totalPerimeterPixels++
				r, g, b, _ := img.At(x, y).RGBA()
				// Convert from 16-bit to 8-bit
				r8 := uint8(r >> 8)
				g8 := uint8(g >> 8)
				b8 := uint8(b >> 8)

				lum := 0.299*float64(r8) + 0.587*float64(g8) + 0.114*float64(b8)
				borderLuminances = append(borderLuminances, lum)

				// Amazon pure white requirement (RGB 250-255 tolerance)
				if r8 >= 250 && g8 >= 250 && b8 >= 250 {
					whitePixels++
				}
			}
		}
	}

	whiteScore := 0.0
	if totalPerimeterPixels > 0 {
		whiteScore = float64(whitePixels) / float64(totalPerimeterPixels)
		whiteScore = math.Round(whiteScore*1000) / 1000
	}

	// Calculate border luminance variance to detect watermarks or text in corners
	if len(borderLuminances) > 0 {
		meanLum := 0.0
		for _, l := range borderLuminances {
			meanLum += l
		}
		meanLum /= float64(len(borderLuminances))

		sumSqDiff := 0.0
		for _, l := range borderLuminances {
			diff := l - meanLum
			sumSqDiff += diff * diff
		}
		borderContrastVariance = sumSqDiff / float64(len(borderLuminances))
	}

	// Watermark or promotional overlay detection heuristic:
	// If background is not pure white OR high edge contrast variance exists in the outer perimeter
	watermarkDetected := false
	if borderContrastVariance > 900.0 && whiteScore < 0.98 {
		watermarkDetected = true
	}

	// Compile diagnostics
	var diagnostics []string
	status := domain.CompliancePass

	if !zoomReady {
		diagnostics = append(diagnostics, fmt.Sprintf("Image resolution (%dx%d) below Amazon 1000x1000px zoom requirement", width, height))
		status = domain.ComplianceWarn
	}

	if whiteScore < 0.900 {
		diagnostics = append(diagnostics, fmt.Sprintf("Main image background is not pure white (score: %.1f%%, Amazon requires >= 98.5%%)", whiteScore*100))
		status = domain.ComplianceFail
	} else if whiteScore < 0.985 {
		diagnostics = append(diagnostics, fmt.Sprintf("Background whiteness score (%.1f%%) slightly below Amazon 98.5%% pure white standard", whiteScore*100))
		if status != domain.ComplianceFail {
			status = domain.ComplianceWarn
		}
	}

	if aspectRatio < 0.8 || aspectRatio > 1.25 {
		diagnostics = append(diagnostics, fmt.Sprintf("Aspect ratio (%.2f) outside standard 1:1 square bounds (0.80 - 1.25)", aspectRatio))
		if status != domain.ComplianceFail {
			status = domain.ComplianceWarn
		}
	}

	if watermarkDetected {
		diagnostics = append(diagnostics, "Potential text overlay or watermark detected in outer border area")
		status = domain.ComplianceFail
	}

	if len(diagnostics) == 0 {
		diagnostics = append(diagnostics, "Fully compliant with Amazon and Google Shopping image specifications")
	}

	return domain.ComplianceReport{
		Width:             width,
		Height:            height,
		AspectRatio:       aspectRatio,
		ZoomReady:         zoomReady,
		WhiteBGScore:      whiteScore,
		WatermarkDetected: watermarkDetected,
		Status:            status,
		Diagnostics:       diagnostics,
	}
}

// ExtractDominantColor samples the central subject area of an image to find the primary color and family.
func ExtractDominantColor(img image.Image) (string, string) {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Focus on central 60% of the image to sample product pixels
	minX := bounds.Min.X + int(float64(width)*0.2)
	maxX := bounds.Max.X - int(float64(width)*0.2)
	minY := bounds.Min.Y + int(float64(height)*0.2)
	maxY := bounds.Max.Y - int(float64(height)*0.2)

	var totalR, totalG, totalB uint64
	sampleCount := 0

	for y := minY; y < maxY; y += 4 {
		for x := minX; x < maxX; x += 4 {
			r, g, b, _ := img.At(x, y).RGBA()
			r8 := uint8(r >> 8)
			g8 := uint8(g >> 8)
			b8 := uint8(b >> 8)

			// Skip near-white background pixels (> 245)
			if r8 > 245 && g8 > 245 && b8 > 245 {
				continue
			}

			totalR += uint64(r8)
			totalG += uint64(g8)
			totalB += uint64(b8)
			sampleCount++
		}
	}

	if sampleCount == 0 {
		return "#FFFFFF", "White"
	}

	avgR := uint8(totalR / uint64(sampleCount))
	avgG := uint8(totalG / uint64(sampleCount))
	avgB := uint8(totalB / uint64(sampleCount))

	hexCode := fmt.Sprintf("#%02X%02X%02X", avgR, avgG, avgB)
	family := ClassifyColor(avgR, avgG, avgB)

	return hexCode, family
}

// ClassifyColor matches an RGB value to common e-commerce color families.
func ClassifyColor(r, g, b uint8) string {
	rf := float64(r)
	gf := float64(g)
	bf := float64(b)

	// Brightness and saturation
	max := math.Max(rf, math.Max(gf, bf))
	min := math.Min(rf, math.Min(gf, bf))
	delta := max - min

	// Achromatic (grayscale)
	if delta < 18 {
		if max > 220 {
			return "White"
		} else if max > 160 {
			return "Light Gray"
		} else if max > 90 {
			return "Charcoal"
		} else {
			return "Black"
		}
	}

	// Blue shades
	if bf > rf && bf > gf {
		if max < 80 {
			return "Navy Blue"
		}
		if gf > 140 {
			return "Sky Blue"
		}
		return "Blue"
	}

	// Green shades
	if gf > rf && gf > bf {
		if max < 100 {
			return "Olive Green"
		}
		if bf > 100 {
			return "Emerald Green"
		}
		return "Green"
	}

	// Red and Warm shades
	if rf > gf && rf > bf {
		if gf < 60 && bf < 60 {
			if rf < 120 {
				return "Burgundy"
			}
			return "Red"
		}
		if gf > 120 && bf < 80 {
			return "Orange"
		}
		if gf > 180 && bf < 120 {
			return "Yellow"
		}
		if gf > 140 && bf > 120 {
			return "Tan Brown"
		}
		if bf > 140 {
			return "Pink"
		}
		return "Brown"
	}

	return "Multi-Color"
}

// FetchAndAuditImage downloads an image from a URL and runs the compliance audit.
func FetchAndAuditImage(ctx context.Context, client *http.Client, imageURL string) (*domain.ComplianceReport, image.Image, error) {
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, imageURL, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid image request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		// Return graceful failure report
		rep := &domain.ComplianceReport{
			Status: domain.ComplianceWarn,
			Diagnostics: []string{
				fmt.Sprintf("Image could not be retrieved from URL (%v)", err),
			},
		}
		return rep, nil, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		rep := &domain.ComplianceReport{
			Status: domain.ComplianceWarn,
			Diagnostics: []string{
				fmt.Sprintf("Image URL returned HTTP status %d", resp.StatusCode),
			},
		}
		return rep, nil, nil
	}

	img, _, err := image.Decode(resp.Body)
	if err != nil {
		rep := &domain.ComplianceReport{
			Status: domain.ComplianceWarn,
			Diagnostics: []string{
				fmt.Sprintf("Failed to decode image bytes: %v", err),
			},
		}
		return rep, nil, nil
	}

	rep := AuditImage(img)
	return &rep, img, nil
}
