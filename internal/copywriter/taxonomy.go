package copywriter

import "strings"

// GoogleTaxonomyEntry represents an item in Google's Product Taxonomy.
type GoogleTaxonomyEntry struct {
	ID   int
	Path string
}

// MapGoogleTaxonomy maps a raw category or product title to official Google Product Taxonomy ID and breadcrumbs.
func MapGoogleTaxonomy(category, title string) GoogleTaxonomyEntry {
	combined := strings.ToLower(category + " " + title)

	if strings.Contains(combined, "shoe") || strings.Contains(combined, "boot") || strings.Contains(combined, "sneaker") || strings.Contains(combined, "footwear") {
		return GoogleTaxonomyEntry{
			ID:   187,
			Path: "Apparel & Accessories > Shoes",
		}
	}

	if strings.Contains(combined, "wallet") || strings.Contains(combined, "bag") || strings.Contains(combined, "tote") {
		return GoogleTaxonomyEntry{
			ID:   6552,
			Path: "Apparel & Accessories > Handbags, Wallets & Cases",
		}
	}

	if strings.Contains(combined, "sunglass") || strings.Contains(combined, "eyewear") {
		return GoogleTaxonomyEntry{
			ID:   178,
			Path: "Apparel & Accessories > Clothing Accessories > Sunglasses",
		}
	}

	if strings.Contains(combined, "dress") || strings.Contains(combined, "gown") {
		return GoogleTaxonomyEntry{
			ID:   227,
			Path: "Apparel & Accessories > Clothing > Dresses",
		}
	}

	if strings.Contains(combined, "skirt") {
		return GoogleTaxonomyEntry{
			ID:   1581,
			Path: "Apparel & Accessories > Clothing > Skirts",
		}
	}

	if strings.Contains(combined, "jacket") || strings.Contains(combined, "blazer") || strings.Contains(combined, "coat") || strings.Contains(combined, "outerwear") {
		return GoogleTaxonomyEntry{
			ID:   213,
			Path: "Apparel & Accessories > Clothing > Outerwear > Coats & Jackets",
		}
	}

	if strings.Contains(combined, "hoodie") || strings.Contains(combined, "sweatshirt") || strings.Contains(combined, "sweater") {
		return GoogleTaxonomyEntry{
			ID:   5322,
			Path: "Apparel & Accessories > Clothing > Shirts & Tops > Sweaters & Knitwear",
		}
	}

	if strings.Contains(combined, "shirt") || strings.Contains(combined, "tee") || strings.Contains(combined, "top") || strings.Contains(combined, "flannel") {
		return GoogleTaxonomyEntry{
			ID:   212,
			Path: "Apparel & Accessories > Clothing > Shirts & Tops",
		}
	}

	if strings.Contains(combined, "pant") || strings.Contains(combined, "trouser") || strings.Contains(combined, "jean") {
		return GoogleTaxonomyEntry{
			ID:   204,
			Path: "Apparel & Accessories > Clothing > Pants",
		}
	}

	// Default Apparel Root
	return GoogleTaxonomyEntry{
		ID:   1604,
		Path: "Apparel & Accessories > Clothing",
	}
}
