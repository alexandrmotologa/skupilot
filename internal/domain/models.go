package domain

import (
	"encoding/json"
	"time"
)

// BatchStatus represents the lifecycle state of an ingested feed batch.
type BatchStatus string

const (
	BatchStatusProcessing     BatchStatus = "PROCESSING"
	BatchStatusReadyForReview BatchStatus = "READY_FOR_REVIEW"
	BatchStatusPublished      BatchStatus = "PUBLISHED"
	BatchStatusFailed         BatchStatus = "FAILED"
)

// PublicationStatus represents individual product SKU publication state.
type PublicationStatus string

const (
	StatusPending       PublicationStatus = "PENDING"
	StatusApproved      PublicationStatus = "APPROVED"
	StatusRejected      PublicationStatus = "REJECTED"
	StatusSyncedShopify PublicationStatus = "SYNCED_SHOPIFY"
	StatusSyncedAmazon  PublicationStatus = "SYNCED_AMAZON"
	StatusSyncedGoogle  PublicationStatus = "SYNCED_GOOGLE"
)

// ComplianceStatus represents Amazon and marketplace image compliance outcome.
type ComplianceStatus string

const (
	CompliancePass ComplianceStatus = "PASS"
	ComplianceWarn ComplianceStatus = "WARN"
	ComplianceFail ComplianceStatus = "FAIL"
)

// CatalogBatch models a batch of products ingested from a raw supplier feed.
type CatalogBatch struct {
	ID             string      `json:"id"`
	Name           string      `json:"name"`
	TotalSKUs      int         `json:"total_skus"`
	EnrichedCount  int         `json:"enriched_count"`
	CompliantCount int         `json:"compliant_count"`
	Status         BatchStatus `json:"status"`
	CreatedAt      time.Time   `json:"created_at"`
	UpdatedAt      time.Time   `json:"updated_at"`
}

// ExtractedAttributes stores structured product attributes discovered from image and text.
type ExtractedAttributes struct {
	Material    string   `json:"material,omitempty"`
	Fit         string   `json:"fit,omitempty"`
	Collar      string   `json:"collar,omitempty"`
	Neckline    string   `json:"neckline,omitempty"`
	Sleeve      string   `json:"sleeve,omitempty"`
	Closure     string   `json:"closure,omitempty"`
	Pattern     string   `json:"pattern,omitempty"`
	ColorHex    string   `json:"color_hex,omitempty"`
	ColorFamily string   `json:"color_family,omitempty"`
	Care        string   `json:"care,omitempty"`
	Features    []string `json:"features,omitempty"`
}

// Product represents a single SKU with raw inputs, image compliance metrics, and enriched copy.
type Product struct {
	ID             string `json:"id"`
	BatchID        string `json:"batch_id"`
	RawSKU         string `json:"raw_sku"`
	RawTitle       string `json:"raw_title"`
	RawDescription string `json:"raw_description,omitempty"`
	RawCategory    string `json:"raw_category,omitempty"`
	RawPrice       float64 `json:"raw_price"`
	Currency       string `json:"currency"`
	Brand          string `json:"brand,omitempty"`
	ImageURL       string `json:"image_url"`

	// Visual Compliance Metrics
	ImageWidth             int              `json:"image_width"`
	ImageHeight            int              `json:"image_height"`
	ImageZoomReady         bool             `json:"image_zoom_ready"`
	ImageWhiteBGScore      float64          `json:"image_white_bg_score"`
	ImageAspectRatio       float64          `json:"image_aspect_ratio"`
	ImageWatermarkDetected bool             `json:"image_watermark_detected"`
	ImageComplianceStatus  ComplianceStatus `json:"image_compliance_status"`

	// Extracted Attributes
	ExtractedAttributes ExtractedAttributes `json:"extracted_attributes"`

	// Enriched Marketing Copy
	OptimizedTitle     string   `json:"optimized_title"`
	BulletPoints       []string `json:"bullet_points"`
	SEODescriptionHTML string   `json:"seo_description_html"`
	GoogleTaxonomyID   int      `json:"google_taxonomy_id"`
	GoogleTaxonomyPath string   `json:"google_taxonomy_path"`

	// Quality and Workflow Status
	QualityScore      int               `json:"quality_score"`
	PublicationStatus PublicationStatus `json:"publication_status"`
	CreatedAt         time.Time         `json:"created_at"`
	UpdatedAt         time.Time         `json:"updated_at"`
}

// ComplianceReport holds detailed diagnostics for image compliance checks.
type ComplianceReport struct {
	Width             int              `json:"width"`
	Height            int              `json:"height"`
	AspectRatio       float64          `json:"aspect_ratio"`
	ZoomReady         bool             `json:"zoom_ready"`
	WhiteBGScore      float64          `json:"white_bg_score"`
	WatermarkDetected bool             `json:"watermark_detected"`
	Status            ComplianceStatus `json:"status"`
	Diagnostics       []string         `json:"diagnostics"`
}

// QualityScorecard decomposes the 0-100 listing quality metric.
type QualityScorecard struct {
	VisualScore     int      `json:"visual_score"`     // 0 - 30
	AttributeScore  int      `json:"attribute_score"`  // 0 - 30
	TitleScore      int      `json:"title_score"`      // 0 - 20
	BulletScore     int      `json:"bullet_score"`     // 0 - 20
	TotalScore      int      `json:"total_score"`      // 0 - 100
	Recommendations []string `json:"recommendations"`
}

// RawFeedItem represents an incoming record from CSV or JSON.
type RawFeedItem struct {
	SKU         string  `json:"sku"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Category    string  `json:"category"`
	Price       float64 `json:"price"`
	Currency    string  `json:"currency"`
	Brand       string  `json:"brand"`
	ImageURL    string  `json:"image_url"`
}

// ShopifyExportRequest parameters for publishing products to Shopify.
type ShopifyExportRequest struct {
	StoreDomain   string `json:"store_domain"`
	AccessToken   string `json:"access_token"`
	BatchID       string `json:"batch_id"`
	PublishStatus string `json:"publish_status"` // "DRAFT" or "ACTIVE"
	OnlyApproved  bool   `json:"only_approved"`
}

// ExportResult contains summary metrics after running an export.
type ExportResult struct {
	BatchID       string `json:"batch_id"`
	Channel       string `json:"channel"`
	ExportedCount int    `json:"exported_count"`
	FailedCount   int    `json:"failed_count"`
	Status        string `json:"status"`
	Details       string `json:"details,omitempty"`
}

// BatchAnalytics provides aggregated compliance and quality diagnostics for a batch.
type BatchAnalytics struct {
	BatchID             string  `json:"batch_id"`
	TotalProducts       int     `json:"total_products"`
	CompliantCount      int     `json:"compliant_count"`
	WarnCount           int     `json:"warn_count"`
	FailCount           int     `json:"fail_count"`
	ComplianceRate      float64 `json:"compliance_rate"`
	AverageQualityScore float64 `json:"average_quality_score"`

	// Issues Breakdown
	NonWhiteBGCount     int `json:"non_white_bg_count"`
	LowResolutionCount  int `json:"low_resolution_count"`
	WatermarkCount      int `json:"watermark_count"`
	MissingMaterialCount int `json:"missing_material_count"`
	MissingFitCount     int `json:"missing_fit_count"`

	// Status Distribution
	PendingCount  int `json:"pending_count"`
	ApprovedCount int `json:"approved_count"`
	RejectedCount int `json:"rejected_count"`

	// Channel Readiness
	AmazonReadyCount  int `json:"amazon_ready_count"`
	ShopifyReadyCount int `json:"shopify_ready_count"`
	GoogleReadyCount  int `json:"google_ready_count"`
}

// MarshalAttributesHelper serializes attributes to JSON string for SQLite storage.
func MarshalAttributes(attr ExtractedAttributes) (string, error) {
	b, err := json.Marshal(attr)
	if err != nil {
		return "{}", err
	}
	return string(b), nil
}

// UnmarshalAttributesHelper deserializes attributes from JSON string.
func UnmarshalAttributes(data string) (ExtractedAttributes, error) {
	var attr ExtractedAttributes
	if data == "" {
		return attr, nil
	}
	err := json.Unmarshal([]byte(data), &attr)
	return attr, err
}

// MarshalBulletsHelper serializes bullets to JSON string.
func MarshalBullets(bullets []string) (string, error) {
	b, err := json.Marshal(bullets)
	if err != nil {
		return "[]", err
	}
	return string(b), nil
}

// UnmarshalBulletsHelper deserializes bullets from JSON string.
func UnmarshalBullets(data string) ([]string, error) {
	var bullets []string
	if data == "" {
		return bullets, nil
	}
	err := json.Unmarshal([]byte(data), &bullets)
	return bullets, err
}
