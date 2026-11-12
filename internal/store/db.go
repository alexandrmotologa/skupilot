package store

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

// DB wraps the SQL database handle and manages schema migrations.
type DB struct {
	*sql.DB
}

// Open creates or connects to the SQLite database at dbPath and executes migrations.
func Open(dbPath string) (*DB, error) {
	// Enable WAL mode, busy timeout, and foreign keys for high throughput and reliability
	dsn := fmt.Sprintf("%s?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=foreign_keys(ON)", dbPath)
	conn, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite database: %w", err)
	}

	if err := conn.Ping(); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("failed to ping sqlite database: %w", err)
	}

	db := &DB{DB: conn}
	if err := db.migrate(); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("failed to run database migrations: %w", err)
	}

	return db, nil
}

// migrate executes the DDL schema statements.
func (db *DB) migrate() error {
	schema := `
	CREATE TABLE IF NOT EXISTS catalog_batches (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		total_skus INTEGER NOT NULL DEFAULT 0,
		enriched_count INTEGER NOT NULL DEFAULT 0,
		compliant_count INTEGER NOT NULL DEFAULT 0,
		status TEXT NOT NULL DEFAULT 'PROCESSING',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS products (
		id TEXT PRIMARY KEY,
		batch_id TEXT NOT NULL REFERENCES catalog_batches(id) ON DELETE CASCADE,
		raw_sku TEXT NOT NULL,
		raw_title TEXT NOT NULL,
		raw_description TEXT,
		raw_category TEXT,
		raw_price REAL DEFAULT 0.0,
		currency TEXT DEFAULT 'USD',
		brand TEXT,
		image_url TEXT NOT NULL,
		
		image_width INTEGER DEFAULT 0,
		image_height INTEGER DEFAULT 0,
		image_zoom_ready BOOLEAN DEFAULT 0,
		image_white_bg_score REAL DEFAULT 0.0,
		image_aspect_ratio REAL DEFAULT 1.0,
		image_watermark_detected BOOLEAN DEFAULT 0,
		image_compliance_status TEXT DEFAULT 'PENDING',
		
		extracted_attributes_json TEXT DEFAULT '{}',
		optimized_title TEXT,
		bullet_points_json TEXT DEFAULT '[]',
		seo_description_html TEXT,
		google_taxonomy_id INTEGER DEFAULT 0,
		google_taxonomy_path TEXT,
		
		quality_score INTEGER DEFAULT 0,
		publication_status TEXT NOT NULL DEFAULT 'PENDING',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_products_batch_id ON products(batch_id);
	CREATE INDEX IF NOT EXISTS idx_products_publication_status ON products(publication_status);
	CREATE INDEX IF NOT EXISTS idx_products_compliance_status ON products(image_compliance_status);

	CREATE TABLE IF NOT EXISTS export_logs (
		id TEXT PRIMARY KEY,
		batch_id TEXT NOT NULL REFERENCES catalog_batches(id) ON DELETE CASCADE,
		channel TEXT NOT NULL,
		item_count INTEGER NOT NULL,
		status TEXT NOT NULL,
		details TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`
	_, err := db.Exec(schema)
	return err
}
