export type BatchStatus = 'PROCESSING' | 'READY_FOR_REVIEW' | 'PUBLISHED' | 'FAILED';

export type PublicationStatus = 'PENDING' | 'APPROVED' | 'REJECTED' | 'SYNCED_SHOPIFY' | 'SYNCED_AMAZON';

export type ComplianceStatus = 'PASS' | 'WARN' | 'FAIL';

export interface CatalogBatch {
  id: string;
  name: string;
  total_skus: number;
  enriched_count: number;
  compliant_count: number;
  status: BatchStatus;
  created_at: string;
  updated_at: string;
}

export interface ExtractedAttributes {
  material?: string;
  fit?: string;
  collar?: string;
  neckline?: string;
  sleeve?: string;
  closure?: string;
  pattern?: string;
  color_hex?: string;
  color_family?: string;
  care?: string;
  features?: string[];
}

export interface Product {
  id: string;
  batch_id: string;
  raw_sku: string;
  raw_title: string;
  raw_description?: string;
  raw_category?: string;
  raw_price: number;
  currency: string;
  brand?: string;
  image_url: string;

  // Visual Compliance Metrics
  image_width: number;
  image_height: number;
  image_zoom_ready: boolean;
  image_white_bg_score: number;
  image_aspect_ratio: number;
  image_watermark_detected: boolean;
  image_compliance_status: ComplianceStatus;

  // Extracted Attributes
  extracted_attributes: ExtractedAttributes;

  // Enriched Marketing Copy
  optimized_title: string;
  bullet_points: string[];
  seo_description_html: string;
  google_taxonomy_id: number;
  google_taxonomy_path: string;

  // Quality & Status
  quality_score: number;
  publication_status: PublicationStatus;
  created_at: string;
  updated_at: string;
}

export interface ProductListResponse {
  total: number;
  limit: number;
  offset: number;
  items: Product[];
}

export interface BatchProgressEvent {
  batch_id: string;
  current: number;
  total: number;
  percent: number;
  current_sku: string;
  product?: Product;
}

export interface BatchAnalytics {
  batch_id: string;
  total_products: number;
  compliant_count: number;
  warn_count: number;
  fail_count: number;
  compliance_rate: number;
  average_quality_score: number;
  non_white_bg_count: number;
  low_resolution_count: number;
  watermark_count: number;
  missing_material_count: number;
  missing_fit_count: number;
  pending_count: number;
  approved_count: number;
  rejected_count: number;
  amazon_ready_count: number;
  shopify_ready_count: number;
  google_ready_count: number;
}

export interface StyleVariant {
  style: string;
  title: string;
  bullets: string[];
  description_html: string;
}
