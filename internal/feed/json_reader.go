package feed

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/alexandrmotologa/skupilot/internal/domain"
)

// ParseJSON parses a JSON catalog feed supporting both flat lists and wrapped structures.
func ParseJSON(r io.Reader) ([]domain.RawFeedItem, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read JSON feed: %w", err)
	}

	trimmed := strings.TrimSpace(string(data))
	if strings.HasPrefix(trimmed, "[") {
		return parseJSONArray(data)
	}

	return parseJSONObject(data)
}

func parseJSONArray(data []byte) ([]domain.RawFeedItem, error) {
	var rawList []map[string]interface{}
	if err := json.Unmarshal(data, &rawList); err != nil {
		return nil, fmt.Errorf("failed to parse JSON product array: %w", err)
	}

	var items []domain.RawFeedItem
	for i, m := range rawList {
		item := mapToRawItem(m, i+1)
		if item.Title != "" || item.ImageURL != "" {
			items = append(items, item)
		}
	}
	return items, nil
}

func parseJSONObject(data []byte) ([]domain.RawFeedItem, error) {
	// First check if it's a standard wrapped object with "products" or "items"
	var wrapper map[string]interface{}
	if err := json.Unmarshal(data, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse JSON object: %w", err)
	}

	var list []interface{}
	if pList, ok := wrapper["products"].([]interface{}); ok {
		list = pList
	} else if iList, ok := wrapper["items"].([]interface{}); ok {
		list = iList
	} else if dList, ok := wrapper["data"].([]interface{}); ok {
		list = dList
	}

	if len(list) == 0 {
		return nil, fmt.Errorf("no product array found in JSON root (expected 'products' or 'items')")
	}

	var items []domain.RawFeedItem
	for i, elem := range list {
		if m, ok := elem.(map[string]interface{}); ok {
			item := mapToRawItem(m, i+1)
			if item.Title != "" || item.ImageURL != "" {
				items = append(items, item)
			}
		}
	}

	return items, nil
}

func mapToRawItem(m map[string]interface{}, idx int) domain.RawFeedItem {
	item := domain.RawFeedItem{
		Currency: "USD",
	}

	// Helper to extract string from multiple possible keys
	getString := func(keys ...string) string {
		for _, k := range keys {
			if v, ok := m[k]; ok && v != nil {
				return strings.TrimSpace(fmt.Sprintf("%v", v))
			}
		}
		return ""
	}

	item.SKU = getString("sku", "product_id", "id", "item_id", "code")
	item.Title = getString("title", "product_title", "name", "product_name")
	item.Description = getString("description", "desc", "body_html", "details")
	item.Category = getString("category", "product_type", "type", "department")
	item.Brand = getString("brand", "vendor", "manufacturer")
	item.ImageURL = getString("image_url", "image", "photo", "img", "src")

	if cur := getString("currency", "currency_code"); cur != "" {
		item.Currency = strings.ToUpper(cur)
	}

	// Handle price as float or string
	for _, pKey := range []string{"price", "retail_price", "amount", "cost"} {
		if v, ok := m[pKey]; ok && v != nil {
			switch val := v.(type) {
			case float64:
				item.Price = val
			case string:
				cleaned := strings.Trim(val, "$€£ ")
				if p, err := strconv.ParseFloat(cleaned, 64); err == nil {
					item.Price = p
				}
			}
			break
		}
	}

	// Fallback for Shopify-style nested images: "images": [{"src": "..."}]
	if item.ImageURL == "" {
		if imgArr, ok := m["images"].([]interface{}); ok && len(imgArr) > 0 {
			if first, ok := imgArr[0].(map[string]interface{}); ok {
				if src, ok := first["src"].(string); ok {
					item.ImageURL = src
				}
			} else if firstStr, ok := imgArr[0].(string); ok {
				item.ImageURL = firstStr
			}
		}
	}

	// Fallback for Shopify-style nested variants: "variants": [{"price": "...", "sku": "..."}]
	if variantsArr, ok := m["variants"].([]interface{}); ok && len(variantsArr) > 0 {
		if firstVar, ok := variantsArr[0].(map[string]interface{}); ok {
			if item.SKU == "" {
				if vSKU, ok := firstVar["sku"].(string); ok {
					item.SKU = vSKU
				}
			}
			if item.Price == 0 {
				if vPrice, ok := firstVar["price"].(string); ok {
					if p, err := strconv.ParseFloat(vPrice, 64); err == nil {
						item.Price = p
					}
				}
			}
		}
	}

	if item.SKU == "" {
		item.SKU = fmt.Sprintf("SKU-%d", idx)
	}

	return item
}
