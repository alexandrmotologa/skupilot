package copywriter

import (
	"fmt"
	"strings"

	"github.com/alexandrmotologa/skupilot/internal/domain"
)

// GenerateBullets creates 5 benefit-driven feature bullets conforming to Amazon Seller Central standards.
func GenerateBullets(raw domain.RawFeedItem, attr domain.ExtractedAttributes) []string {
	material := attr.Material
	if material == "" {
		material = "Premium high-grade fabric"
	}

	fit := attr.Fit
	if fit == "" {
		fit = "Contemporary tailored cut"
	}

	closure := attr.Closure
	if closure == "" {
		closure = "Precision stitched hardware"
	}

	care := attr.Care
	if care == "" {
		care = "Machine wash cold on gentle cycle, tumble dry low"
	}

	colorDesc := attr.ColorFamily
	if colorDesc == "" {
		colorDesc = "timeless neutral"
	}

	bullets := []string{
		fmt.Sprintf("PREMIUM %s: Engineered with authentic %s to provide natural breathability, long-lasting comfort, and superior hand-feel throughout the day.",
			strings.ToUpper(cleanMaterialName(material)), material),

		fmt.Sprintf("REFINED %s: Designed with a %s that offers a flattering drape without restricting natural range of movement.",
			strings.ToUpper(strings.ReplaceAll(fit, " ", "_")), fit),

		fmt.Sprintf("PRECISION CRAFTSMANSHIP: Features %s and reinforced high-stress seams to ensure shape retention through extended daily wear.",
			closure),

		fmt.Sprintf("EASY CARE & MAINTENANCE: Simple garment care with %s to preserve fabric integrity, texture, and vibrant %s color.",
			care, colorDesc),

		fmt.Sprintf("VERSATILE STYLING: An effortless wardrobe staple that transitions seamlessly from refined daytime occasions to relaxed evening settings.",
		),
	}

	return bullets
}
