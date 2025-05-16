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
	"context"
	"errors"
	"testing"
	"time"

	authpkg "github.com/andranikasd/Zeno/internal/aws/auth"
	"github.com/andranikasd/Zeno/internal/config"
	"github.com/aws/aws-sdk-go/aws/session"
)

// TestFetchAndPrintReports_AuthFail verifies that when AWS auth fails,
// FetchAndPrintReports returns that error immediately.
func TestFetchAndPrintReports_AuthFail(t *testing.T) {
	// Stub authpkg.NewSession to always error.
	origNewSession := authpkg.NewSession
	authpkg.NewSession = func(cfg *config.AWSConfig) (*session.Session, error) {
		return nil, errors.New("mock auth failure")
	}
	defer func() { authpkg.NewSession = origNewSession }()

	cfg := &config.Config{
		AWS: config.AWSConfig{
			Region: "us-west-2",
			Auth: []config.AuthMethod{
				{Type: "credentials", AccessKeyID: "X", SecretKey: "Y"},
			},
		},
		CUR: config.CURConfig{
			Bucket: "bucket",
			Prefix: "prefix",
			Format: "csv",
		},
	}

	err := FetchAndPrintReports(context.Background(), cfg, time.Now())
	if err == nil || !errors.Is(err, errors.New("mock auth failure")) {
		t.Fatalf("expected auth error, got %v", err)
	}
}
