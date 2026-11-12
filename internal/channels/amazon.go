package channels

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/alexandrmotologa/skupilot/internal/domain"
)

// GenerateAmazonFlatFile produces a valid tab-separated Amazon Inventory Loader file.
func GenerateAmazonFlatFile(products []domain.Product) []byte {
	var buf bytes.Buffer

	// Standard Amazon Inventory Loader TSV header row
	headers := []string{
		"item_sku",
		"external_product_id",
		"external_product_id_type",
		"item_name",
		"brand_name",
		"feed_product_type",
		"standard_price",
		"quantity",
		"main_image_url",
		"bullet_point1",
		"bullet_point2",
		"bullet_point3",
		"bullet_point4",
		"bullet_point5",
		"color_name",
		"color_map",
		"material_type",
		"closure_type",
	}

	buf.WriteString(strings.Join(headers, "\t") + "\n")

	for _, p := range products {
		title := p.OptimizedTitle
		if title == "" {
			title = p.RawTitle
		}

		bullet1, bullet2, bullet3, bullet4, bullet5 := "", "", "", "", ""
		if len(p.BulletPoints) > 0 {
			bullet1 = cleanTSV(p.BulletPoints[0])
		}
		if len(p.BulletPoints) > 1 {
			bullet2 = cleanTSV(p.BulletPoints[1])
		}
		if len(p.BulletPoints) > 2 {
			bullet3 = cleanTSV(p.BulletPoints[2])
		}
		if len(p.BulletPoints) > 3 {
			bullet4 = cleanTSV(p.BulletPoints[3])
		}
		if len(p.BulletPoints) > 4 {
			bullet5 = cleanTSV(p.BulletPoints[4])
		}

		priceStr := fmt.Sprintf("%.2f", p.RawPrice)

		row := []string{
			cleanTSV(p.RawSKU),
			cleanTSV(p.RawSKU), // fallback external id
			"SKU",
			cleanTSV(title),
			cleanTSV(p.Brand),
			cleanTSV(p.RawCategory),
			priceStr,
			"100", // default initial quantity
			cleanTSV(p.ImageURL),
			bullet1,
			bullet2,
			bullet3,
			bullet4,
			bullet5,
			cleanTSV(p.ExtractedAttributes.ColorFamily),
			cleanTSV(p.ExtractedAttributes.ColorFamily),
			cleanTSV(p.ExtractedAttributes.Material),
			cleanTSV(p.ExtractedAttributes.Closure),
		}

		buf.WriteString(strings.Join(row, "\t") + "\n")
	}

	return buf.Bytes()
}

func cleanTSV(s string) string {
	s = strings.ReplaceAll(s, "\t", " ")
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	return strings.TrimSpace(s)
}
