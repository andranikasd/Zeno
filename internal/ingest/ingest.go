// internal/ingest/s3.go
package ingest

import (
	"archive/zip"
	"compress/gzip"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	zcfg "github.com/andranikasd/Zeno/internal/config"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Manifest struct {
	ReportKeys []string `json:"reportKeys"`
}

func loadAWSConfig(cfg *zcfg.Config) (aws.Config, error) {
	accessKey := os.Getenv("AWS_ACCESS_KEY_ID")
	secretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	sessionToken := os.Getenv("AWS_SESSION_TOKEN")

	if accessKey == "" {
		accessKey = cfg.AWSAccessKeyID
	}
	if secretKey == "" {
		secretKey = cfg.AWSSecretAccessKey
	}

	if accessKey == "" || secretKey == "" {
		return aws.Config{}, fmt.Errorf("❌ AWS credentials are missing. Set them in config.yaml or as environment variables")
	}

	log.Printf("🔐 Using AWS credentials for user: %s", accessKey)
	return config.LoadDefaultConfig(
		context.TODO(),
		config.WithRegion(cfg.S3Region),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(accessKey, secretKey, sessionToken),
		),
	)
}

func FindManifestFiles(cfg *zcfg.Config) ([]string, error) {
	awsCfg, err := loadAWSConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("AWS config error: %w", err)
	}

	client := s3.NewFromConfig(awsCfg)
	var manifests []string

	log.Println("📡 Listing manifest files in S3 bucket...")
	paginator := s3.NewListObjectsV2Paginator(client, &s3.ListObjectsV2Input{
		Bucket: &cfg.S3Bucket,
		Prefix: &cfg.S3Prefix,
	})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			return nil, fmt.Errorf("pagination error: %w", err)
		}
		for _, obj := range page.Contents {
			if strings.HasSuffix(*obj.Key, "Manifest.json") {
				manifests = append(manifests, *obj.Key)
				log.Printf("📄 Found manifest: %s", *obj.Key)
			}
		}
	}

	sort.Slice(manifests, func(i, j int) bool {
		return manifestTimestamp(manifests[i]) > manifestTimestamp(manifests[j])
	})

	if !cfg.IngestAll && len(manifests) > 1 {
		manifests = manifests[:1]
	}

	return manifests, nil
}

func manifestTimestamp(manifestKey string) string {
	parts := strings.Split(manifestKey, "/")
	for _, part := range parts {
		if strings.Contains(part, "T") && strings.HasSuffix(part, "Z") {
			return part
		}
	}
	return "default"
}

func ParseManifest(cfg *zcfg.Config, key string) ([]string, error) {
	awsCfg, err := loadAWSConfig(cfg)
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(awsCfg)
	log.Printf("📥 Downloading manifest: %s", key)

	_, ts := filepath.Split(filepath.Dir(key))
	manifestPath := filepath.Join(cfg.CachePath, "manifests", fmt.Sprintf("%s.json", ts))
	os.MkdirAll(filepath.Dir(manifestPath), 0755)

	obj, err := client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: &cfg.S3Bucket,
		Key:    &key,
	})
	if err != nil {
		return nil, fmt.Errorf("S3 GetObject error: %w", err)
	}
	defer obj.Body.Close()

	raw, err := io.ReadAll(obj.Body)
	if err != nil {
		return nil, fmt.Errorf("read manifest error: %w", err)
	}
	os.WriteFile(manifestPath, raw, 0644)

	var manifest Manifest
	if err := json.Unmarshal(raw, &manifest); err != nil {
		return nil, fmt.Errorf("JSON decode error: %w", err)
	}
	log.Printf("📊 Manifest has %d report keys", len(manifest.ReportKeys))
	return manifest.ReportKeys, nil
}

func extractGzipFile(gzPath, outputDir string) (string, error) {
	inFile, err := os.Open(gzPath)
	if err != nil {
		return "", fmt.Errorf("open gzip error: %w", err)
	}
	defer inFile.Close()

	gzReader, err := gzip.NewReader(inFile)
	if err != nil {
		return "", fmt.Errorf("gzip reader error: %w", err)
	}
	defer gzReader.Close()

	outPath := filepath.Join(outputDir, strings.TrimSuffix(filepath.Base(gzPath), ".gz"))
	outFile, err := os.Create(outPath)
	if err != nil {
		return "", fmt.Errorf("create output error: %w", err)
	}
	defer outFile.Close()

	if _, err := io.Copy(outFile, gzReader); err != nil {
		return "", fmt.Errorf("copy gzip error: %w", err)
	}

	log.Printf("✅ Extracted: %s", outPath)
	return outPath, nil
}

func DownloadAndExtractCSVZips(cfg *zcfg.Config, manifestKey string, keys []string) ([]string, error) {
	awsCfg, err := loadAWSConfig(cfg)
	if err != nil {
		return nil, err
	}
	client := s3.NewFromConfig(awsCfg)

	ts := manifestTimestamp(manifestKey)
	localDir := filepath.Join(cfg.CachePath, "data", ts)
	os.MkdirAll(localDir, 0755)

	var extracted []string

	for _, key := range keys {
		log.Printf("⬇️ Downloading zip file: %s", key)

		outPath := filepath.Join(localDir, filepath.Base(key))
		f, err := os.Create(outPath)
		if err != nil {
			return nil, fmt.Errorf("create zip error: %w", err)
		}
		resp, err := client.GetObject(context.TODO(), &s3.GetObjectInput{
			Bucket: &cfg.S3Bucket,
			Key:    &key,
		})
		if err != nil {
			f.Close()
			return nil, fmt.Errorf("GetObject failed: %w", err)
		}
		if _, err := io.Copy(f, resp.Body); err != nil {
			f.Close()
			resp.Body.Close()
			return nil, fmt.Errorf("download copy error: %w", err)
		}
		f.Close()
		resp.Body.Close()

		log.Printf("📦 Saved zip: %s", outPath)

		if strings.HasSuffix(outPath, ".gz") {
			outFile, err := extractGzipFile(outPath, localDir)
			if err != nil {
				return nil, err
			}
			extracted = append(extracted, outFile)
			continue
		}

		r, err := zip.OpenReader(outPath)
		if err != nil {
			return nil, fmt.Errorf("zip open error: %w", err)
		}
		for _, zf := range r.File {
			rc, err := zf.Open()
			if err != nil {
				continue
			}
			extractPath := filepath.Join(localDir, filepath.Base(zf.Name))
			of, err := os.Create(extractPath)
			if err != nil {
				rc.Close()
				continue
			}
			if _, err := io.Copy(of, rc); err != nil {
				of.Close()
				rc.Close()
				continue
			}
			of.Close()
			rc.Close()
			extracted = append(extracted, extractPath)
			log.Printf("✅ Extracted: %s", extractPath)
		}
		r.Close()
	}
	return extracted, nil
}

func loadIntoDuckDB(cfg *zcfg.Config, db *sql.DB, csvPath string) error {
	table := "cur_raw"
	query := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s AS SELECT * FROM read_csv_auto('%s', AUTO_DETECT=TRUE, HEADER=TRUE);`, table, csvPath)
	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("DuckDB load error: %w", err)
	}
	log.Printf("🦆 Loaded into DuckDB: %s", csvPath)
	return nil
}

func ProcessCUR(cfg *zcfg.Config) error {
	log.Println("🚀 Starting CUR ingest pipeline...")

	db, err := sql.Open("duckdb", cfg.DatabasePath)
	if err != nil {
		return fmt.Errorf("❌ Failed to open DuckDB: %w", err)
	}
	defer db.Close()

	manifests, err := FindManifestFiles(cfg)
	if err != nil {
		return fmt.Errorf("❌ Failed to find manifests: %w", err)
	}
	if len(manifests) == 0 {
		log.Println("⚠️  No CUR manifests found. Exiting.")
		return nil
	}

	for _, manifest := range manifests {
		log.Printf("🔍 Processing manifest: %s", manifest)

		reportKeys, err := ParseManifest(cfg, manifest)
		if err != nil {
			log.Printf("❌ Could not parse manifest: %v", err)
			continue
		}

		files, err := DownloadAndExtractCSVZips(cfg, manifest, reportKeys)
		if err != nil {
			log.Printf("❌ Error downloading or extracting CSVs: %v", err)
			continue
		}

		for _, file := range files {
			if err := loadIntoDuckDB(cfg, db, file); err != nil {
				log.Printf("❌ Failed loading file to DuckDB: %v", err)
			}
		}

		log.Printf("📁 Finished manifest. %d files extracted and loaded.", len(files))
	}
	log.Println("✅ CUR ingestion complete.")
	return nil
}