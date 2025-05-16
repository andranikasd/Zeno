// Copyright 2015 The Zeno Authors
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

package config

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"
)

// testLogger discards log output during tests.
var dummyLogger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
// TestLoadFiles verifies Load against YAML files in test_data.
func TestLoadFiles(t *testing.T) {

tests := []struct{
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

baseDir := "test_data"
for _, tc := range tests {
	stName := tc.fileName
	t.Run(stName, func(t *testing.T) {
		path := filepath.Join(baseDir, stName)
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read file %q: %v", stName, err)
		}

		_, err = Load(data, dummyLogger)
		if tc.wantErr {
			if err == nil {
				t.Errorf("%s: expected error, got nil", stName)
			}
		} else {
			if err != nil {
				t.Errorf("%s: expected no error, got %v", stName, err)
			}
		}
	})
}
}