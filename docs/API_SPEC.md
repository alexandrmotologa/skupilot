# skupilot REST and SSE API Specification

skupilot exposes an HTTP REST API alongside Server-Sent Events (SSE) for streaming batch progress.

Base URL: `http://localhost:8080/api`

---

## Endpoints

### 1. System and Health

#### `GET /api/health`
Returns the status of the server and database.
Response:
```json
{
  "status": "ok",
  "version": "1.0.0",
  "database": "connected",
  "uptime_seconds": 124
}
```

#### `GET /api/stats`
Returns aggregated catalog statistics.
Response:
```json
{
  "total_batches": 4,
  "total_products": 250,
  "compliant_products": 238,
  "pending_review": 12,
  "average_quality_score": 94.2
}
```

---

## 2. Batches and Catalog Ingestion

#### `GET /api/batches`
Returns all ingested supplier batches.
Response:
```json
[
  {
    "id": "batch_71fa82",
    "name": "Spring 2026 Apparel Supplier Feed",
    "total_skus": 120,
    "enriched_count": 120,
    "compliant_count": 114,
    "status": "READY_FOR_REVIEW",
    "created_at": "2026-11-12T10:00:00Z"
  }
]
```

#### `POST /api/batches/upload`
Uploads a raw CSV, TSV, or JSON supplier feed and begins background ingestion.
Request: `multipart/form-data`
- `file`: The feed file.
- `name`: Human-readable batch name.
- `category_hint`: Optional category hint (e.g. `apparel`, `footwear`).

Response:
```json
{
  "batch_id": "batch_71fa82",
  "total_skus": 120,
  "status": "PROCESSING",
  "message": "Batch uploaded and queued for processing"
}
```

#### `GET /api/batches/:id/events` (Server-Sent Events)
Streams real-time progress for batch enrichment and image compliance checking.
Event format:
```
event: progress
data: {"batch_id":"batch_71fa82","current":42,"total":120,"percent":35,"current_sku":"SKU-8921"}

event: complete
data: {"batch_id":"batch_71fa82","enriched_count":120,"compliant_count":114}
```

---

## 3. Products

#### `GET /api/batches/:id/products`
Retrieves products for a specific batch with optional filtering.
Query parameters:
- `status`: Filter by publication status (`PENDING`, `APPROVED`, `REJECTED`, `SYNCED_SHOPIFY`).
- `compliance`: Filter by image compliance (`PASS`, `WARN`, `FAIL`).
- `search`: Search raw title or SKU.
- `limit`: Number of items (default: 50).
- `offset`: Pagination offset (default: 0).

Response:
```json
{
  "total": 120,
  "limit": 50,
  "offset": 0,
  "items": [
    {
      "id": "prod_01",
      "batch_id": "batch_71fa82",
      "raw_sku": "LIN-SH-001",
      "raw_title": "Mens Linen Shirt White Casual",
      "image_url": "https://example.com/images/shirt_001.jpg",
      "image_zoom_ready": true,
      "image_white_bg_score": 0.992,
      "image_watermark_detected": false,
      "image_compliance_status": "PASS",
      "extracted_attributes": {
        "material": "100% Breathable Linen",
        "collar": "Mandarin Band Collar",
        "fit": "Tailored Slim Fit",
        "color_hex": "#FFFFFF",
        "color_family": "Pure White"
      },
      "optimized_title": "Nordic Weave Men's 100% Linen Band Collar Casual Summer Shirt - Crisp White",
      "bullet_points": [
        "100% Breathable Linen: Natural woven fibers keep you cool in hot climates.",
        "Modern Band Collar: Clean mandarin neckline without standard fold-down bulk.",
        "Tailored Slim Silhouette: Tapered waist provides clean drape without restricting movement.",
        "Durable Shell Buttons: Reinforced cross-stitched button placket withstands repeated laundering.",
        "Versatile Wardrobe Essential: Pairs seamlessly with tailored trousers or casual linen shorts."
      ],
      "seo_description_html": "<p>Crafted from 100% natural breathable linen...</p>",
      "google_taxonomy_id": 212,
      "google_taxonomy_path": "Apparel & Accessories > Clothing > Shirts & Tops",
      "quality_score": 98,
      "publication_status": "APPROVED"
    }
  ]
}
```

#### `GET /api/products/:id`
Retrieves a single product with full audit details and history.

#### `PATCH /api/products/:id`
Updates product attributes, copy, or approval status.
Request body:
```json
{
  "optimized_title": "Custom Updated Title",
  "publication_status": "APPROVED",
  "extracted_attributes": {
    "material": "95% Cotton, 5% Linen"
  }
}
```

---

## 4. Channel Export Endpoints

#### `POST /api/batches/:id/export/shopify`
Generates and publishes products to Shopify Admin API.
Request body:
```json
{
  "store_domain": "my-brand.myshopify.com",
  "access_token": "shpat_xxxxxxxxxxxx",
  "publish_status": "DRAFT",
  "only_approved": true
}
```
Response:
```json
{
  "batch_id": "batch_71fa82",
  "exported_count": 114,
  "failed_count": 0,
  "channel": "SHOPIFY",
  "status": "SUCCESS"
}
```

#### `GET /api/batches/:id/export/amazon`
Downloads the Amazon Inventory Loader Flat File (.tsv).
Content-Type: `text/tab-separated-values`

#### `GET /api/batches/:id/export/google`
Downloads the Google Merchant Center Product Feed (.xml).
Content-Type: `application/xml`
