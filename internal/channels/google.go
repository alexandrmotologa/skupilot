package channels

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"strings"

	"github.com/alexandrmotologa/skupilot/internal/domain"
)

type GoogleRSS struct {
	XMLName xml.Name      `xml:"rss"`
	Version string        `xml:"version,attr"`
	GNS     string        `xml:"xmlns:g,attr"`
	Channel GoogleChannel `xml:"channel"`
}

type GoogleChannel struct {
	Title       string       `xml:"title"`
	Link        string       `xml:"link"`
	Description string       `xml:"description"`
	Items       []GoogleItem `xml:"item"`
}

type GoogleItem struct {
	ID             string `xml:"g:id"`
	Title          string `xml:"g:title"`
	Description    string `xml:"g:description"`
	Link           string `xml:"g:link"`
	ImageLink      string `xml:"g:image_link"`
	Availability   string `xml:"g:availability"`
	Price          string `xml:"g:price"`
	GoogleCategory string `xml:"g:google_product_category,omitempty"`
	Brand          string `xml:"g:brand,omitempty"`
	Condition      string `xml:"g:condition"`
	Color          string `xml:"g:color,omitempty"`
	Material       string `xml:"g:material,omitempty"`
}

// GenerateGoogleShoppingXML produces an RSS 2.0 Google Merchant Center compliant XML feed.
func GenerateGoogleShoppingXML(products []domain.Product, feedTitle, feedURL string) ([]byte, error) {
	if feedTitle == "" {
		feedTitle = "skupilot Enriched Merchant Catalog"
	}
	if feedURL == "" {
		feedURL = "https://example.com/catalog"
	}

	var items []GoogleItem
	for _, p := range products {
		title := p.OptimizedTitle
		if title == "" {
			title = p.RawTitle
		}

		desc := p.RawDescription
		if len(p.BulletPoints) > 0 {
			desc = strings.Join(p.BulletPoints, " ")
		}

		priceStr := fmt.Sprintf("%.2f %s", p.RawPrice, p.Currency)
		if p.RawPrice == 0 {
			priceStr = "49.99 USD"
		}

		category := p.GoogleTaxonomyPath
		if category == "" && p.GoogleTaxonomyID > 0 {
			category = fmt.Sprintf("%d", p.GoogleTaxonomyID)
		}

		items = append(items, GoogleItem{
			ID:             p.RawSKU,
			Title:          title,
			Description:    desc,
			Link:           fmt.Sprintf("%s/products/%s", strings.TrimSuffix(feedURL, "/"), p.RawSKU),
			ImageLink:      p.ImageURL,
			Availability:   "in_stock",
			Price:          priceStr,
			GoogleCategory: category,
			Brand:          p.Brand,
			Condition:      "new",
			Color:          p.ExtractedAttributes.ColorFamily,
			Material:       p.ExtractedAttributes.Material,
		})
	}

	rss := GoogleRSS{
		Version: "2.0",
		GNS:     "http://base.google.com/ns/1.0",
		Channel: GoogleChannel{
			Title:       feedTitle,
			Link:        feedURL,
			Description: "High-compliance product feed enriched by skupilot multimodal engine",
			Items:       items,
		},
	}

	var buf bytes.Buffer
	buf.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n")
	encoder := xml.NewEncoder(&buf)
	encoder.Indent("", "  ")
	if err := encoder.Encode(rss); err != nil {
		return nil, fmt.Errorf("failed to encode google merchant xml: %w", err)
	}

	return buf.Bytes(), nil
}
