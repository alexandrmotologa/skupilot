package copywriter_test

import (
	"testing"

	"github.com/alexandrmotologa/skupilot/internal/copywriter"
	"github.com/alexandrmotologa/skupilot/internal/domain"
)

func TestGenerateAllStyles(t *testing.T) {
	raw := domain.RawFeedItem{
		SKU:      "LIN-001",
		Title:    "Mens Linen Shirt",
		Brand:    "Nordic Weave",
		Category: "Apparel > Shirts",
	}
	attr := domain.ExtractedAttributes{
		Material:    "100% Breathable Linen",
		Fit:         "Tailored Slim Fit",
		Collar:      "Mandarin Band Collar",
		ColorFamily: "Oatmeal Cream",
		ColorHex:    "#EAE6DF",
		Closure:     "Shell Buttons",
		Care:        "Machine wash cold",
	}

	styles := copywriter.GenerateAllStyles(raw, attr)

	// Check Amazon A9
	amz, ok := styles[string(copywriter.StyleAmazonA9)]
	if !ok || len(amz.Bullets) != 5 {
		t.Errorf("missing or invalid Amazon style: %+v", amz)
	}

	// Check Shopify Editorial
	shp, ok := styles[string(copywriter.StyleShopifyEditorial)]
	if !ok || shp.Title == "" || len(shp.Bullets) == 0 {
		t.Errorf("missing or invalid Shopify Editorial style: %+v", shp)
	}

	// Check Google Shopping
	goog, ok := styles[string(copywriter.StyleGoogleShopping)]
	if !ok || len(goog.Title) > 150 {
		t.Errorf("google title exceeds 150 chars: %s", goog.Title)
	}
}
