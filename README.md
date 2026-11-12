<p align="center">
  <img src="docs/images/logo.png?raw=true" alt="skupilot Logo" width="130" style="border-radius: 24px;" />
</p>

<h1 align="center">skupilot</h1>

<p align="center">
  <strong>Autonomous Multimodal E-Commerce Catalog Enrichment and Marketplace Compliance Engine</strong>
</p>

<p align="center">
  <a href="#the-problem">Problem</a> •
  <a href="#studio-tour">Studio Tour</a> •
  <a href="#what-skupilot-does">Features</a> •
  <a href="#architecture-overview">Architecture</a> •
  <a href="#quick-start">Quick Start</a> •
  <a href="#documentation">Docs</a>
</p>

---

skupilot is a self-hosted single-binary Go service with an embedded React 19 merchandising studio. It automates catalog ingestion, multimodal image auditing, attribute extraction, and marketplace copywriting for high-volume e-commerce brands, distributors, and multi-vendor marketplaces.

---

## Studio Tour

### Live Merchandising Studio & Interactive Flow
skupilot features a self-contained web studio compiled directly into the Go executable via Go embed. It provides instant visual feedback, batch progress streaming via Server-Sent Events (SSE), and deep inspection modals.

<p align="center">
  <img src="docs/images/skupilot_demo.gif?raw=true" alt="skupilot Live Demo Flow" width="95%" style="border-radius: 8px; box-shadow: 0 4px 20px rgba(0,0,0,0.15);" />
</p>

### Merchandising Studio Dashboard
The desktop workspace provides real-time health metrics, listing pass rates, quality score distributions, and defect-type quick filters across thousands of SKUs.

<p align="center">
  <img src="docs/images/skupilot_dashboard.png?raw=true" alt="skupilot Desktop Merchandising Studio" width="95%" style="border-radius: 8px;" />
</p>

### Multimodal Inspection & Algorithmic Copywriting Modal
Audit individual products with pixel-level RGB 255 compliance measurements, VLM-extracted attributes (materials, cut, neckline, closure), Amazon A9 formulaic titles, and 5 benefit-driven feature bullets with live character counters and style switching.

<p align="center">
  <img src="docs/images/skupilot_audit_modal.png?raw=true" alt="skupilot Product Inspection Modal" width="95%" style="border-radius: 8px;" />
</p>

### 1-Click Multi-Channel Sync
Export compliant feeds directly to Shopify Admin API (GraphQL mutations with metafields), Amazon Inventory Loader (.tsv flat files), Google Merchant Center RSS 2.0 (XML), or Master PIM CSV.

<p align="center">
  <img src="docs/images/skupilot_channel_export.png?raw=true" alt="skupilot Channel Export Modal" width="95%" style="border-radius: 8px;" />
</p>

### Responsive Mobile Interface
The React 19 studio adapts dynamically to any viewport width, offering mobile inventory management on phones and tablets.

<p align="center">
  <img src="docs/images/skupilot_mobile.png?raw=true" alt="skupilot Mobile Studio Viewport" width="380" style="border-radius: 8px;" />
</p>

---

## The Problem

High-volume catalog ingestion faces recurring operational bottlenecks:
- Missing Technical Attributes: Over 60% of supplier feeds omit search filter attributes such as collar styles, exact materials, silhouette cuts, and color codes.
- Strict Image Standards: Marketplaces like Amazon automatically suppress listings whose main image lacks a pure white RGB 255 background, fails the 1000 x 1000 px zoom threshold, or contains watermarks.
- Subpar Conversion Copy: Raw supplier titles and descriptions are often unformatted, keyword-stuffed strings that reduce customer trust and search rank.
- Manual Labor Costs: Formatting new inventory manually costs between $5 and $15 per SKU and introduces delays.

---

## What skupilot Does

1. High-Throughput Catalog Ingestion: Streams large CSV, TSV, JSON, and Shopify product dumps without buffering entire datasets into memory.
2. Mathematical Image Compliance Auditor:
   - Measures corner and perimeter background whiteness using a strict RGB (255, 255, 255) purity threshold.
   - Verifies resolution zoom readiness (minimum 1000 x 1000 pixels).
   - Detects edge density and high-contrast watermarks or text overlays.
   - Validates aspect ratios against marketplace square standards (0.8 to 1.25).
3. Automated Image Remediation:
   - Whitens non-compliant perimeter pixels to pure RGB 255 with smooth edge alpha transitions.
   - Centers and pads images onto a 1000 x 1000 px square canvas with zero distortion.
4. Multimodal VLM Attribute Extraction:
   - Uses Vision-Language Models (Gemini 1.5/2.5 Flash, OpenAI GPT-4o-mini) and an offline computer vision heuristic engine to inspect product photos.
   - Extracts fabric composition, garment cut, collar type, neckline, closure mechanism, and exact hex color codes directly from pixels.
5. Algorithmic Copywriting & Quality Scoring:
   - Formulates Amazon A9 compliant titles between 150 and 200 characters, filtering prohibited promotional claims.
   - Synthesizes 5 structured, benefit-driven feature bullets.
   - Generates clean semantic HTML product descriptions.
   - Resolves Google Product Taxonomy numeric IDs and breadcrumb paths.
   - Computes a comprehensive Listing Quality Scorecard (0 to 100).
6. 1-Click Multi-Channel Sync:
   - Publishes directly to the Shopify Admin API using GraphQL mutations and custom metafields.
   - Generates Amazon Inventory Loader flat files (.tsv).
   - Exports Google Merchant Center product XML feeds.
   - Outputs universal Master PIM CSV files with 31 enriched columns.
7. Embedded React 19 Studio:
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
        |  - Auto-Whiten & 1:1 Padding  |   +-------------------------------+
        +-------------------------------+                   |
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
                      |  - Multi-Style Switcher (A9, D2C, GG) |
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
        |  - Master PIM CSV Export      |   |  - Interactive Analytics      |
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

Run the full test suite with race detection:
```bash
go test -v -race -coverprofile=coverage.txt ./...
```

---

## License

MIT License. See [LICENSE](LICENSE) for details.
