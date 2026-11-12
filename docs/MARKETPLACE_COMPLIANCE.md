# Marketplace Compliance and Feed Standards

skupilot enforces compliance rules derived from Amazon Seller Central style guides, Google Merchant Center product data specifications, and Shopify GraphQL Admin API conventions.

---

## 1. Amazon Marketplace Compliance

### Image Requirements (Main Listing Image)
Amazon listings that violate main image standards are automatically search-suppressed.
- Pure White Background: The background of the main image must be pure white, evaluated as RGB (255, 255, 255) with a minimum 98.5% compliance threshold on boundary sample points.
- Zoom Capability: Images must have minimum dimensions of 1000 pixels along the longest side (1600 x 1600 recommended).
- Product Fill: The product must fill 85% or more of the image frame.
- Prohibited Overlays: No watermarks, borders, text, promotional stickers, or inset images.

### Title Guidelines (A9 Algorithm Formula)
- Structure:
  `[Brand] + [Target Audience] + [Material] + [Product Type] + [Key Style / Feature] + [Color]`
- Length Bounds:
  - Minimum: 80 characters.
  - Recommended: 150 to 180 characters.
  - Maximum Hard Cap: 200 characters (titles over 200 characters risk suppression).
- Prohibited Phrases:
  - Promotional terms: "Best Seller", "Free Shipping", "Sale", "Discount", "Top Rated".
  - Subjective claims: "Perfect Quality", "Number One", "Guaranteed".
  - Punctuation abuse: All-caps words, repeated exclamation marks, emoji characters.

### 5 Feature Bullet Points
- Exactly five bullets summarizing key product attributes.
- Each bullet starts with a capitalized feature headline followed by a colon and the benefit description.
- Recommended bullet length: 150 to 250 characters each.
- Total character count for all five bullets should stay under 1,000 characters.

---

## 2. Google Shopping and Merchant Center XML

### Google Product Taxonomy
Google requires a standardized taxonomy identifier (`g:google_product_category`). Numeric IDs and hierarchical paths map directly to standard Google definitions:
- Apparel > Clothing > Shirts & Tops: ID `212`
- Apparel > Clothing > Pants: ID `204`
- Apparel > Clothing > Dresses: ID `227`
- Apparel > Shoes: ID `187`
- Apparel > Handbags, Wallets & Cases: ID `6552`

### Required XML Elements
- `<g:id>`: Unique SKU or inventory identifier.
- `<g:title>`: Clean product title under 150 characters.
- `<g:description>`: Semantic description without keyword stuffing.
- `<g:link>`: Canonical product landing page URL.
- `<g:image_link>`: High-resolution image URL.
- `<g:availability>`: `in_stock`, `out_of_stock`, or `preorder`.
- `<g:price>`: Value and ISO 4217 currency code (e.g., `89.00 USD`).
- `<g:brand>`: Verified brand name.
- `<g:color>` and `<g:material>`: Extracted visual attributes.

---

## 3. Shopify GraphQL Admin API

### Product Creation Payload
Shopify GraphQL endpoint: `https://{shop}.myshopify.com/admin/api/2024-10/graphql.json`

Using the `productCreate` mutation:
```graphql
mutation productCreate($input: ProductInput!) {
  productCreate(input: $input) {
    product {
      id
      title
      handle
      status
    }
    userErrors {
      field
      message
    }
  }
}
```

### Metafields for Technical Attributes
Extracted visual attributes are stored as standard Shopify Metafields under the `custom` namespace:
- `custom.fabric_material`: String (e.g. "100% Breathable Linen")
- `custom.collar_style`: String (e.g. "Mandarin Band Collar")
- `custom.fit_silhouette`: String (e.g. "Tailored Slim Fit")
- `custom.hex_color`: Single-line text (e.g. "#EAE6DF")
- `custom.compliance_score`: Integer (0 to 100)
