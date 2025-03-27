package ingest

import (
	"database/sql"
	"fmt"

	"github.com/andranikasd/Zeno/internal/config"
)

func ProcessCUR(cfg *config.Config, db *sql.DB) error {
	fmt.Println("[STUB] Ingest CUR from S3 bucket:", cfg.S3Bucket)
	// TODO: Use AWS SDK to download and parse CSV files
	return nil
}
