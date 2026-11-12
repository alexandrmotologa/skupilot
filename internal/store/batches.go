package store

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/alexandrmotologa/skupilot/internal/domain"
	"github.com/google/uuid"
)

// CreateBatch persists a new catalog batch record.
func (db *DB) CreateBatch(name string, totalSKUs int) (*domain.CatalogBatch, error) {
	id := fmt.Sprintf("batch_%s", uuid.New().String()[:8])
	now := time.Now().UTC()

	query := `
	INSERT INTO catalog_batches (id, name, total_skus, enriched_count, compliant_count, status, created_at, updated_at)
	VALUES (?, ?, ?, 0, 0, ?, ?, ?)
	`
	_, err := db.Exec(query, id, name, totalSKUs, string(domain.BatchStatusProcessing), now, now)
	if err != nil {
		return nil, fmt.Errorf("failed to create batch: %w", err)
	}

	return &domain.CatalogBatch{
		ID:             id,
		Name:           name,
		TotalSKUs:      totalSKUs,
		EnrichedCount:  0,
		CompliantCount: 0,
		Status:         domain.BatchStatusProcessing,
		CreatedAt:      now,
		UpdatedAt:      now,
	}, nil
}

// GetBatch retrieves a batch by ID.
func (db *DB) GetBatch(id string) (*domain.CatalogBatch, error) {
	query := `
	SELECT id, name, total_skus, enriched_count, compliant_count, status, created_at, updated_at
	FROM catalog_batches
	WHERE id = ?
	`
	row := db.QueryRow(query, id)

	var b domain.CatalogBatch
	var status string
	err := row.Scan(&b.ID, &b.Name, &b.TotalSKUs, &b.EnrichedCount, &b.CompliantCount, &status, &b.CreatedAt, &b.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get batch: %w", err)
	}

	b.Status = domain.BatchStatus(status)
	return &b, nil
}

// ListBatches returns all batches ordered by creation date descending.
func (db *DB) ListBatches() ([]domain.CatalogBatch, error) {
	query := `
	SELECT id, name, total_skus, enriched_count, compliant_count, status, created_at, updated_at
	FROM catalog_batches
	ORDER BY created_at DESC
	`
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list batches: %w", err)
	}
	defer rows.Close()

	var batches []domain.CatalogBatch
	for rows.Next() {
		var b domain.CatalogBatch
		var status string
		if err := rows.Scan(&b.ID, &b.Name, &b.TotalSKUs, &b.EnrichedCount, &b.CompliantCount, &status, &b.CreatedAt, &b.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan batch row: %w", err)
		}
		b.Status = domain.BatchStatus(status)
		batches = append(batches, b)
	}

	return batches, nil
}

// UpdateBatchProgress increments progress counters and updates status.
func (db *DB) UpdateBatchProgress(id string, enrichedCount, compliantCount int, status domain.BatchStatus) error {
	query := `
	UPDATE catalog_batches
	SET enriched_count = ?, compliant_count = ?, status = ?, updated_at = ?
	WHERE id = ?
	`
	_, err := db.Exec(query, enrichedCount, compliantCount, string(status), time.Now().UTC(), id)
	return err
}

// LogExport saves an export transaction log.
func (db *DB) LogExport(batchID, channel string, count int, status, details string) error {
	id := fmt.Sprintf("exp_%s", uuid.New().String()[:8])
	query := `
	INSERT INTO export_logs (id, batch_id, channel, item_count, status, details, created_at)
	VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	_, err := db.Exec(query, id, batchID, channel, count, status, details, time.Now().UTC())
	return err
}
