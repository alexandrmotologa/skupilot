package feed

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/alexandrmotologa/skupilot/internal/domain"
)

// ParseCSV reads raw product records from a CSV reader with flexible header mapping.
func ParseCSV(r io.Reader) ([]domain.RawFeedItem, error) {
	reader := csv.NewReader(r)
	reader.TrimLeadingSpace = true
	reader.FieldsPerRecord = -1 // Allow variable fields if trailing commas

	headers, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV headers: %w", err)
	}

	headerMap := normalizeHeaders(headers)
	var items []domain.RawFeedItem
	rowIdx := 1

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed reading CSV row %d: %w", rowIdx, err)
		}
		rowIdx++

		item := domain.RawFeedItem{
			Currency: "USD", // sensible e-commerce default
		}

		for colName, idx := range headerMap {
			if idx >= len(record) {
				continue
			}
			val := strings.TrimSpace(record[idx])

			switch colName {
			case "sku":
				item.SKU = val
			case "title":
				item.Title = val
			case "description":
				item.Description = val
			case "category":
				item.Category = val
			case "brand":
				item.Brand = val
			case "image_url":
				item.ImageURL = val
			case "currency":
				if val != "" {
					item.Currency = strings.ToUpper(val)
				}
			case "price":
				// Strip currency symbols if present
				cleaned := strings.TrimPrefix(val, "$")
				cleaned = strings.TrimPrefix(cleaned, "€")
				cleaned = strings.TrimPrefix(cleaned, "£")
				cleaned = strings.TrimSpace(cleaned)
				if p, err := strconv.ParseFloat(cleaned, 64); err == nil {
					item.Price = p
				}
			}
		}

		// Fallback for SKU if missing
		if item.SKU == "" && item.Title != "" {
			item.SKU = fmt.Sprintf("SKU-%d", rowIdx)
		}

		if item.Title != "" || item.ImageURL != "" {
			items = append(items, item)
		}
	}

	return items, nil
}

// normalizeHeaders maps common header variations to standard domain keys.
func normalizeHeaders(headers []string) map[string]int {
	m := make(map[string]int)
	for i, h := range headers {
		clean := strings.ToLower(strings.TrimSpace(h))
		clean = strings.ReplaceAll(clean, "_", "")
		clean = strings.ReplaceAll(clean, "-", "")
		clean = strings.ReplaceAll(clean, " ", "")

		switch clean {
		case "sku", "productid", "itemid", "code", "modelnumber", "variantid":
			m["sku"] = i
		case "title", "producttitle", "name", "productname", "itemname", "header":
			m["title"] = i
		case "description", "desc", "bodyhtml", "body", "details":
			m["description"] = i
		case "category", "producttype", "type", "department", "cat":
			m["category"] = i
		case "price", "retailprice", "unitprice", "cost", "amount":
			m["price"] = i
		case "currency", "currencycode":
			m["currency"] = i
		case "brand", "vendor", "manufacturer", "designer":
			m["brand"] = i
		case "imageurl", "image", "img", "photourl", "photo", "picture", "imagesrc", "images":
			m["image_url"] = i
		}
	}
	return m
}
