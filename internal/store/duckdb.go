// internal/store/duckdb.go
package store

import (
	"database/sql"
	"fmt"

	_ "github.com/marcboeker/go-duckdb"
)

func Open(path string) (*sql.DB, error) {
	db, err := sql.Open("duckdb", path)
	if err != nil {
		return nil, fmt.Errorf("open duckdb error: %w", err)
	}
	return db, nil
}

func IngestCSV(db *sql.DB, tableName, filePath string) error {
	query := fmt.Sprintf(`CREATE OR REPLACE TABLE %s AS SELECT * FROM read_csv_auto('%s');`, tableName, filePath)
	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("ingest csv error: %w", err)
	}
	return nil
}
