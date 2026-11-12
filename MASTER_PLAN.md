# skupilot: Architecture and Engineering Master Plan
> Autonomous Multimodal E-Commerce Catalog Enrichment and Marketplace Compliance Engine
> Target: Generative AI Automation, Retail Tech and E-Commerce Product Engineering, Multimodal Attribute Extraction

---

## 1. Executive Summary and Market Context

### 1.1 Catalog Ingestion in Modern E-Commerce
High-volume e-commerce brands, distributors, and multi-vendor marketplaces (Shopify, Amazon, Zalando, Walmart) ingest thousands of new product SKUs weekly from supplier feeds and wholesaler spreadsheets:
- Missing Technical Attributes: Over 60% of supplier records omit search filter attributes (for example sleeve length, collar style, fabric composition, closure type, and exact color hex), preventing products from appearing in filtered marketplace searches.
- Strict Marketplace Image Rejections: Amazon requires pure white backgrounds (RGB 255, 255, 255) on main images, minimum 1000 x 1000 px resolution for zoom, and prohibits watermarks or promotional text. Non-compliant listings are suppressed or banned.
- Ineffective Conversion Copy: Supplier descriptions are typically unformatted technical bullets or keyword-stuffed strings that hurt search rankings and reduce buyer trust.
- Manual Labor Costs: Data entry teams and photo editors cost between $5 and $15 per SKU, taking days or weeks to format and publish new inventory.

### 1.2 The skupilot Solution
skupilot is a self-hosted single-binary Go daemon with an embedded React 19 merchandising studio. It automates catalog ingestion, attribute extraction, and copy generation end-to-end:
- Multimodal Visual Attribute Extraction: Uses Vision-Language Models (VLMs) and computer vision algorithms to inspect product photos and extract granular specifications (fabric texture, silhouette, necklines, hardware finish) directly from pixels.
- Automated Image Compliance Auditor: Verifies background color purity (RGB 255), resolution zoom readiness, aspect ratio, and watermark absence.
- Algorithmic A9 and Google SEO Copywriting: Generates marketplace-optimized titles, 5 benefit-driven bullet points, and clean HTML descriptions adhering to platform character limits.
- 1-Click Marketplace Export: Syncs enriched catalogs directly to Shopify Admin API, generates Amazon Inventory Loader flat files (TSV), and exports Google Merchant Center XML feeds.

---

## 2. High-Level System Architecture

```
                       +---------------------------------------------+
                       |           Raw Supplier Catalog Feed         |
                       |       (CSV, JSON, Shopify Export, TSV)      |
                       +---------------------------------------------+
                                              |
                                              v
                       +---------------------------------------------+
                       |               skupilot Engine               |
                       |        (Single Go Binary, Zero CGO)         |
                       +---------------------------------------------+
                                              |
                         +--------------------+--------------------+
                         |                                         |
                         v                                         v
         +-------------------------------+         +-------------------------------+
         |     Image Compliance Auditor  |         |   Multimodal VLM Extractor    |
         |  - RGB 255 Pure White Check   |         |  - Cut, Fit, and Silhouette   |
         |  - Resolution >= 1000x1000 px |         |  - Fabric and Material Detect |
         |  - Watermark and Overlay Math |         |  - Exact Hex and Color Family |
         +-------------------------------+         +-------------------------------+
                         |                                         |
                         +--------------------+--------------------+
                                              |
                                              v
                       +---------------------------------------------+
                       |          Enriched Attribute Record          |
                       |  - Category: Apparel > Men > Shirts         |
                       |  - Material: 100% Breathable Linen          |
                       |  - Collar: Mandarin Band Collar             |
                       |  - Color: #EAE6DF (Oatmeal Cream)           |
                       +---------------------------------------------+
                                              |
                                              v
                       +---------------------------------------------+
                       |        Algorithmic Copywriter Engine        |
                       |  - Formulates Amazon A9 / SEO Title         |
                       |  - Synthesizes 5 Benefit-Driven Bullets     |
                       |  - Enforces Strict Character Bounds         |
                       +---------------------------------------------+
                                              |
                                              v
                       +---------------------------------------------+
                       |         Marketplace Compliance Gate         |
                       |  - Amazon Style Guide Checks                |
                       |  - Google Shopping Taxonomy Mapping         |
                       |  - Quality Scorecard (0 - 100%)             |
                       +---------------------------------------------+
                                              |
                         +--------------------+--------------------+
                         |                                         |
                         v                                         v
         +-------------------------------+         +-------------------------------+
         |      1-Click Channel Sync     |         |    React 19 Merchandising     |
         |  - Shopify GraphQL Admin API  |         |  - Visual Grid with Status    |
         |  - Amazon Flat File (.tsv)    |         |  - Image Defect Annotations   |
         |  - Google Merchant Center XML |         |  - Live SSE Batch Streaming   |
         +-------------------------------+         +-------------------------------+
```

---

## 3. Mathematical Image Compliance Verification

For each main product image I with dimensions W x H:
1. Resolution and Zoom Gate:
   ZoomReady = (W >= 1000) and (H >= 1000)
2. Background Whiteness Metric:
   Sample corner and perimeter pixel regions P_corner (10px border along all 4 edges). For each pixel p = (R, G, B):
   Whiteness Score = (1 / |P_corner|) * SUM[I(R >= 250 and G >= 250 and B >= 250)]
   Amazon Compliance Requirement: Whiteness Score >= 0.985.
3. Aspect Ratio:
   Aspect Ratio = W / H in [0.8, 1.25] (Target: 1.0 Square).
4. Watermark and High-Contrast Edge Detection:
   Calculates edge density and contrast variance along the border regions to flag stamped logos, text overlays, or supplier watermarks.

---

## 4. SQLite Catalog Schema

```sql
-- Ingested Supplier Batches
CREATE TABLE IF NOT EXISTS catalog_batches (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    total_skus INTEGER NOT NULL,
    enriched_count INTEGER DEFAULT 0,
    compliant_count INTEGER DEFAULT 0,
    status TEXT NOT NULL,              -- 'PROCESSING', 'READY_FOR_REVIEW', 'PUBLISHED'
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Individual Product SKU Records
CREATE TABLE IF NOT EXISTS products (
    id TEXT PRIMARY KEY,
    batch_id TEXT NOT NULL REFERENCES catalog_batches(id) ON DELETE CASCADE,
    raw_sku TEXT NOT NULL,
    raw_title TEXT NOT NULL,
    raw_description TEXT,
    raw_category TEXT,
    raw_price REAL DEFAULT 0.0,
    currency TEXT DEFAULT 'USD',
    image_url TEXT NOT NULL,
    
    -- Visual Compliance Metrics
    image_width INTEGER DEFAULT 0,
    image_height INTEGER DEFAULT 0,
    image_zoom_ready BOOLEAN DEFAULT 0,
    image_white_bg_score REAL DEFAULT 0.0,
    image_aspect_ratio REAL DEFAULT 1.0,
    image_watermark_detected BOOLEAN DEFAULT 0,
    image_compliance_status TEXT DEFAULT 'PENDING', -- 'PASS', 'WARN', 'FAIL'
    
    -- Extracted Attributes (JSON)
    extracted_attributes_json TEXT,    -- {material: "linen", collar: "mandarin", color_hex: "#EAE6DF", fit: "slim"}
    
    -- Enriched Marketing Copy
    optimized_title TEXT,
    bullet_points_json TEXT,           -- ["5 benefit-driven bullets"]
    seo_description_html TEXT,
    google_taxonomy_id INTEGER DEFAULT 0,
    google_taxonomy_path TEXT,
    
    -- Overall Quality and Publication Status
    quality_score INTEGER DEFAULT 0,   -- 0 to 100
    publication_status TEXT NOT NULL,  -- 'PENDING', 'APPROVED', 'REJECTED', 'SYNCED_SHOPIFY', 'SYNCED_AMAZON'
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Channel Export Audit Logs
CREATE TABLE IF NOT EXISTS export_logs (
    id TEXT PRIMARY KEY,
    batch_id TEXT NOT NULL REFERENCES catalog_batches(id) ON DELETE CASCADE,
    channel TEXT NOT NULL,             -- 'SHOPIFY', 'AMAZON', 'GOOGLE'
    item_count INTEGER NOT NULL,
    status TEXT NOT NULL,              -- 'SUCCESS', 'FAILED'
    details TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

---

## 5. Repository and Directory Structure

```
skupilot/
|-- .github/
|   `-- workflows/
|       |-- test.yml
|       `-- release.yml
|-- cmd/
|   `-- skupilot/
|       `-- main.go                 # CLI entrypoint (enrich, audit, export, serve)
|-- docs/
|   |-- ARCHITECTURE.md             # Detailed engineering design
|   |-- API_SPEC.md                 # REST and SSE API specification
|   `-- MARKETPLACE_COMPLIANCE.md   # Platform style guidelines and rules
|-- internal/
|   |-- config/                     # Configuration and channel credentials
|   |-- domain/                     # Core domain entities and data structures
|   |-- feed/                       # Raw Catalog Ingestion
|   |   |-- csv_reader.go           # High-throughput CSV streamer
|   |   |-- json_reader.go          # JSON catalog parser
|   |   `-- shopify_importer.go     # Shopify products import parser
|   |-- vision/                     # Multimodal Vision and Image Processing
|   |   |-- compliance.go           # Whiteness, dimensions, and watermark math
|   |   |-- vlm.go                  # Multimodal VLM client (Gemini / OpenAI / Heuristic)
|   |   `-- prompts.go              # Structured extraction prompts
|   |-- copywriter/                 # Algorithmic Marketplace Copy Generator
|   |   |-- title.go                # Amazon A9 formulaic title synthesizer
|   |   |-- bullets.go              # 5 benefit-driven feature bullets
|   |   |-- description.go          # HTML SEO product description
|   |   |-- taxonomy.go             # Google Shopping category ID mapper
|   |   `-- scorecard.go            # 0 to 100 listing quality metric
|   |-- channels/                   # E-Commerce Channel Exporters
|   |   |-- shopify.go              # Shopify GraphQL Admin API exporter
|   |   |-- amazon.go               # Amazon Inventory Loader TSV generator
|   |   `-- google.go               # Google Merchant Center XML feed generator
|   |-- store/                      # SQLite Persistence Layer (Zero CGO)
|   |   |-- db.go
|   |   |-- products.go
|   |   `-- batches.go
|   `-- api/                        # HTTP REST and SSE Server
|       |-- server.go
|       |-- handlers.go
|       `-- sse.go
|-- web/                            # Embedded React 19 Merchandising Studio
|   |-- index.html
|   |-- package.json
|   |-- vite.config.ts
|   `-- src/
|       |-- App.tsx
|       |-- components/
|       |   |-- ProductGrid.tsx     # Merchandising catalog view
|       |   |-- ComplianceBadge.tsx # Visual indicators for Amazon / Google
|       |   |-- ProductDetail.tsx   # Side-by-side attribute and copy editor
|       |   |-- ChannelSync.tsx     # 1-Click channel exporter modal
|       |   |-- BatchUpload.tsx     # Upload CSV/JSON with live SSE progress
|       |   `-- QualityRadar.tsx    # Scorecard breakdown component
|       `-- lib/
|           `-- api.ts              # Frontend client for Go API
|-- examples/
|   |-- sample_apparel_feed.csv
|   |-- sample_footwear_feed.json
|   `-- sample_accessories_feed.csv
|-- go.mod
|-- go.sum
|-- Makefile
|-- LICENSE
`-- README.md
```

---

## 6. CLI Command Surface

```bash
# Start the web merchandising studio
skupilot serve --port 8080 --db ./catalog.db

# Ingest and enrich a raw supplier feed CSV
skupilot enrich --input ./supplier_feed.csv --category apparel --model gemini-2.5-flash

# Audit a catalog for marketplace image and title compliance
skupilot audit --db ./catalog.db --strict-amazon

# Export enriched products to Shopify via GraphQL Admin API
skupilot export shopify \
  --store "mystore.myshopify.com" \
  --token "${SHOPIFY_ADMIN_TOKEN}" \
  --batch-id batch_71fa82

# Export to Amazon Inventory Loader Flat File (TSV)
skupilot export amazon --batch-id batch_71fa82 --output ./amazon_feed.tsv

# Export to Google Merchant Center XML Feed
skupilot export google --batch-id batch_71fa82 --output ./google_feed.xml
```

---

## 7. Implementation Roadmap (5 Phases)

### Phase 1: Feed Parsers and Image Compliance Engine
- Implement high-speed CSV/JSON/TSV product feed streamer.
- Build image analysis engine calculating corner background whiteness, resolution, and watermark heuristics.
- Flag non-compliant images and output compliance diagnostics.

### Phase 2: Multimodal VLM Attribute Extractor
- Build visual inspection system extracting granular technical attributes per category.
- Map extracted attributes to normalized taxonomy enums (Google Product Taxonomy).
- Support dual-mode execution: Cloud VLM (Gemini / OpenAI) with an offline computer vision heuristic fallback engine.

### Phase 3: Algorithmic Copywriter and Listing Scorecard
- Implement Amazon A9 formulaic title generator respecting marketplace character limits (150 to 200 characters).
- Synthesize 5 customer-facing feature bullets with benefit hooks.
- Compute listing Quality Scorecard (0 to 100%) across visual compliance, attribute density, and copywriting quality.

### Phase 4: Channel Integrations (Shopify, Amazon, Google)
- Implement Shopify GraphQL Admin API connector creating draft or active products with metafields.
- Implement Amazon Flat File TSV generator.
- Implement Google Merchant Center XML feed generator.

### Phase 5: Embedded React 19 Merchandising Studio and REST/SSE Server
- Build high-density visual product grid with zoomable images and compliance badges.
- Add inline editable attributes and instant re-generation.
- Implement SSE server for real-time batch processing feedback.
- Package frontend into single Go binary using Go embed.
