// Copyright 2025 The Zeno Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cur

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	awssdk "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"

	authpkg "github.com/andranikasd/Zeno/internal/aws/auth"
	"github.com/andranikasd/Zeno/internal/config"
)

// FetchAndPrintReports authenticates with AWS, lists CUR files for the given date,
// downloads and prints their first few lines in CSV or stubs Parquet.
func FetchAndPrintReports(ctx context.Context, cfg *config.Config, date time.Time) error {
	// 1) Authenticate via internal/aws/auth
	sess, err := authpkg.NewSession(&cfg.AWS)
	if err != nil {
		return fmt.Errorf("AWS authentication failed: %w", err)
	}

	// 2) Build S3 client
	s3svc := s3.New(sess)

	// 3) AWS CUR path uses <prefix>/<YYYYMMDD-YYYYMMDD>/
	pattern := date.Format("20060102") + "-" + date.Format("20060102")
	prefix := filepath.Join(cfg.CUR.Prefix, pattern)

	// 4) List objects under that prefix
	var keys []string
	listIn := &s3.ListObjectsV2Input{
		Bucket: awssdk.String(cfg.CUR.Bucket),
		Prefix: awssdk.String(prefix),
	}
	if err := s3svc.ListObjectsV2PagesWithContext(ctx, listIn, func(page *s3.ListObjectsV2Output, _ bool) bool {
		for _, obj := range page.Contents {
			keys = append(keys, awssdk.StringValue(obj.Key))
		}
		return true
	}); err != nil {
		return fmt.Errorf("list CUR objects: %w", err)
	}
	if len(keys) == 0 {
		log.Printf("no CUR reports found for date %s", date.Format("2006-01-02"))
		return nil
	}

	// 5) Download each object with the downloader
	downloader := s3manager.NewDownloader(sess)
	for _, key := range keys {
		buf := aws.NewWriteAtBuffer([]byte{})
		_, err := downloader.DownloadWithContext(ctx, buf, &s3.GetObjectInput{
			Bucket: awssdk.String(cfg.CUR.Bucket),
			Key:    awssdk.String(key),
		})
		if err != nil {
			return fmt.Errorf("download failed for %s: %w", key, err)
		}

		data := buf.Bytes()
		// 6) If GZIP-compressed, decompress
		if strings.HasSuffix(key, ".gz") {
			gzr, err := gzip.NewReader(bytes.NewReader(data))
			if err != nil {
				return fmt.Errorf("gzip reader for %s: %w", key, err)
			}
			decompressed, err := io.ReadAll(gzr)
			gzr.Close()
			if err != nil {
				return fmt.Errorf("gzip decompress for %s: %w", key, err)
			}
			data = decompressed
		}

		fmt.Printf("=== Report: %s ===\n", key)
		switch strings.ToLower(cfg.CUR.Format) {
		case "", "csv":
			scanner := bufio.NewScanner(bytes.NewReader(data))
			for i := 0; i < 5 && scanner.Scan(); i++ {
				fmt.Println(scanner.Text())
			}
			if err := scanner.Err(); err != nil {
				return fmt.Errorf("reading CSV %s: %w", key, err)
			}
		case "parquet":
			// Placeholder for Parquet parsing
			fmt.Println("(Parquet report; parsing not implemented)")
		default:
			return fmt.Errorf("unsupported CUR format %q", cfg.CUR.Format)
		}
	}

	return nil
}
