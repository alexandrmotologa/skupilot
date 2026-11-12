package channels

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"strings"

	"github.com/alexandrmotologa/skupilot/internal/domain"
)

// GenerateEnrichedCSV produces a universal Product Information Management (PIM) CSV.
func GenerateEnrichedCSV(products []domain.Product) ([]byte, error) {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	headers := []string{
		"sku",
		"raw_title",
		"brand",
		"category",
		"price",
		"currency",
		"image_url",
		"image_width",
		"image_height",
		"image_zoom_ready",
		"image_white_bg_score",
		"image_compliance_status",
		"extracted_material",
		"extracted_fit",
		"extracted_collar",
		"extracted_sleeve",
		"extracted_closure",
		"extracted_color_hex",
		"extracted_color_family",
		"extracted_care",
		"google_taxonomy_id",
		"google_taxonomy_path",
		"quality_score",
		"publication_status",
		"optimized_title",
		"bullet_1",
		"bullet_2",
		"bullet_3",
		"bullet_4",
		"bullet_5",
		"seo_description_html",
	}

	if err := writer.Write(headers); err != nil {
		return nil, err
	}

	for _, p := range products {
		b1, b2, b3, b4, b5 := "", "", "", "", ""
		if len(p.BulletPoints) > 0 {
			b1 = p.BulletPoints[0]
		}
		if len(p.BulletPoints) > 1 {
			b2 = p.BulletPoints[1]
		}
		if len(p.BulletPoints) > 2 {
			b3 = p.BulletPoints[2]
		}
		if len(p.BulletPoints) > 3 {
			b4 = p.BulletPoints[3]
		}
		if len(p.BulletPoints) > 4 {
			b5 = p.BulletPoints[4]
		}

		zoomReadyStr := "false"
		if p.ImageZoomReady {
			zoomReadyStr = "true"
		}

		row := []string{
			p.RawSKU,
			p.RawTitle,
			p.Brand,
			p.RawCategory,
			fmt.Sprintf("%.2f", p.RawPrice),
			p.Currency,
			p.ImageURL,
			fmt.Sprintf("%d", p.ImageWidth),
			fmt.Sprintf("%d", p.ImageHeight),
			zoomReadyStr,
			fmt.Sprintf("%.3f", p.ImageWhiteBGScore),
			string(p.ImageComplianceStatus),
			p.ExtractedAttributes.Material,
			p.ExtractedAttributes.Fit,
			p.ExtractedAttributes.Collar,
			p.ExtractedAttributes.Sleeve,
			p.ExtractedAttributes.Closure,
			p.ExtractedAttributes.ColorHex,
			p.ExtractedAttributes.ColorFamily,
			p.ExtractedAttributes.Care,
			fmt.Sprintf("%d", p.GoogleTaxonomyID),
			p.GoogleTaxonomyPath,
			fmt.Sprintf("%d", p.QualityScore),
			string(p.PublicationStatus),
			p.OptimizedTitle,
			b1,
			b2,
			b3,
			b4,
			b5,
			strings.ReplaceAll(p.SEODescriptionHTML, "\n", " "),
		}

		if err := writer.Write(row); err != nil {
			return nil, err
		}
	}

	writer.Flush()
	return buf.Bytes(), writer.Error()
}
