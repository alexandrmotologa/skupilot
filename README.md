# skupilot

> Autonomous Multimodal E-Commerce Catalog Enrichment and Marketplace Compliance Engine

skupilot is a self-hosted single-binary Go service with an embedded React 19 merchandising studio. It automates catalog ingestion, multimodal image auditing, attribute extraction, and marketplace copywriting for high-volume e-commerce brands, distributors, and multi-vendor marketplaces.

---

## The Problem

High-volume catalog ingestion faces recurring operational bottlenecks:
- Missing Technical Attributes: Over 60% of supplier feeds omit search filter attributes such as collar styles, exact materials, silhouette cuts, and color codes.
- Strict Image Standards: Marketplaces like Amazon automatically suppress listings whose main image lacks a pure white RGB 255 background, fails the 1000 x 1000 px zoom threshold, or contains watermarks.
- Subpar Conversion Copy: Raw supplier titles and descriptions are often unformatted, keyword-stuffed strings that reduce customer trust and search rank.
- Manual Labor Costs: Formatting new inventory manually costs between $5 and $15 per SKU and introduces delays.

## What skupilot Does

1. High-Throughput Catalog Ingestion: Streams large CSV, TSV, JSON, and Shopify product dumps without buffering entire datasets into memory.
2. Mathematical Image Compliance Auditor:
   - Measures corner and perimeter background whiteness using a strict RGB (255, 255, 255) purity threshold.
   - Verifies resolution zoom readiness (minimum 1000 x 1000 pixels).
   - Detects edge density and high-contrast watermarks or text overlays.
   - Validates aspect ratios against marketplace square standards (0.8 to 1.25).
3. Multimodal VLM Attribute Extraction:
   - Uses Vision-Language Models (Gemini 1.5/2.5 Flash, OpenAI GPT-4o-mini) and an offline computer vision heuristic engine to inspect product photos.
   - Extracts fabric composition, garment cut, collar type, neckline, closure mechanism, and exact hex color codes directly from pixels.
4. Algorithmic Copywriting & Quality Scoring:
   - Formulates Amazon A9 compliant titles between 150 and 200 characters, filtering prohibited promotional claims.
   - Synthesizes 5 structured, benefit-driven feature bullets.
   - Generates clean semantic HTML product descriptions.
   - Resolves Google Product Taxonomy numeric IDs and breadcrumb paths.
   - Computes a comprehensive Listing Quality Scorecard (0 to 100).
5. 1-Click Multi-Channel Sync:
   - Publishes directly to the Shopify Admin API using GraphQL mutations and custom metafields.
   - Generates Amazon Inventory Loader flat files (.tsv).
   - Exports Google Merchant Center product XML feeds.
6. Embedded React 19 Studio:
   - Single-binary deployment with zero CGO dependencies (pure Go SQLite).
   - Merchandising studio UI with real-time SSE progress streaming.

---

## Architecture Overview

```
                      +---------------------------------------+
                      |       Raw Supplier Catalog Feed       |
                      |   (CSV, TSV, JSON, Shopify Export)    |
                      +---------------------------------------+
                                          |
                                          v
                      +---------------------------------------+
                      |            skupilot Engine            |
                      |          (Single Go Binary)           |
                      +---------------------------------------+
                                          |
                        +-----------------+-----------------+
                        |                                   |
                        v                                   v
        +-------------------------------+   +-------------------------------+
        |    Image Compliance Auditor   |   |   Multimodal VLM Extractor    |
        |  - RGB 255 Pure White Check   |   |  - Fabric & Material Detect   |
        |  - Zoom Resolution >= 1000px  |   |  - Cut, Fit, & Silhouette     |
        |  - Watermark & Overlay Math   |   |  - Exact Hex & Color Family   |
        +-------------------------------+   +-------------------------------+
                        |                                   |
                        +-----------------+-----------------+
                                          |
                                          v
                      +---------------------------------------+
                      |     Algorithmic Copywriter Engine     |
                      |  - Amazon A9 Title Builder (150-200c) |
                      |  - 5 Benefit-Driven Feature Bullets   |
                      |  - Google Taxonomy Category Mapping   |
                      |  - Listing Quality Scorecard (0-100)  |
                      +---------------------------------------+
                                          |
                        +-----------------+-----------------+
                        |                                   |
                        v                                   v
        +-------------------------------+   +-------------------------------+
        |       Multi-Channel Sync      |   |    React 19 Studio (Embed)    |
        |  - Shopify GraphQL Admin API  |   |  - Visual Merchandising Grid  |
        |  - Amazon Flat File (.tsv)    |   |  - Image Defect Diagnostics   |
        |  - Google Merchant XML Feed   |   |  - Live SSE Batch Progress    |
        +-------------------------------+   +-------------------------------+
```

---

## Installation & Requirements

- Go 1.22 or higher
- Node.js 20 or higher (for building the web studio)
- Modern web browser

### Build from Source

```bash
# Clone the repository
git clone https://github.com/alexandrmotologa/skupilot.git
cd skupilot

# Build the embedded React 19 studio
cd web
npm install
npm run build
cd ..

# Build the single Go binary
go build -o bin/skupilot ./cmd/skupilot
```

---

## Quick Start

### 1. Launch the Merchandising Studio
```bash
./bin/skupilot serve --port 8080 --db ./catalog.db
```
Open your browser at `http://localhost:8080` to access the visual merchandising studio.

### 2. Ingest and Enrich a Catalog Feed via CLI
```bash
./bin/skupilot enrich \
  --input ./examples/sample_apparel_feed.csv \
  --category apparel \
  --model gemini-2.5-flash \
  --db ./catalog.db
```

### 3. Run Marketplace Compliance Audit
```bash
./bin/skupilot audit --db ./catalog.db --strict-amazon
```

### 4. Auto-Remediate Non-Compliant Images
```bash
./bin/skupilot autofix --db ./catalog.db
```
Whitens outer perimeter pixels to pure RGB 255 and pads images onto a 1000 x 1000 px square canvas, elevating compliance pass rates to 100%.

### 5. Export to Channels

#### Export to Master PIM CSV
```bash
./bin/skupilot export csv --batch-id batch_default --output ./dist/enriched_catalog.csv
```

#### Export to Amazon Inventory Loader (TSV)
```bash
./bin/skupilot export amazon --batch-id batch_default --output ./dist/amazon_feed.tsv
```

#### Export to Google Merchant Center (XML)
```bash
./bin/skupilot export google --batch-id batch_default --output ./dist/google_feed.xml
```

#### Publish to Shopify Admin API
```bash
./bin/skupilot export shopify \
  --store "brand-store.myshopify.com" \
  --token "${SHOPIFY_ADMIN_TOKEN}" \
  --batch-id batch_default
```

---

## Configuration

Environment variables can be provided in a `.env` file or exported directly:

| Variable | Description | Default |
|---|---|---|
| `SKUPILOT_PORT` | HTTP server port | `8080` |
| `SKUPILOT_DB` | SQLite database file path | `./catalog.db` |
| `GEMINI_API_KEY` | Google Gemini API key for VLM | `""` (falls back to local heuristic) |
| `OPENAI_API_KEY` | OpenAI API key for VLM | `""` |
| `SHOPIFY_STORE_DOMAIN` | Shopify store domain | `""` |
| `SHOPIFY_ADMIN_TOKEN` | Shopify Admin API access token | `""` |

---

## Documentation

- [Architecture Design](docs/ARCHITECTURE.md): Deep dive into mathematical algorithms, data structures, and pipeline flows.
- [REST and SSE API Specification](docs/API_SPEC.md): Complete API documentation for all server endpoints.
- [Marketplace Compliance Rules](docs/MARKETPLACE_COMPLIANCE.md): Amazon, Shopify, and Google Shopping requirements and formulas.

---

## Testing

Run the full test suite with coverage:
```bash
go test -v -race -coverprofile=coverage.txt ./...
```

---

## License

MIT License. See [LICENSE](LICENSE) for details.
