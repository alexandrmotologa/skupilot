package vision

import (
	"image"
	"image/color"
	"image/draw"
	"math"

	"github.com/alexandrmotologa/skupilot/internal/domain"
)

// AutoWhitenAndPad transforms an image into an Amazon-compliant pure white 1:1 square image.
// It centers the product, turns the border into pure RGB(255, 255, 255), and ensures minimum 1000x1000px dimensions.
func AutoWhitenAndPad(img image.Image, minTargetDim int) (image.Image, domain.ComplianceReport) {
	if minTargetDim < 1000 {
		minTargetDim = 1000
	}

	srcBounds := img.Bounds()
	srcW := srcBounds.Dx()
	srcH := srcBounds.Dy()

	// Determine square canvas size (at least minTargetDim or max of original dimensions)
	canvasDim := srcW
	if srcH > canvasDim {
		canvasDim = srcH
	}
	if canvasDim < minTargetDim {
		canvasDim = minTargetDim
	}

	// Create square RGBA image filled with pure white RGB(255, 255, 255)
	dst := image.NewRGBA(image.Rect(0, 0, canvasDim, canvasDim))
	whiteColor := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	draw.Draw(dst, dst.Bounds(), &image.Uniform{whiteColor}, image.Point{}, draw.Src)

	// Calculate centering coordinates with 5% safety margin padding
	usableDim := int(float64(canvasDim) * 0.90)
	scale := 1.0
	if srcW > usableDim || srcH > usableDim {
		scaleW := float64(usableDim) / float64(srcW)
		scaleH := float64(usableDim) / float64(srcH)
		scale = math.Min(scaleW, scaleH)
	}

	scaledW := int(float64(srcW) * scale)
	scaledH := int(float64(srcH) * scale)

	offsetX := (canvasDim - scaledW) / 2
	offsetY := (canvasDim - scaledH) / 2

	// Copy pixels from source with background whitening filter
	// Sample corner colors from source to identify background threshold
	cornerThreshold := estimateSourceBackgroundThreshold(img)

	for y := 0; y < scaledH; y++ {
		srcY := srcBounds.Min.Y + int(float64(y)/scale)
		dstY := offsetY + y
		for x := 0; x < scaledW; x++ {
			srcX := srcBounds.Min.X + int(float64(x)/scale)
			dstX := offsetX + x

			r, g, b, a := img.At(srcX, srcY).RGBA()
			r8 := uint8(r >> 8)
			g8 := uint8(g >> 8)
			b8 := uint8(b >> 8)
			a8 := uint8(a >> 8)

			// If pixel is near the detected background or close to white, force to pure white RGB(255, 255, 255)
			if isBackgroundPixel(r8, g8, b8, cornerThreshold) {
				dst.Set(dstX, dstY, whiteColor)
			} else {
				dst.Set(dstX, dstY, color.RGBA{R: r8, G: g8, B: b8, A: a8})
			}
		}
	}

	// Re-audit transformed image
	report := AuditImage(dst)
	return dst, report
}

// GenerateDefectHeatmap creates an overlay showing non-compliant pixels highlighted in red.
func GenerateDefectHeatmap(img image.Image) image.Image {
	bounds := img.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()

	heatmap := image.NewRGBA(image.Rect(0, 0, w, h))

	borderThickness := 15
	if borderThickness > w/8 {
		borderThickness = w / 8
	}
	if borderThickness > h/8 {
		borderThickness = h / 8
	}

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			r8 := uint8(r >> 8)
			g8 := uint8(g >> 8)
			b8 := uint8(b >> 8)

			isBorder := (x < bounds.Min.X+borderThickness) ||
				(x >= bounds.Max.X-borderThickness) ||
				(y < bounds.Min.Y+borderThickness) ||
				(y >= bounds.Max.Y-borderThickness)

			if isBorder {
				// If border pixel is not pure white RGB(>=250), highlight in semi-transparent red
				if r8 < 250 || g8 < 250 || b8 < 250 {
					heatmap.Set(x, y, color.RGBA{R: 239, G: 68, B: 68, A: 190}) // red highlight
				} else {
					// Compliant pure white pixel: subtle emerald tint
					heatmap.Set(x, y, color.RGBA{R: 16, G: 185, B: 129, A: 40})
				}
			} else {
				// Interior product area: keep transparent or original subdued
				heatmap.Set(x, y, color.RGBA{R: r8, G: g8, B: b8, A: 255})
			}
		}
	}

	return heatmap
}

func estimateSourceBackgroundThreshold(img image.Image) uint8 {
	bounds := img.Bounds()
	corners := [][2]int{
		{bounds.Min.X, bounds.Min.Y},
		{bounds.Max.X - 1, bounds.Min.Y},
		{bounds.Min.X, bounds.Max.Y - 1},
		{bounds.Max.X - 1, bounds.Max.Y - 1},
	}

	var sumLum uint64
	for _, c := range corners {
		r, g, b, _ := img.At(c[0], c[1]).RGBA()
		r8 := uint8(r >> 8)
		g8 := uint8(g >> 8)
		b8 := uint8(b >> 8)
		lum := uint64(0.299*float64(r8) + 0.587*float64(g8) + 0.114*float64(b8))
		sumLum += lum
	}

	avgLum := uint8(sumLum / 4)
	if avgLum > 210 {
		return avgLum - 15
	}
	return 235
}

func isBackgroundPixel(r, g, b, threshold uint8) bool {
	// If pixel is brighter than threshold and low color saturation, consider background
	if r >= threshold && g >= threshold && b >= threshold {
		rf := float64(r)
		gf := float64(g)
		bf := float64(b)
		diff := math.Max(rf, math.Max(gf, bf)) - math.Min(rf, math.Min(gf, bf))
		return diff < 25
	}
	return false
}
