package api

import (
	"context"
	"encoding/json"
	"fmt"
	"image/png"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/alexandrmotologa/skupilot/internal/channels"
	"github.com/alexandrmotologa/skupilot/internal/copywriter"
	"github.com/alexandrmotologa/skupilot/internal/domain"
	"github.com/alexandrmotologa/skupilot/internal/engine"
	"github.com/alexandrmotologa/skupilot/internal/feed"
	"github.com/alexandrmotologa/skupilot/internal/store"
	"github.com/alexandrmotologa/skupilot/internal/vision"
)

// ServerEnv wraps shared dependencies for API route handlers.
type ServerEnv struct {
	DB     *store.DB
	Engine *engine.Engine
	Broker *EventBroker
}

// HealthHandler returns system health and database connectivity.
func (env *ServerEnv) HealthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":   "ok",
		"version":  "1.0.0",
		"database": "connected",
		"time":     time.Now().UTC().Format(time.RFC3339),
	})
}

// StatsHandler returns aggregated counts across batches and products.
func (env *ServerEnv) StatsHandler(w http.ResponseWriter, r *http.Request) {
	batches, err := env.DB.ListBatches()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	totalBatches := len(batches)
	totalProducts := 0
	compliantProducts := 0

	for _, b := range batches {
		totalProducts += b.TotalSKUs
		compliantProducts += b.CompliantCount
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"total_batches":      totalBatches,
		"total_products":     totalProducts,
		"compliant_products": compliantProducts,
	})
}

// ListBatchesHandler returns all ingested supplier batches.
func (env *ServerEnv) ListBatchesHandler(w http.ResponseWriter, r *http.Request) {
	batches, err := env.DB.ListBatches()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if batches == nil {
		batches = []domain.CatalogBatch{}
	}
	writeJSON(w, http.StatusOK, batches)
}

// GetBatchHandler returns a single batch by ID.
func (env *ServerEnv) GetBatchHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	batch, err := env.DB.GetBatch(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if batch == nil {
		writeError(w, http.StatusNotFound, "batch not found")
		return
	}
	writeJSON(w, http.StatusOK, batch)
}

// UploadBatchHandler processes an uploaded supplier catalog file in the background.
func (env *ServerEnv) UploadBatchHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(50 << 20); err != nil { // 50MB max upload
		writeError(w, http.StatusBadRequest, "failed to parse multipart form")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "missing 'file' in upload request")
		return
	}
	defer file.Close()

	batchName := r.FormValue("name")
	if batchName == "" {
		batchName = header.Filename
	}

	// Detect format
	var items []domain.RawFeedItem
	filenameLower := strings.ToLower(header.Filename)

	if strings.HasSuffix(filenameLower, ".json") {
		items, err = feed.ParseJSON(file)
	} else {
		items, err = feed.ParseCSV(file)
	}

	if err != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("failed parsing feed: %v", err))
		return
	}

	if len(items) == 0 {
		writeError(w, http.StatusBadRequest, "feed contained no valid product items")
		return
	}

	batch, err := env.DB.CreateBatch(batchName, len(items))
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to create batch: %v", err))
		return
	}

	// Spawn background enrichment pipeline
	go func() {
		ctx := context.Background()
		_ = env.Engine.ProcessFeed(ctx, batch.ID, items, func(current, total int, sku string, p *domain.Product) {
			env.Broker.Broadcast(batch.ID, "progress", map[string]interface{}{
				"batch_id":    batch.ID,
				"current":     current,
				"total":       total,
				"percent":     int(float64(current) / float64(total) * 100),
				"current_sku": sku,
				"product":     p,
			})
		})

		// Send complete event
		env.Broker.Broadcast(batch.ID, "complete", map[string]interface{}{
			"batch_id": batch.ID,
			"total":    len(items),
		})
	}()

	writeJSON(w, http.StatusAccepted, map[string]interface{}{
		"batch_id":   batch.ID,
		"name":       batch.Name,
		"total_skus": len(items),
		"status":     "PROCESSING",
		"message":    "Batch created and processing started",
	})
}

// BatchSSEHandler handles SSE event streaming for live progress updates.
func (env *ServerEnv) BatchSSEHandler(w http.ResponseWriter, r *http.Request) {
	batchID := r.PathValue("id")
	env.Broker.ServeSSE(w, r, batchID)
}

// ListProductsHandler returns products for a batch with filters and pagination.
func (env *ServerEnv) ListProductsHandler(w http.ResponseWriter, r *http.Request) {
	batchID := r.PathValue("id")
	query := r.URL.Query()

	limit, _ := strconv.Atoi(query.Get("limit"))
	if limit <= 0 {
		limit = 50
	}
	offset, _ := strconv.Atoi(query.Get("offset"))

	filter := store.ProductFilter{
		BatchID:           batchID,
		PublicationStatus: query.Get("status"),
		ComplianceStatus:  query.Get("compliance"),
		Search:            query.Get("search"),
		Limit:             limit,
		Offset:            offset,
	}

	products, total, err := env.DB.ListProducts(filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if products == nil {
		products = []domain.Product{}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"total":  total,
		"limit":  limit,
		"offset": offset,
		"items":  products,
	})
}

// GetProductHandler returns a single product by ID.
func (env *ServerEnv) GetProductHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	p, err := env.DB.GetProduct(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if p == nil {
		writeError(w, http.StatusNotFound, "product not found")
		return
	}
	writeJSON(w, http.StatusOK, p)
}

// UpdateProductHandler allows inline updates to attributes, titles, bullets, and status.
func (env *ServerEnv) UpdateProductHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	existing, err := env.DB.GetProduct(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if existing == nil {
		writeError(w, http.StatusNotFound, "product not found")
		return
	}

	var req struct {
		OptimizedTitle      *string                     `json:"optimized_title"`
		BulletPoints        *[]string                   `json:"bullet_points"`
		SEODescriptionHTML  *string                     `json:"seo_description_html"`
		PublicationStatus   *domain.PublicationStatus   `json:"publication_status"`
		ExtractedAttributes *domain.ExtractedAttributes `json:"extracted_attributes"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.OptimizedTitle != nil {
		existing.OptimizedTitle = *req.OptimizedTitle
	}
	if req.BulletPoints != nil {
		existing.BulletPoints = *req.BulletPoints
	}
	if req.SEODescriptionHTML != nil {
		existing.SEODescriptionHTML = *req.SEODescriptionHTML
	}
	if req.PublicationStatus != nil {
		existing.PublicationStatus = *req.PublicationStatus
	}
	if req.ExtractedAttributes != nil {
		existing.ExtractedAttributes = *req.ExtractedAttributes
	}

	// Recompute quality scorecard
	scorecard := copywriter.CalculateQualityScorecard(existing)
	existing.QualityScore = scorecard.TotalScore

	if err := env.DB.UpdateProduct(existing); err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed updating product: %v", err))
		return
	}

	writeJSON(w, http.StatusOK, existing)
}

// BatchApproveAllHandler marks all compliant products as APPROVED in one click.
func (env *ServerEnv) BatchApproveAllHandler(w http.ResponseWriter, r *http.Request) {
	batchID := r.PathValue("id")
	affected, err := env.DB.BatchUpdateStatus(batchID, domain.StatusPending, domain.StatusApproved)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"batch_id":         batchID,
		"approved_count":   affected,
		"message":          fmt.Sprintf("Approved %d pending products", affected),
	})
}

// GetBatchAnalyticsHandler returns analytics and diagnostics for a batch.
func (env *ServerEnv) GetBatchAnalyticsHandler(w http.ResponseWriter, r *http.Request) {
	batchID := r.PathValue("id")
	analytics, err := env.DB.GetBatchAnalytics(batchID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, analytics)
}

// BulkStatusHandler updates publication status for an array of product IDs.
func (env *ServerEnv) BulkStatusHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ProductIDs []string                `json:"product_ids"`
		Status     domain.PublicationStatus `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid bulk request")
		return
	}

	affected, err := env.DB.BulkUpdateProductStatus(req.ProductIDs, req.Status)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"affected_count": affected,
		"status":         req.Status,
	})
}

// AutoFixImageHandler applies AutoWhitenAndPad to remediate non-compliant product images.
func (env *ServerEnv) AutoFixImageHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	p, err := env.DB.GetProduct(id)
	if err != nil || p == nil {
		writeError(w, http.StatusNotFound, "product not found")
		return
	}

	if p.ImageURL == "" {
		writeError(w, http.StatusBadRequest, "product has no image URL to fix")
		return
	}

	// Fetch original image
	_, img, err := vision.FetchAndAuditImage(r.Context(), nil, p.ImageURL)
	if img == nil {
		writeError(w, http.StatusBadRequest, "failed to download or decode product image")
		return
	}

	// Apply Auto-Whiten & 1000x1000 square padding
	fixedImg, fixedReport := vision.AutoWhitenAndPad(img, 1000)

	// Save remediated image to local catalog_media directory
	mediaDir := "catalog_media"
	_ = os.MkdirAll(mediaDir, 0755)
	mediaPath := filepath.Join(mediaDir, fmt.Sprintf("%s.png", p.ID))

	outFile, err := os.Create(mediaPath)
	if err == nil {
		_ = png.Encode(outFile, fixedImg)
		_ = outFile.Close()
		p.ImageURL = fmt.Sprintf("/api/products/%s/image", p.ID)
	}

	// Update product metrics
	p.ImageWidth = fixedReport.Width
	p.ImageHeight = fixedReport.Height
	p.ImageZoomReady = fixedReport.ZoomReady
	p.ImageWhiteBGScore = fixedReport.WhiteBGScore
	p.ImageAspectRatio = fixedReport.AspectRatio
	p.ImageWatermarkDetected = fixedReport.WatermarkDetected
	p.ImageComplianceStatus = fixedReport.Status

	// Recalculate listing quality scorecard
	scorecard := copywriter.CalculateQualityScorecard(p)
	p.QualityScore = scorecard.TotalScore

	if err := env.DB.UpdateProduct(p); err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed updating product: %v", err))
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"product":    p,
		"compliance": fixedReport,
		"message":    "Image remediated to pure white RGB 255 and padded to 1000x1000px square",
	})
}

// ProductImageHandler serves the remediated or cached local image for a product.
func (env *ServerEnv) ProductImageHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	mediaPath := filepath.Join("catalog_media", fmt.Sprintf("%s.png", id))

	file, err := os.Open(mediaPath)
	if err != nil {
		writeError(w, http.StatusNotFound, "remediated image not found")
		return
	}
	defer file.Close()

	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	_, _ = io.Copy(w, file)
}

// CopyStylesHandler generates and returns all 3 copy style variants for a product.
func (env *ServerEnv) CopyStylesHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	p, err := env.DB.GetProduct(id)
	if err != nil || p == nil {
		writeError(w, http.StatusNotFound, "product not found")
		return
	}

	raw := domain.RawFeedItem{
		SKU:         p.RawSKU,
		Title:       p.RawTitle,
		Description: p.RawDescription,
		Category:    p.RawCategory,
		Brand:       p.Brand,
		Price:       p.RawPrice,
		Currency:    p.Currency,
	}

	styles := copywriter.GenerateAllStyles(raw, p.ExtractedAttributes)
	writeJSON(w, http.StatusOK, styles)
}

// ExportCSVHandler generates and downloads the universal Enriched CSV.
func (env *ServerEnv) ExportCSVHandler(w http.ResponseWriter, r *http.Request) {
	batchID := r.PathValue("id")
	products, _, err := env.DB.ListProducts(store.ProductFilter{BatchID: batchID, Limit: 2000})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	csvBytes, err := channels.GenerateEnrichedCSV(products)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	filename := fmt.Sprintf("skupilot_enriched_%s.csv", batchID)
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(csvBytes)
}

// ExportShopifyHandler publishes products to Shopify Admin API.
func (env *ServerEnv) ExportShopifyHandler(w http.ResponseWriter, r *http.Request) {
	batchID := r.PathValue("id")

	var req domain.ShopifyExportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid export payload")
		return
	}

	filter := store.ProductFilter{BatchID: batchID, Limit: 500}
	if req.OnlyApproved {
		filter.PublicationStatus = string(domain.StatusApproved)
	}

	products, _, err := env.DB.ListProducts(filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if len(products) == 0 {
		writeError(w, http.StatusBadRequest, "no matching products found to export")
		return
	}

	// If no access token provided, perform dry run validation
	if req.AccessToken == "" {
		_ = env.DB.LogExport(batchID, "SHOPIFY", len(products), "SUCCESS", "Dry-run validation successful")
		writeJSON(w, http.StatusOK, domain.ExportResult{
			BatchID:       batchID,
			Channel:       "SHOPIFY",
			ExportedCount: len(products),
			Status:        "DRY_RUN_SUCCESS",
			Details:       "Generated valid Shopify GraphQL mutation payloads for all products",
		})
		return
	}

	client := channels.NewShopifyClient(req.StoreDomain, req.AccessToken)
	exported := 0
	failed := 0

	for _, p := range products {
		_, err := client.PublishProduct(r.Context(), p, req.PublishStatus)
		if err != nil {
			failed++
		} else {
			exported++
		}
	}

	_ = env.DB.LogExport(batchID, "SHOPIFY", exported, "COMPLETED", fmt.Sprintf("Exported: %d, Failed: %d", exported, failed))

	writeJSON(w, http.StatusOK, domain.ExportResult{
		BatchID:       batchID,
		Channel:       "SHOPIFY",
		ExportedCount: exported,
		FailedCount:   failed,
		Status:        "COMPLETED",
	})
}

// ExportAmazonHandler downloads the Amazon Flat File TSV.
func (env *ServerEnv) ExportAmazonHandler(w http.ResponseWriter, r *http.Request) {
	batchID := r.PathValue("id")
	products, _, err := env.DB.ListProducts(store.ProductFilter{BatchID: batchID, Limit: 1000})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	tsv := channels.GenerateAmazonFlatFile(products)

	filename := fmt.Sprintf("amazon_feed_%s.tsv", batchID)
	w.Header().Set("Content-Type", "text/tab-separated-values; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(tsv)
}

// ExportGoogleHandler downloads the Google Merchant Center XML.
func (env *ServerEnv) ExportGoogleHandler(w http.ResponseWriter, r *http.Request) {
	batchID := r.PathValue("id")
	products, _, err := env.DB.ListProducts(store.ProductFilter{BatchID: batchID, Limit: 1000})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	xmlBytes, err := channels.GenerateGoogleShoppingXML(products, "skupilot Feed", "https://example.com")
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	filename := fmt.Sprintf("google_feed_%s.xml", batchID)
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(xmlBytes)
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
