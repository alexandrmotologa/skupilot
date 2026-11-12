package channels_test

import (
	"strings"
	"testing"

	"github.com/alexandrmotologa/skupilot/internal/channels"
	"github.com/alexandrmotologa/skupilot/internal/domain"
)

func TestShopifyInputBuilder(t *testing.T) {
	p := domain.Product{
		RawSKU:         "LIN-001",
		RawTitle:       "Men Linen Shirt",
		OptimizedTitle: "Nordic Weave Men's 100% Linen Band Collar Casual Summer Shirt - Crisp White",
		Brand:          "Nordic Weave",
		RawCategory:    "Apparel > Shirts",
		QualityScore:   96,
		ExtractedAttributes: domain.ExtractedAttributes{
			Material:    "100% Breathable Linen",
			Collar:      "Mandarin Band Collar",
			Fit:         "Tailored Slim Fit",
			ColorHex:    "#FFFFFF",
			ColorFamily: "Crisp White",
		},
	}

	input := channels.BuildShopifyProductInput(p, "DRAFT")
	if input.Title != p.OptimizedTitle {
		t.Errorf("expected title to match optimized title, got: %s", input.Title)
	}
	if input.Status != "DRAFT" {
		t.Errorf("expected status DRAFT, got: %s", input.Status)
	}
	if len(input.Metafields) < 4 {
		t.Errorf("expected at least 4 metafields, got %d", len(input.Metafields))
	}
}

func TestAmazonFlatFileGenerator(t *testing.T) {
	products := []domain.Product{
		{
			RawSKU:         "LIN-001",
			OptimizedTitle: "Nordic Weave Linen Shirt",
			Brand:          "Nordic Weave",
			RawCategory:    "Shirts",
			RawPrice:       65.0,
			ImageURL:       "https://example.com/shirt.jpg",
			BulletPoints: []string{
				"PREMIUM LINEN: 100% Pure linen.",
				"REFINED FIT: Flattering silhouette.",
			},
			ExtractedAttributes: domain.ExtractedAttributes{
				ColorFamily: "White",
				Material:    "Linen",
			},
		},
	}

	tsvBytes := channels.GenerateAmazonFlatFile(products)
	tsv := string(tsvBytes)

	if !strings.Contains(tsv, "item_sku\texternal_product_id") {
		t.Errorf("missing TSV header row")
	}
	if !strings.Contains(tsv, "LIN-001\tLIN-001\tSKU\tNordic Weave Linen Shirt") {
		t.Errorf("missing product row in TSV: %s", tsv)
	}
}

func TestGoogleMerchantXMLGenerator(t *testing.T) {
	products := []domain.Product{
		{
			RawSKU:             "LIN-001",
			OptimizedTitle:     "Nordic Weave Linen Shirt",
			Brand:              "Nordic Weave",
			RawPrice:           65.0,
			Currency:           "USD",
			ImageURL:           "https://example.com/shirt.jpg",
			GoogleTaxonomyID:   212,
			GoogleTaxonomyPath: "Apparel & Accessories > Clothing > Shirts & Tops",
		},
	}

	xmlBytes, err := channels.GenerateGoogleShoppingXML(products, "Test Store", "https://store.example.com")
	if err != nil {
		t.Fatalf("failed to generate Google XML: %v", err)
	}

	xmlStr := string(xmlBytes)
	if !strings.Contains(xmlStr, "<rss version=\"2.0\" xmlns:g=\"http://base.google.com/ns/1.0\">") {
		t.Errorf("missing Google XML RSS namespace header")
	}
	if !strings.Contains(xmlStr, "<g:id>LIN-001</g:id>") {
		t.Errorf("missing g:id element in XML: %s", xmlStr)
	}
	if !strings.Contains(xmlStr, "<g:google_product_category>Apparel &amp; Accessories &gt; Clothing &gt; Shirts &amp; Tops</g:google_product_category>") &&
		!strings.Contains(xmlStr, "Apparel & Accessories > Clothing > Shirts & Tops") {
		t.Errorf("missing category in XML: %s", xmlStr)
	}
}
