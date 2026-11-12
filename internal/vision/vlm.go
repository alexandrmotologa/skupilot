package vision

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/alexandrmotologa/skupilot/internal/domain"
)

// VLMProvider is the interface for multimodal attribute extraction.
type VLMProvider interface {
	ExtractAttributes(ctx context.Context, img image.Image, raw domain.RawFeedItem) (*domain.ExtractedAttributes, error)
}

// ChainedVLM tries cloud providers first and falls back to deterministic heuristic analysis.
type ChainedVLM struct {
	providers []VLMProvider
}

// NewDefaultVLM creates a chained provider based on available environment credentials.
func NewDefaultVLM(geminiKey, openAIKey string) VLMProvider {
	var providers []VLMProvider

	if geminiKey != "" {
		providers = append(providers, NewGeminiVLM(geminiKey, "gemini-1.5-flash"))
	}
	if openAIKey != "" {
		providers = append(providers, NewOpenAIVLM(openAIKey, "gpt-4o-mini"))
	}

	// Always append the offline heuristic engine as the robust baseline fallback
	providers = append(providers, NewHeuristicVLM())

	return &ChainedVLM{providers: providers}
}

func (c *ChainedVLM) ExtractAttributes(ctx context.Context, img image.Image, raw domain.RawFeedItem) (*domain.ExtractedAttributes, error) {
	var lastErr error
	for _, p := range c.providers {
		attrs, err := p.ExtractAttributes(ctx, img, raw)
		if err == nil && attrs != nil {
			return attrs, nil
		}
		lastErr = err
	}
	return nil, fmt.Errorf("all VLM providers failed, last error: %w", lastErr)
}

// -----------------------------------------------------------------------------
// Gemini VLM Implementation
// -----------------------------------------------------------------------------

type GeminiVLM struct {
	apiKey     string
	model      string
	httpClient *http.Client
}

func NewGeminiVLM(apiKey, model string) *GeminiVLM {
	if model == "" {
		model = "gemini-1.5-flash"
	}
	return &GeminiVLM{
		apiKey:     apiKey,
		model:      model,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func (g *GeminiVLM) ExtractAttributes(ctx context.Context, img image.Image, raw domain.RawFeedItem) (*domain.ExtractedAttributes, error) {
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", g.model, g.apiKey)

	userText := fmt.Sprintf("Analyze this product image for SKU %s.\nTitle: %s\nDescription: %s\nCategory: %s\nBrand: %s",
		raw.SKU, raw.Title, raw.Description, raw.Category, raw.Brand)

	parts := []map[string]interface{}{
		{"text": SystemPromptForVLM + "\n\n" + userText},
	}

	if img != nil {
		buf := new(bytes.Buffer)
		if err := jpeg.Encode(buf, img, &jpeg.Options{Quality: 80}); err == nil {
			b64Img := base64.StdEncoding.EncodeToString(buf.Bytes())
			parts = append(parts, map[string]interface{}{
				"inline_data": map[string]interface{}{
					"mime_type": "image/jpeg",
					"data":      b64Img,
				},
			})
		}
	}

	payload := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"role":  "user",
				"parts": parts,
			},
		},
		"generationConfig": map[string]interface{}{
			"response_mime_type": "application/json",
			"temperature":        0.1,
		},
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("gemini api returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var geminiResp struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&geminiResp); err != nil {
		return nil, err
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("empty response from gemini vlm")
	}

	rawJSON := geminiResp.Candidates[0].Content.Parts[0].Text
	rawJSON = cleanJSONString(rawJSON)

	var attrs domain.ExtractedAttributes
	if err := json.Unmarshal([]byte(rawJSON), &attrs); err != nil {
		return nil, fmt.Errorf("failed to parse gemini json: %w", err)
	}

	return &attrs, nil
}

// -----------------------------------------------------------------------------
// OpenAI VLM Implementation
// -----------------------------------------------------------------------------

type OpenAIVLM struct {
	apiKey     string
	model      string
	httpClient *http.Client
}

func NewOpenAIVLM(apiKey, model string) *OpenAIVLM {
	if model == "" {
		model = "gpt-4o-mini"
	}
	return &OpenAIVLM{
		apiKey:     apiKey,
		model:      model,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func (o *OpenAIVLM) ExtractAttributes(ctx context.Context, img image.Image, raw domain.RawFeedItem) (*domain.ExtractedAttributes, error) {
	url := "https://api.openai.com/v1/chat/completions"

	userContent := []map[string]interface{}{
		{
			"type": "text",
			"text": fmt.Sprintf("Analyze this product.\nTitle: %s\nDescription: %s\nCategory: %s\nBrand: %s",
				raw.Title, raw.Description, raw.Category, raw.Brand),
		},
	}

	if img != nil {
		buf := new(bytes.Buffer)
		if err := jpeg.Encode(buf, img, &jpeg.Options{Quality: 80}); err == nil {
			b64Img := base64.StdEncoding.EncodeToString(buf.Bytes())
			userContent = append(userContent, map[string]interface{}{
				"type": "image_url",
				"image_url": map[string]interface{}{
					"url": fmt.Sprintf("data:image/jpeg;base64,%s", b64Img),
				},
			})
		}
	}

	payload := map[string]interface{}{
		"model": o.model,
		"messages": []map[string]interface{}{
			{"role": "system", "content": SystemPromptForVLM},
			{"role": "user", "content": userContent},
		},
		"response_format": map[string]string{"type": "json_object"},
		"temperature":     0.1,
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+o.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := o.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("openai api returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var openAIResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&openAIResp); err != nil {
		return nil, err
	}

	if len(openAIResp.Choices) == 0 {
		return nil, fmt.Errorf("empty choices from openai vlm")
	}

	rawJSON := cleanJSONString(openAIResp.Choices[0].Message.Content)
	var attrs domain.ExtractedAttributes
	if err := json.Unmarshal([]byte(rawJSON), &attrs); err != nil {
		return nil, fmt.Errorf("failed to parse openai json: %w", err)
	}

	return &attrs, nil
}

// -----------------------------------------------------------------------------
// Offline Heuristic & Computer Vision VLM Implementation
// -----------------------------------------------------------------------------

type HeuristicVLM struct{}

func NewHeuristicVLM() *HeuristicVLM {
	return &HeuristicVLM{}
}

func (h *HeuristicVLM) ExtractAttributes(ctx context.Context, img image.Image, raw domain.RawFeedItem) (*domain.ExtractedAttributes, error) {
	attrs := &domain.ExtractedAttributes{
		Features: []string{},
	}

	// 1. Image Pixel Analysis (Color & Tone)
	if img != nil {
		hex, family := ExtractDominantColor(img)
		attrs.ColorHex = hex
		attrs.ColorFamily = family
	}

	// Combine all textual tokens for keyword analysis
	text := strings.ToLower(fmt.Sprintf("%s %s %s %s", raw.Title, raw.Description, raw.Category, raw.Brand))

	// 2. Material Detection
	materials := []struct {
		keyword string
		label   string
	}{
		{"linen", "100% Breathable Linen"},
		{"silk", "Pure Mulberry Silk"},
		{"merino wool", "100% Merino Wool"},
		{"cashmere", "Ultra-Soft Cashmere"},
		{"wool", "Premium Wool Blend"},
		{"flannel", "Brushed Cotton Flannel"},
		{"fleece", "Heavyweight Cotton Fleece"},
		{"denim", "100% Washed Cotton Denim"},
		{"leather", "Full-Grain Leather"},
		{"calfskin", "Italian Calfskin Leather"},
		{"canvas", "Heavyweight Waxed Canvas"},
		{"titanium", "Aerospace-Grade Titanium"},
		{"cotton", "100% Combed Cotton"},
		{"mesh", "Breathable Engineered Mesh"},
		{"velvet", "Plush Cotton Velvet"},
	}
	for _, m := range materials {
		if strings.Contains(text, m.keyword) {
			attrs.Material = m.label
			break
		}
	}
	if attrs.Material == "" {
		attrs.Material = "Premium Cotton Blend"
	}

	// 3. Collar & Neckline Detection
	collars := []struct {
		keyword string
		collar  string
		neck    string
	}{
		{"mandarin", "Mandarin Band Collar", "Band Collar"},
		{"band collar", "Mandarin Band Collar", "Band Collar"},
		{"cowl neck", "", "Cowl Neckline"},
		{"crewneck", "", "Classic Crew Neck"},
		{"crew neck", "", "Classic Crew Neck"},
		{"v-neck", "", "V-Neckline"},
		{"peak lapel", "Tailored Peak Lapel", "Notch Lapel"},
		{"notch lapel", "Classic Notch Lapel", "Notch Lapel"},
		{"hooded", "Drawstring Hood", "Hooded"},
		{"hoodie", "Drawstring Hood", "Hooded"},
		{"button-down", "Button-Down Point Collar", "Point Collar"},
	}
	for _, c := range collars {
		if strings.Contains(text, c.keyword) {
			if c.collar != "" {
				attrs.Collar = c.collar
			}
			if c.neck != "" {
				attrs.Neckline = c.neck
			}
			break
		}
	}

	// 4. Fit & Silhouette
	fits := []struct {
		keyword string
		label   string
	}{
		{"slim", "Tailored Slim Fit"},
		{"oversized", "Relaxed Oversized Fit"},
		{"relaxed", "Relaxed Casual Fit"},
		{"pleated", "A-Line Pleated Silhouette"},
		{"tailored", "Structured Tailored Fit"},
		{"tapered", "Tapered Athletic Fit"},
		{"regular", "Classic Regular Fit"},
	}
	for _, f := range fits {
		if strings.Contains(text, f.keyword) {
			attrs.Fit = f.label
			break
		}
	}
	if attrs.Fit == "" {
		attrs.Fit = "Modern Regular Fit"
	}

	// 5. Sleeve Length
	if strings.Contains(text, "short sleeve") || strings.Contains(text, "t-shirt") || strings.Contains(text, "tee") {
		attrs.Sleeve = "Short Sleeve"
	} else if strings.Contains(text, "sleeveless") || strings.Contains(text, "tank") || strings.Contains(text, "slip dress") {
		attrs.Sleeve = "Sleeveless"
	} else if strings.Contains(text, "long sleeve") || strings.Contains(text, "sweater") || strings.Contains(text, "hoodie") || strings.Contains(text, "jacket") || strings.Contains(text, "blazer") {
		attrs.Sleeve = "Long Sleeve"
	}

	// 6. Closure Mechanism
	if strings.Contains(text, "button") || strings.Contains(text, "shank") {
		attrs.Closure = "Reinforced Button Placket"
	} else if strings.Contains(text, "pullover") || strings.Contains(text, "hoodie") || strings.Contains(text, "sweater") {
		attrs.Closure = "Pullover Overhead"
	} else if strings.Contains(text, "zip") || strings.Contains(text, "zipper") {
		attrs.Closure = "Full-Zip Fastening"
	} else if strings.Contains(text, "lace") || strings.Contains(text, "sneakers") {
		attrs.Closure = "Waxed Lace-Up"
	} else if strings.Contains(text, "elastic") || strings.Contains(text, "slip-on") {
		attrs.Closure = "Elastic Slip-On Gusset"
	}

	// 7. Pattern
	if strings.Contains(text, "plaid") || strings.Contains(text, "check") {
		attrs.Pattern = "Classic Buffalo Plaid"
	} else if strings.Contains(text, "striped") || strings.Contains(text, "stripe") {
		attrs.Pattern = "Subtle Yarn-Dyed Stripe"
	} else if strings.Contains(text, "heather") {
		attrs.Pattern = "Textured Heather Knit"
	} else if strings.Contains(text, "floral") {
		attrs.Pattern = "Botanical Floral Print"
	} else {
		attrs.Pattern = "Clean Solid Color"
	}

	// 8. Care Instructions
	if strings.Contains(attrs.Material, "Silk") || strings.Contains(attrs.Material, "Wool") || strings.Contains(attrs.Material, "Cashmere") {
		attrs.Care = "Dry Clean Only / Gentle Hand Wash Cold"
	} else if strings.Contains(attrs.Material, "Leather") {
		attrs.Care = "Specialist Leather Clean Only"
	} else {
		attrs.Care = "Machine Wash Cold Gentle, Tumble Dry Low"
	}

	// 9. Highlight Features
	if attrs.Material != "" {
		attrs.Features = append(attrs.Features, fmt.Sprintf("Crafted from %s", attrs.Material))
	}
	if attrs.Fit != "" {
		attrs.Features = append(attrs.Features, fmt.Sprintf("Cut with %s", attrs.Fit))
	}
	if attrs.Closure != "" {
		attrs.Features = append(attrs.Features, fmt.Sprintf("Constructed with %s", attrs.Closure))
	}

	return attrs, nil
}

// cleanJSONString removes possible markdown code blocks ```json ... ``` from LLM outputs.
func cleanJSONString(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "```json") {
		s = strings.TrimPrefix(s, "```json")
	} else if strings.HasPrefix(s, "```") {
		s = strings.TrimPrefix(s, "```")
	}
	if strings.HasSuffix(s, "```") {
		s = strings.TrimSuffix(s, "```")
	}
	return strings.TrimSpace(s)
}
