package channels

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/alexandrmotologa/skupilot/internal/domain"
)

// ShopifyMetafield models a typed Shopify metafield.
type ShopifyMetafield struct {
	Namespace string `json:"namespace"`
	Key       string `json:"key"`
	Value     string `json:"value"`
	Type      string `json:"type"`
}

// ShopifyProductInput represents the GraphQL productCreate mutation variable input.
type ShopifyProductInput struct {
	Title           string             `json:"title"`
	DescriptionHTML string             `json:"descriptionHtml"`
	Vendor          string             `json:"vendor"`
	ProductType     string             `json:"productType"`
	Status          string             `json:"status"` // "DRAFT" or "ACTIVE"
	Tags            []string           `json:"tags"`
	Metafields      []ShopifyMetafield `json:"metafields"`
}

// BuildShopifyProductInput transforms an enriched domain.Product into Shopify GraphQL input.
func BuildShopifyProductInput(p domain.Product, publishStatus string) ShopifyProductInput {
	if publishStatus == "" {
		publishStatus = "DRAFT"
	}

	tags := []string{"skupilot-enriched"}
	if p.ExtractedAttributes.ColorFamily != "" {
		tags = append(tags, p.ExtractedAttributes.ColorFamily)
	}
	if p.ExtractedAttributes.Material != "" {
		tags = append(tags, p.ExtractedAttributes.Material)
	}
	if p.ExtractedAttributes.Fit != "" {
		tags = append(tags, p.ExtractedAttributes.Fit)
	}

	metafields := []ShopifyMetafield{
		{Namespace: "custom", Key: "quality_score", Value: fmt.Sprintf("%d", p.QualityScore), Type: "number_integer"},
	}

	if p.ExtractedAttributes.Material != "" {
		metafields = append(metafields, ShopifyMetafield{
			Namespace: "custom", Key: "fabric_material", Value: p.ExtractedAttributes.Material, Type: "single_line_text_field",
		})
	}
	if p.ExtractedAttributes.Collar != "" {
		metafields = append(metafields, ShopifyMetafield{
			Namespace: "custom", Key: "collar_style", Value: p.ExtractedAttributes.Collar, Type: "single_line_text_field",
		})
	}
	if p.ExtractedAttributes.Fit != "" {
		metafields = append(metafields, ShopifyMetafield{
			Namespace: "custom", Key: "fit_silhouette", Value: p.ExtractedAttributes.Fit, Type: "single_line_text_field",
		})
	}
	if p.ExtractedAttributes.ColorHex != "" {
		metafields = append(metafields, ShopifyMetafield{
			Namespace: "custom", Key: "hex_color", Value: p.ExtractedAttributes.ColorHex, Type: "single_line_text_field",
		})
	}

	title := p.OptimizedTitle
	if title == "" {
		title = p.RawTitle
	}

	return ShopifyProductInput{
		Title:           title,
		DescriptionHTML: p.SEODescriptionHTML,
		Vendor:          p.Brand,
		ProductType:     p.RawCategory,
		Status:          publishStatus,
		Tags:            tags,
		Metafields:      metafields,
	}
}

// ShopifyClient coordinates publishing products via Shopify GraphQL Admin API.
type ShopifyClient struct {
	StoreDomain string
	AccessToken string
	HTTPClient  *http.Client
}

func NewShopifyClient(storeDomain, token string) *ShopifyClient {
	cleanDomain := strings.TrimPrefix(storeDomain, "https://")
	cleanDomain = strings.TrimPrefix(cleanDomain, "http://")
	cleanDomain = strings.TrimSuffix(cleanDomain, "/")

	return &ShopifyClient{
		StoreDomain: cleanDomain,
		AccessToken: token,
		HTTPClient:  &http.Client{Timeout: 30 * time.Second},
	}
}

// PublishProduct creates a product on Shopify using GraphQL mutation.
func (c *ShopifyClient) PublishProduct(ctx context.Context, p domain.Product, publishStatus string) (string, error) {
	if c.AccessToken == "" || c.StoreDomain == "" {
		return "", fmt.Errorf("shopify store domain and access token required")
	}

	input := BuildShopifyProductInput(p, publishStatus)

	mutation := `
	mutation productCreate($input: ProductInput!) {
		productCreate(input: $input) {
			product {
				id
				title
				status
			}
			userErrors {
				field
				message
			}
		}
	}
	`

	reqPayload := map[string]interface{}{
		"query": mutation,
		"variables": map[string]interface{}{
			"input": input,
		},
	}

	bodyBytes, err := json.Marshal(reqPayload)
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf("https://%s/admin/api/2024-10/graphql.json", c.StoreDomain)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Shopify-Access-Token", c.AccessToken)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("shopify request failed: %w", err)
	}
	defer resp.Body.Close()

	respBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("shopify api returned status %d: %s", resp.StatusCode, string(respBytes))
	}

	var res struct {
		Data struct {
			ProductCreate struct {
				Product struct {
					ID    string `json:"id"`
					Title string `json:"title"`
				} `json:"product"`
				UserErrors []struct {
					Field   []string `json:"field"`
					Message string   `json:"message"`
				} `json:"userErrors"`
			} `json:"productCreate"`
		} `json:"data"`
	}

	if err := json.Unmarshal(respBytes, &res); err != nil {
		return "", fmt.Errorf("failed to parse shopify response: %w", err)
	}

	if len(res.Data.ProductCreate.UserErrors) > 0 {
		return "", fmt.Errorf("shopify error: %s", res.Data.ProductCreate.UserErrors[0].Message)
	}

	return res.Data.ProductCreate.Product.ID, nil
}
