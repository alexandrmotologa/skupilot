package copywriter

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/alexandrmotologa/skupilot/internal/domain"
)

// Banned promotional words prohibited by Amazon A9 style guides.
var bannedPromotionalPhrases = []string{
	"best seller", "bestseller", "free shipping", "top rated", "cheap",
	"discount", "sale", "guarantee", "100% satisfaction", "perfect quality",
	"number one", "#1", "limited time", "special offer", "hot sale",
}

// GenerateA9Title synthesizes an Amazon A9 and marketplace optimized product title.
// Formula: [Brand] + [Target/Gender] + [Material] + [Product Type] + [Key Style/Feature] + [Color]
func GenerateA9Title(raw domain.RawFeedItem, attr domain.ExtractedAttributes) string {
	brand := strings.TrimSpace(raw.Brand)
	if brand == "" {
		brand = "Generic"
	}

	target := detectTargetGender(raw.Title, raw.Category)
	material := cleanMaterialName(attr.Material)
	productType := detectProductType(raw.Title, raw.Category)

	var features []string
	if attr.Collar != "" {
		features = append(features, attr.Collar)
	} else if attr.Neckline != "" {
		features = append(features, attr.Neckline)
	}

	if attr.Fit != "" && !strings.Contains(productType, attr.Fit) {
		features = append(features, attr.Fit)
	}

	colorDesc := attr.ColorFamily
	if colorDesc == "" {
		colorDesc = "Classic"
	}

	// Build raw parts
	parts := []string{brand}
	if target != "" {
		parts = append(parts, target)
	}
	if material != "" {
		parts = append(parts, material)
	}
	parts = append(parts, productType)

	if len(features) > 0 {
		parts = append(parts, strings.Join(features, ", "))
	}

	parts = append(parts, colorDesc)

	title := strings.Join(parts, " ")
	title = SanitizeTitle(title)

	// Enforce 150-200 characters bound
	return formatTitleLength(title, raw.Title, colorDesc)
}

// SanitizeTitle removes prohibited promotional jargon and excessive punctuation.
func SanitizeTitle(title string) string {
	lower := strings.ToLower(title)
	for _, banned := range bannedPromotionalPhrases {
		if strings.Contains(lower, banned) {
			re := strings.ReplaceAll(lower, banned, "")
			// Update title casing with cleaned version
			title = re
			lower = strings.ToLower(title)
		}
	}

	// Remove consecutive spaces and unwanted symbols
	words := strings.Fields(title)
	var cleaned []string
	for _, w := range words {
		w = strings.Trim(w, "!*#$")
		if w != "" {
			cleaned = append(cleaned, titleCase(w))
		}
	}

	return strings.Join(cleaned, " ")
}

func formatTitleLength(currentTitle, fallbackRaw, colorDesc string) string {
	// If title is under 120 characters, add clarifying context to reach optimal 150-180 length
	if len(currentTitle) < 130 {
		suffix := fmt.Sprintf(" - Versatile Everyday Essential in %s", colorDesc)
		if len(currentTitle)+len(suffix) <= 200 {
			currentTitle += suffix
		}
	}

	// Hard cap at 200 characters without cutting words in half
	if len(currentTitle) > 200 {
		runes := []rune(currentTitle)
		cut := string(runes[:197])
		lastSpace := strings.LastIndex(cut, " ")
		if lastSpace > 140 {
			cut = cut[:lastSpace]
		}
		return strings.TrimSpace(cut)
	}

	return currentTitle
}

func detectTargetGender(title, category string) string {
	combined := strings.ToLower(title + " " + category)
	if strings.Contains(combined, "women") || strings.Contains(combined, "ladies") {
		return "Women's"
	}
	if strings.Contains(combined, "men") || strings.Contains(combined, "gentlemen") {
		return "Men's"
	}
	if strings.Contains(combined, "kids") || strings.Contains(combined, "boys") || strings.Contains(combined, "girls") {
		return "Kids'"
	}
	return "Unisex"
}

func detectProductType(title, category string) string {
	combined := strings.ToLower(title + " " + category)
	types := []struct {
		kw    string
		label string
	}{
		{"chelsea boot", "Chelsea Ankle Boots"},
		{"sneaker", "Low-Top Casual Sneakers"},
		{"boot", "Heritage Leather Boots"},
		{"wallet", "Bifold Leather Wallet"},
		{"tote bag", "Everyday Canvas Tote Bag"},
		{"sunglasses", "Polarized Aviator Sunglasses"},
		{"slip dress", "Evening Slip Dress"},
		{"dress", "Silhouette Casual Dress"},
		{"hoodie", "Fleece Pullover Hoodie"},
		{"sweater", "Knit Crewneck Sweater"},
		{"blazer", "Tailored Suit Blazer"},
		{"denim jacket", "Trucker Denim Jacket"},
		{"jacket", "Classic Layering Jacket"},
		{"skirt", "Pleated Athletic Tennis Skirt"},
		{"flannel", "Brushed Flannel Button-Down Shirt"},
		{"shirt", "Casual Button-Down Summer Shirt"},
	}

	for _, t := range types {
		if strings.Contains(combined, t.kw) {
			return t.label
		}
	}

	return "Apparel Essential"
}

func cleanMaterialName(mat string) string {
	if mat == "" {
		return ""
	}
	// Strip marketing prefixes like "100% Breathable", keep essential fabric
	mat = strings.TrimPrefix(mat, "100% Breathable ")
	mat = strings.TrimPrefix(mat, "Pure ")
	mat = strings.TrimPrefix(mat, "Ultra-Soft ")
	return mat
}

func titleCase(s string) string {
	if len(s) == 0 {
		return s
	}
	// Small words in titles
	lower := strings.ToLower(s)
	if lower == "and" || lower == "in" || lower == "with" || lower == "for" || lower == "of" {
		return lower
	}
	r := []rune(lower)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}
