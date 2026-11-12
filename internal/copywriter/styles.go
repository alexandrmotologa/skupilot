package copywriter

import (
	"fmt"
	"strings"

	"github.com/alexandrmotologa/skupilot/internal/domain"
)

// CopyStyle represents the target merchandising voice and channel format.
type CopyStyle string

const (
	StyleAmazonA9          CopyStyle = "AMAZON_A9"
	StyleShopifyEditorial  CopyStyle = "SHOPIFY_EDITORIAL"
	StyleGoogleShopping    CopyStyle = "GOOGLE_SHOPPING"
)

// StyleVariant holds a complete set of title, bullets, and description generated for a style.
type StyleVariant struct {
	Style           CopyStyle `json:"style"`
	Title           string    `json:"title"`
	Bullets         []string  `json:"bullets"`
	DescriptionHTML string    `json:"description_html"`
}

// GenerateAllStyles produces copy variants across Amazon, Shopify D2C, and Google Shopping.
func GenerateAllStyles(raw domain.RawFeedItem, attr domain.ExtractedAttributes) map[string]StyleVariant {
	m := make(map[string]StyleVariant)

	// 1. Amazon A9 Style
	a9Title := GenerateA9Title(raw, attr)
	a9Bullets := GenerateBullets(raw, attr)
	a9Desc := GenerateSEODescriptionHTML(raw, attr, a9Bullets)
	m[string(StyleAmazonA9)] = StyleVariant{
		Style:           StyleAmazonA9,
		Title:           a9Title,
		Bullets:         a9Bullets,
		DescriptionHTML: a9Desc,
	}

	// 2. Shopify Editorial / D2C Luxury Style
	edTitle, edBullets, edDesc := generateShopifyEditorial(raw, attr)
	m[string(StyleShopifyEditorial)] = StyleVariant{
		Style:           StyleShopifyEditorial,
		Title:           edTitle,
		Bullets:         edBullets,
		DescriptionHTML: edDesc,
	}

	// 3. Google Shopping Spec-First Style
	googTitle, googBullets, googDesc := generateGoogleShopping(raw, attr)
	m[string(StyleGoogleShopping)] = StyleVariant{
		Style:           StyleGoogleShopping,
		Title:           googTitle,
		Bullets:         googBullets,
		DescriptionHTML: googDesc,
	}

	return m
}

func generateShopifyEditorial(raw domain.RawFeedItem, attr domain.ExtractedAttributes) (string, []string, string) {
	brand := raw.Brand
	if brand == "" {
		brand = "The Studio"
	}

	material := cleanMaterialName(attr.Material)
	if material == "" {
		material = "Organic Cotton"
	}

	colorDesc := attr.ColorFamily
	if colorDesc == "" {
		colorDesc = "Classic"
	}

	pType := detectProductType(raw.Title, raw.Category)

	// Clean, evocative title for D2C storefront
	title := fmt.Sprintf("%s The %s in %s", brand, pType, colorDesc)

	bullets := []string{
		fmt.Sprintf("Naturally refined: Woven from %s with a soft, lived-in hand feel.", strings.ToLower(material)),
		fmt.Sprintf("Effortless drape: Cut in a %s designed to move with ease.", strings.ToLower(attr.Fit)),
		fmt.Sprintf("Subtle detailing: Finished with %s and clean seamlines.", strings.ToLower(attr.Closure)),
		fmt.Sprintf("Considered care: %s to preserve natural fibers.", strings.ToLower(attr.Care)),
	}

	var sb strings.Builder
	sb.WriteString("<div class=\"editorial-story\">\n")
	sb.WriteString(fmt.Sprintf("  <p class=\"lead\">An elevated essential designed for timeless everyday wear. The %s pairs %s with an effortless %s silhouette, offering unparalleled comfort and quiet elegance.</p>\n",
		pType, strings.ToLower(material), strings.ToLower(attr.Fit)))
	sb.WriteString("  <div class=\"styling-note\">\n")
	sb.WriteString(fmt.Sprintf("    <p><strong>Styling Note:</strong> Wear it casually unbuttoned over a crisp base layer, or styled with tailored trousers for an understated evening look.</p>\n"))
	sb.WriteString("  </div>\n")
	sb.WriteString("</div>")

	return title, bullets, sb.String()
}

func generateGoogleShopping(raw domain.RawFeedItem, attr domain.ExtractedAttributes) (string, []string, string) {
	brand := raw.Brand
	if brand == "" {
		brand = "Generic"
	}

	target := detectTargetGender(raw.Title, raw.Category)
	pType := detectProductType(raw.Title, raw.Category)
	material := cleanMaterialName(attr.Material)
	colorDesc := attr.ColorFamily
	if colorDesc == "" {
		colorDesc = "Multi"
	}

	// Google Shopping recommends: [Brand] + [Gender] + [Type] + [Material] + [Color] - under 150 chars
	title := fmt.Sprintf("%s %s %s, %s, %s", brand, target, pType, material, colorDesc)
	if len(title) > 145 {
		title = title[:145]
	}

	bullets := []string{
		fmt.Sprintf("Material: %s", attr.Material),
		fmt.Sprintf("Color: %s (%s)", attr.ColorFamily, attr.ColorHex),
		fmt.Sprintf("Fit: %s", attr.Fit),
		fmt.Sprintf("Closure: %s", attr.Closure),
		fmt.Sprintf("Care: %s", attr.Care),
	}

	desc := fmt.Sprintf("%s. Material: %s. Color: %s. Fit: %s. Suitable for %s.",
		raw.Title, attr.Material, colorDesc, attr.Fit, target)

	return title, bullets, desc
}
