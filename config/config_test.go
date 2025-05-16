// Copyright 2015 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"
)

// discard logger
var testLogger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

func TestLoadFiles(t *testing.T) {
    tests := []struct {
        fileName string
        wantErr  bool
    }{
        {"config-valid-1.yaml", false},
        {"config-valid-2.yaml", false},
        {"config-valid-3.yaml", false},
        {"config-invalid-1.yaml", true},
        {"config-invalid-2.yaml", true},
        {"config-invalid-3.yaml", true},
    }

    baseDir := filepath.Join("test_data")
    for _, tc := range tests {
        t.Run(tc.fileName, func(t *testing.T) {
            path := filepath.Join(baseDir, tc.fileName)
            data, err := os.ReadFile(path)
            if err != nil {
                t.Fatalf("failed to read %s: %v", tc.fileName, err)
            }
            _, err = Load(string(data), testLogger)
            if tc.wantErr {
                if err == nil {
                    t.Errorf("%s: expected error, got nil", tc.fileName)
                }
            } else {
                if err != nil {
                    t.Errorf("%s: expected no error, got %v", tc.fileName, err)
                }
            }
        })
    }
}