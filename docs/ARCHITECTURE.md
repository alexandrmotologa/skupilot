# skupilot Architecture

skupilot is an autonomous e-commerce catalog enrichment and marketplace compliance engine. It is designed as a single-binary service written in Go, embedding a React 19 merchandising studio.

The system ingests raw, unstructured product feeds from suppliers, performs mathematical image audits against marketplace requirements, extracts technical attributes using multimodal models and computer vision, synthesizes marketplace copy, and exports ready-to-publish feeds to Shopify, Amazon, and Google Shopping.

---

## System Components

```
+--------------------------------------------------------------------------+
|                            skupilot Core Engine                          |
+--------------------------------------------------------------------------+
| 1. Feed Ingestion                                                        |
|    - High-throughput CSV, TSV, and JSON streaming parsers                |
|    - Shopify products export reader                                      |
+--------------------------------------------------------------------------+
| 2. Image Compliance Auditor                                              |
|    - Mathematical RGB 255 pure white background calculator               |
|    - Minimum 1000x1000 px zoom readiness check                           |
|    - Aspect ratio range validator (0.8 to 1.25)                          |
|    - Watermark and text overlay edge detector                            |
|    - Automated border padding and background whitening filter            |
+--------------------------------------------------------------------------+
| 3. Multimodal VLM Extractor                                              |
|    - Visual attribute extraction for apparel, footwear, and accessories  |
|    - Extracts material, silhouette, collar, neckline, and closure        |
|    - Extracts exact hex color and color family                           |
|    - Dual-engine: Gemini / OpenAI VLM with offline heuristic fallback    |
+--------------------------------------------------------------------------+
| 4. Algorithmic Copywriter & Quality Scorer                               |
|    - Amazon A9 title synthesizer (150 to 200 characters limit)           |
|    - 5 benefit-driven feature bullets generator                          |
|    - Clean HTML SEO product description generator                        |
|    - Google Product Category taxonomy mapper (numeric ID and path)       |
|    - 0 to 100% Listing Quality Scorecard                                 |
+--------------------------------------------------------------------------+
| 5. Multi-Channel Exporter                                                |
|    - Shopify Admin GraphQL mutation builder and publisher                |
|    - Amazon Inventory Loader flat file (TSV) writer                      |
|    - Google Merchant Center XML feed generator                           |
+--------------------------------------------------------------------------+
| 6. Storage & Web API                                                     |
|    - Zero-CGO SQLite embedded database (modernc.org/sqlite)              |
|    - REST API with Server-Sent Events (SSE) for live streaming progress  |
|    - Embedded React 19 merchandising studio with full visual catalog     |
+--------------------------------------------------------------------------+
```

---

## Detailed Component Specifications

### 1. Feed Ingestion Engine (`internal/feed`)
The feed engine parses large supplier catalog files without loading the entire payload into RAM at once.
- `csv_reader.go`: Uses `encoding/csv` with buffered IO. Detects header aliases for SKU (`sku`, `product_id`, `item_id`), title (`title`, `product_name`, `name`), price, currency, category, and image URL (`image`, `image_url`, `photo`, `img`).
- `json_reader.go`: Uses streaming `json.Decoder` to parse arrays of products or nested catalogue objects.
- `shopify_importer.go`: Maps standard Shopify Product JSON dumps into internal domain entities.

### 2. Image Compliance Auditor (`internal/vision/compliance.go`)
Marketplaces like Amazon enforce strict image standards on main product images. Non-compliant listings are search-suppressed.

#### Whiteness Score Calculation
The auditor samples the perimeter of the image (a 10-pixel border band across all four sides). For every pixel `p = (R, G, B)` in the perimeter set `P_corner`:
```
whiteness_score = (count of pixels where R >= 250, G >= 250, B >= 250) / total perimeter pixels
```
- Amazon Pass requirement: `whiteness_score >= 0.985`.
- Warning threshold: `0.900 <= whiteness_score < 0.985`.
- Fail threshold: `whiteness_score < 0.900`.

#### Zoom Readiness Check
Amazon and Walmart require images to have at least 1000 pixels on the longest side (recommended 1000 x 1000 or greater) to activate zoom on product detail pages.
```
zoom_ready = (width >= 1000) AND (height >= 1000)
```

#### Watermark and Text Overlay Detection
The auditor inspects high-contrast edge gradients in the corner quadrants. If high-frequency variations occur in regions that should be empty background, the image is flagged for human review or automated cropping.

### 3. Multimodal VLM Extractor (`internal/vision/vlm.go`)
Product photos contain details that suppliers omit in raw feeds. The VLM extractor inspects the image pixels to extract structured attributes:
- Apparel: Material composition (e.g. linen, silk, cotton, denim), silhouette (e.g. slim fit, relaxed fit, A-line), collar style (e.g. mandarin, notch lapel, crew neck), sleeve length, closure type (e.g. button-front, zip, pullover), pattern (e.g. solid, striped, floral).
- Footwear: Upper material (e.g. full-grain leather, canvas, mesh), sole type (e.g. lugged rubber, stacked leather, EVA), toe style (e.g. almond, round, cap toe), fastening.
- Color: Dominant hex code (e.g. `#EAE6DF`) and standardized color family name (e.g. `Oatmeal Cream`, `Navy Blue`).

The extractor supports both cloud VLMs (Google Gemini 1.5/2.5 Flash, OpenAI GPT-4o-mini) and an integrated offline heuristic computer vision engine that uses Go image color quantization and text token heuristics to guarantee 100% test coverage and air-gapped operation.

### 4. Algorithmic Copywriter & Quality Scorecard (`internal/copywriter`)
- Title Synthesizer: Follows the Amazon A9 formula:
  `[Brand] + [Gender/Target] + [Material] + [Product Type] + [Key Feature / Silhouette] + [Color]`
  Strictly caps title lengths between 150 and 200 characters. Strips prohibited promotional claims like "Best Seller", "Free Shipping", or "100% Quality Guaranteed".
- 5 Feature Bullets: Produces five distinct bullet points highlighting:
  1. Material and Breathability
  2. Silhouette and Fit
  3. Construction and Hardware
  4. Care and Durability
  5. Styling and Occasion
- Quality Scorecard (0 to 100):
  - Visual Compliance (30 points): Pure white background, zoom dimensions, aspect ratio.
  - Attribute Completeness (30 points): Material, fit, color, collar/closure detected.
  - Title Optimization (20 points): Correct formula, length bounds, no prohibited words.
  - Bullet Points Quality (20 points): Exactly 5 bullets, length and structure checks.

### 5. Channel Exporters (`internal/channels`)
- Shopify GraphQL: Generates `productCreate` / `productUpdate` operations targeting the Shopify Admin API (2024-10 version), assigning tags, variants, and product metafields for extracted visual attributes.
- Amazon Flat File: Generates valid tab-separated (.tsv) inventory loader records matching Amazon Seller Central requirements.
- Google Merchant Center: Generates RSS 2.0 / Google Shopping XML feeds with exact Google Product Category IDs (e.g., `212` for Shirts & Tops).

### 6. Embedded Database & Studio (`internal/store`, `web/`)
The database uses `modernc.org/sqlite`, a pure Go SQLite implementation that does not require CGO. The React 19 frontend compiles into static assets and is embedded into the final Go binary with `//go:embed all:dist`.
