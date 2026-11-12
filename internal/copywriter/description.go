package copywriter

import (
	"fmt"
	"strings"

	"github.com/alexandrmotologa/skupilot/internal/domain"
)

// GenerateSEODescriptionHTML crafts a clean semantic HTML description suitable for Shopify and marketplace listings.
func GenerateSEODescriptionHTML(raw domain.RawFeedItem, attr domain.ExtractedAttributes, bullets []string) string {
	var sb strings.Builder

	// Brand / Opening paragraph
	sb.WriteString("<div class=\"product-description\">\n")
	sb.WriteString(fmt.Sprintf("  <p><strong>The %s</strong> by %s delivers elevated design with an uncompromising focus on everyday functionality and comfort.</p>\n",
		raw.Title, raw.Brand))

	// Highlight overview
	if raw.Description != "" {
		sb.WriteString(fmt.Sprintf("  <p>%s</p>\n", raw.Description))
	}

	// Bulleted highlights
	sb.WriteString("  <h3>Key Highlights & Specifications</h3>\n")
	sb.WriteString("  <ul>\n")
	for _, b := range bullets {
		parts := strings.SplitN(b, ":", 2)
		if len(parts) == 2 {
			sb.WriteString(fmt.Sprintf("    <li><strong>%s:</strong>%s</li>\n", strings.TrimSpace(parts[0]), parts[1]))
		} else {
			sb.WriteString(fmt.Sprintf("    <li>%s</li>\n", b))
		}
	}
	sb.WriteString("  </ul>\n")

	// Technical Details Grid
	sb.WriteString("  <h3>Product Details</h3>\n")
	sb.WriteString("  <ul>\n")
	if attr.Material != "" {
		sb.WriteString(fmt.Sprintf("    <li><strong>Material:</strong> %s</li>\n", attr.Material))
	}
	if attr.Fit != "" {
		sb.WriteString(fmt.Sprintf("    <li><strong>Fit:</strong> %s</li>\n", attr.Fit))
	}
	if attr.Collar != "" {
		sb.WriteString(fmt.Sprintf("    <li><strong>Collar:</strong> %s</li>\n", attr.Collar))
	}
	if attr.Neckline != "" {
		sb.WriteString(fmt.Sprintf("    <li><strong>Neckline:</strong> %s</li>\n", attr.Neckline))
	}
	if attr.Closure != "" {
		sb.WriteString(fmt.Sprintf("    <li><strong>Closure:</strong> %s</li>\n", attr.Closure))
	}
	if attr.Pattern != "" {
		sb.WriteString(fmt.Sprintf("    <li><strong>Pattern:</strong> %s</li>\n", attr.Pattern))
	}
	if attr.Care != "" {
		sb.WriteString(fmt.Sprintf("    <li><strong>Care:</strong> %s</li>\n", attr.Care))
	}
	sb.WriteString("  </ul>\n")
	sb.WriteString("</div>")

	return sb.String()
}
