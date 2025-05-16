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

package aws

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"log/slog"

	"github.com/andranikasd/Zeno/internal/config"
)

// testLogger discards log output during auth tests.
var testLogger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

// TestNewSessionWithStaticCredentials verifies that a session is created
// when a valid credentials auth method is provided.
func TestNewSessionWithStaticCredentials(t *testing.T) {
	cfg := &config.AWSConfig{
		Region: "us-west-2",
		Auth: []config.AuthMethod{{
			Type:        "credentials",
			AccessKeyID: "STATIC_ID",
			SecretKey:   "STATIC_SECRET",
		}},
	}
	sess, err := NewSession(cfg)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	creds, err := sess.Config.Credentials.Get()
	if err != nil {
		t.Fatalf("expected credentials.Get to succeed, got %v", err)
	}

	if creds.AccessKeyID != "STATIC_ID" || creds.SecretAccessKey != "STATIC_SECRET" {
		t.Errorf("unexpected credentials: %v", creds)
	}
}

// TestNewSessionWithProfile verifies that a session is created when a
// valid profile auth method is provided using AWS_SHARED_CREDENTIALS_FILE.
func TestNewSessionWithProfile(t *testing.T) {
	tmpDir := t.TempDir()
	credContent := []byte("[demo]\naws_access_key_id=PROFILE_ID\naws_secret_access_key=PROFILE_SECRET\n")
	credFile := filepath.Join(tmpDir, "credentials")
	if err := os.WriteFile(credFile, credContent, 0600); err != nil {
		t.Fatalf("write credentials file: %v", err)
	}

	// Point AWS SDK to our temp credentials file
	existing := os.Getenv("AWS_SHARED_CREDENTIALS_FILE")
	_ = os.Setenv("AWS_SHARED_CREDENTIALS_FILE", credFile)
	defer func() {
		_ = os.Setenv("AWS_SHARED_CREDENTIALS_FILE", existing)
	}()

	cfg := &config.AWSConfig{
		Region: "us-east-1",
		Auth: []config.AuthMethod{{
			Type:        "profile",
			ProfileName: "demo",
		}},
	}
	sess, err := NewSession(cfg)
	if err != nil {
		t.Fatalf("expected no error for profile auth, got %v", err)
	}

	credsValue, err := sess.Config.Credentials.Get()
	if err != nil {
		t.Fatalf("credentials.Get failed: %v", err)
	}

	if credsValue.AccessKeyID != "PROFILE_ID" || credsValue.SecretAccessKey != "PROFILE_SECRET" {
		t.Errorf("unexpected profile credentials: %v", credsValue)
	}
}

// TestNewSessionWithIAMRole ensures that IAM auth fails when the role is invalid.
func TestNewSessionWithIAMRole(t *testing.T) {
	cfg := &config.AWSConfig{
		Region: "us-west-1",
		Auth: []config.AuthMethod{{
			Type:         "iam",
			RoleToAssume: "arn:aws:iam::000000000000:role/NonExistent",
		}},
	}

	_, err := NewSession(cfg)
	if err == nil || !strings.Contains(err.Error(), "no valid AWS auth method") {
		t.Errorf("expected IAM failure, got %v", err)
	}
}

// TestNewSessionWithUnknownType verifies that unknown auth types return an error.
func TestNewSessionWithUnknownType(t *testing.T) {
	cfg := &config.AWSConfig{
		Region: "us-east-2",
		Auth: []config.AuthMethod{{
			Type: "unsupported",
		}},
	}

	_, err := NewSession(cfg)
	if err == nil || !strings.Contains(err.Error(), "unknown auth type") {
		t.Errorf("expected unknown type error, got %v", err)
	}
}