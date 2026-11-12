package store

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/alexandrmotologa/skupilot/internal/domain"
	"github.com/google/uuid"
)

// ProductFilter parameters for querying products.
type ProductFilter struct {
	BatchID           string
	PublicationStatus string
	ComplianceStatus  string
	Search            string
	Limit             int
	Offset            int
}

// InsertProduct saves a new product SKU record.
func (db *DB) InsertProduct(p *domain.Product) error {
	if p.ID == "" {
		p.ID = fmt.Sprintf("prod_%s", uuid.New().String()[:8])
	}
	now := time.Now().UTC()
	if p.CreatedAt.IsZero() {
		p.CreatedAt = now
	}
	p.UpdatedAt = now

	attrJSON, err := domain.MarshalAttributes(p.ExtractedAttributes)
	if err != nil {
		return fmt.Errorf("failed to marshal attributes: %w", err)
	}

	bulletsJSON, err := domain.MarshalBullets(p.BulletPoints)
	if err != nil {
		return fmt.Errorf("failed to marshal bullets: %w", err)
	}

	query := `
	INSERT INTO products (
		id, batch_id, raw_sku, raw_title, raw_description, raw_category, raw_price, currency, brand, image_url,
		image_width, image_height, image_zoom_ready, image_white_bg_score, image_aspect_ratio, image_watermark_detected, image_compliance_status,
		extracted_attributes_json, optimized_title, bullet_points_json, seo_description_html,
		google_taxonomy_id, google_taxonomy_path, quality_score, publication_status, created_at, updated_at
	) VALUES (
		?, ?, ?, ?, ?, ?, ?, ?, ?, ?,
		?, ?, ?, ?, ?, ?, ?,
		?, ?, ?, ?,
		?, ?, ?, ?, ?, ?
	)
	`
	_, err = db.Exec(query,
		p.ID, p.BatchID, p.RawSKU, p.RawTitle, p.RawDescription, p.RawCategory, p.RawPrice, p.Currency, p.Brand, p.ImageURL,
		p.ImageWidth, p.ImageHeight, p.ImageZoomReady, p.ImageWhiteBGScore, p.ImageAspectRatio, p.ImageWatermarkDetected, string(p.ImageComplianceStatus),
		attrJSON, p.OptimizedTitle, bulletsJSON, p.SEODescriptionHTML,
		p.GoogleTaxonomyID, p.GoogleTaxonomyPath, p.QualityScore, string(p.PublicationStatus), p.CreatedAt, p.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert product: %w", err)
	}
	return nil
}

// GetProduct retrieves a product by its ID.
func (db *DB) GetProduct(id string) (*domain.Product, error) {
	query := `
	SELECT
		id, batch_id, raw_sku, raw_title, raw_description, raw_category, raw_price, currency, brand, image_url,
		image_width, image_height, image_zoom_ready, image_white_bg_score, image_aspect_ratio, image_watermark_detected, image_compliance_status,
		extracted_attributes_json, optimized_title, bullet_points_json, seo_description_html,
		google_taxonomy_id, google_taxonomy_path, quality_score, publication_status, created_at, updated_at
	FROM products
	WHERE id = ?
	`
	row := db.QueryRow(query, id)

	p, err := scanProduct(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get product %s: %w", id, err)
	}
	return p, nil
}

// ListProducts retrieves filtered products and the total count.
func (db *DB) ListProducts(filter ProductFilter) ([]domain.Product, int, error) {
	var whereClauses []string
	var args []interface{}

	if filter.BatchID != "" {
		whereClauses = append(whereClauses, "batch_id = ?")
		args = append(args, filter.BatchID)
	}
	if filter.PublicationStatus != "" {
		whereClauses = append(whereClauses, "publication_status = ?")
		args = append(args, filter.PublicationStatus)
	}
	if filter.ComplianceStatus != "" {
		whereClauses = append(whereClauses, "image_compliance_status = ?")
		args = append(args, filter.ComplianceStatus)
	}
	if filter.Search != "" {
		whereClauses = append(whereClauses, "(raw_sku LIKE ? OR raw_title LIKE ? OR optimized_title LIKE ?)")
		pattern := "%" + filter.Search + "%"
		args = append(args, pattern, pattern, pattern)
	}

	whereSQL := ""
	if len(whereClauses) > 0 {
		whereSQL = "WHERE " + strings.Join(whereClauses, " AND ")
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM products %s", whereSQL)
	var total int
	if err := db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count products: %w", err)
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	selectQuery := fmt.Sprintf(`
	SELECT
		id, batch_id, raw_sku, raw_title, raw_description, raw_category, raw_price, currency, brand, image_url,
		image_width, image_height, image_zoom_ready, image_white_bg_score, image_aspect_ratio, image_watermark_detected, image_compliance_status,
		extracted_attributes_json, optimized_title, bullet_points_json, seo_description_html,
		google_taxonomy_id, google_taxonomy_path, quality_score, publication_status, created_at, updated_at
	FROM products
	%s
	ORDER BY created_at DESC
	LIMIT ? OFFSET ?
	`, whereSQL)

	selectArgs := append(args, limit, offset)
	rows, err := db.Query(selectQuery, selectArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list products: %w", err)
	}
	defer rows.Close()

	var products []domain.Product
	for rows.Next() {
		p, err := scanProduct(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan product: %w", err)
		}
		products = append(products, *p)
	}

	return products, total, nil
}

// UpdateProduct updates product details and refreshed quality score.
func (db *DB) UpdateProduct(p *domain.Product) error {
	p.UpdatedAt = time.Now().UTC()

	attrJSON, err := domain.MarshalAttributes(p.ExtractedAttributes)
	if err != nil {
		return fmt.Errorf("failed to marshal attributes: %w", err)
	}
	bulletsJSON, err := domain.MarshalBullets(p.BulletPoints)
	if err != nil {
		return fmt.Errorf("failed to marshal bullets: %w", err)
	}

	query := `
	UPDATE products SET
		optimized_title = ?,
		bullet_points_json = ?,
		seo_description_html = ?,
		extracted_attributes_json = ?,
		quality_score = ?,
		publication_status = ?,
		image_compliance_status = ?,
		updated_at = ?
	WHERE id = ?
	`
	_, err = db.Exec(query,
		p.OptimizedTitle,
		bulletsJSON,
		p.SEODescriptionHTML,
		attrJSON,
		p.QualityScore,
		string(p.PublicationStatus),
		string(p.ImageComplianceStatus),
		p.UpdatedAt,
		p.ID,
	)
	return err
}

// BatchUpdateStatus changes publication status for multiple products in a batch.
func (db *DB) BatchUpdateStatus(batchID string, fromStatus, toStatus domain.PublicationStatus) (int64, error) {
	query := `
	UPDATE products
	SET publication_status = ?, updated_at = ?
	WHERE batch_id = ? AND publication_status = ?
	`
	res, err := db.Exec(query, string(toStatus), time.Now().UTC(), batchID, string(fromStatus))
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// Helper scanner to reduce duplication between QueryRow and Query.
type scannable interface {
	Scan(dest ...interface{}) error
}

func scanProduct(s scannable) (*domain.Product, error) {
	var p domain.Product
	var complianceStatus, pubStatus string
	var attrJSON, bulletsJSON string
	var rawDesc, rawCat, brand, optTitle, seoHTML, taxPath sql.NullString

	err := s.Scan(
		&p.ID, &p.BatchID, &p.RawSKU, &p.RawTitle, &rawDesc, &rawCat, &p.RawPrice, &p.Currency, &brand, &p.ImageURL,
		&p.ImageWidth, &p.ImageHeight, &p.ImageZoomReady, &p.ImageWhiteBGScore, &p.ImageAspectRatio, &p.ImageWatermarkDetected, &complianceStatus,
		&attrJSON, &optTitle, &bulletsJSON, &seoHTML,
		&p.GoogleTaxonomyID, &taxPath, &p.QualityScore, &pubStatus, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	p.RawDescription = rawDesc.String
	p.RawCategory = rawCat.String
	p.Brand = brand.String
	p.OptimizedTitle = optTitle.String
	p.SEODescriptionHTML = seoHTML.String
	p.GoogleTaxonomyPath = taxPath.String
	p.ImageComplianceStatus = domain.ComplianceStatus(complianceStatus)
	p.PublicationStatus = domain.PublicationStatus(pubStatus)

	attr, err := domain.UnmarshalAttributes(attrJSON)
	if err == nil {
		p.ExtractedAttributes = attr
	}

	bullets, err := domain.UnmarshalBullets(bulletsJSON)
	if err == nil {
		p.BulletPoints = bullets
	}

	return &p, nil
}
