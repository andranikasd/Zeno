package store

import (
	"database/sql"

	_ "github.com/marcboeker/go-duckdb"
)

func InitDuckDB(path string) (*sql.DB, error) {
	db, err := sql.Open("duckdb", path)
	if err != nil {
		return nil, err
	}
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS aws_costs (
		line_item_id TEXT,
		service TEXT,
		usage_date DATE,
		cost FLOAT
	)`)
	if err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}