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
	"fmt"
	"log/slog"
	"os"

	"gopkg.in/yaml.v2"
)

// Config models the top-level YAML configuration for Zeno.
type Config struct {
	AWS AWSConfig `yaml:"aws"`
	CUR CURConfig  `yaml:"cur"`
}

// AWSConfig specifies AWS settings, including region and auth methods.
type AWSConfig struct {
	Region string       `yaml:"region"`
	Auth   []AuthMethod `yaml:"auth"`
}

// AuthMethod represents an AWS authentication method.
// Type must be one of "credentials", "profile", or "iam".
type AuthMethod struct {
	Type         string `yaml:"type"`
	AccessKeyID  string `yaml:"accessKeyID,omitempty"`
	SecretKey    string `yaml:"secretKey,omitempty"`
	ProfilePath  string `yaml:"path,omitempty"`
	ProfileName  string `yaml:"profile,omitempty"`
	RoleToAssume string `yaml:"role-to-assume,omitempty"`
}

// CURConfig specifies Cost & Usage Report settings.
type CURConfig struct {
	Bucket   string `yaml:"bucket"`   // S3 bucket storing CUR files
	Prefix   string `yaml:"prefix"`   // S3 key prefix for CUR objects
	Region   string `yaml:"region"`   // optional override for bucket region
	Schedule string `yaml:"schedule"` // cron schedule for daily ingestion
}

// LoadFile reads and parses the YAML configuration at path.
func LoadFile(path string, logger *slog.Logger) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file %q: %w", path, err)
	}

	cfg, err := Load(data, logger)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

// Load unmarshals the YAML data into a Config, validates it, and logs metadata.
func Load(data []byte, logger *slog.Logger) (*Config, error) {
	var cfg Config
	if err := yaml.UnmarshalStrict(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	// Validate region
	if cfg.AWS.Region == "" {
		return nil, fmt.Errorf("config.aws.region is required")
	}

	// Validate auth methods
	if len(cfg.AWS.Auth) == 0 {
		return nil, fmt.Errorf("config.aws.auth must have at least one method")
	}

	// Validate CUR section
	if cfg.CUR.Bucket == "" {
		return nil, fmt.Errorf("config.cur.bucket is required")
	}
	if cfg.CUR.Prefix == "" {
		return nil, fmt.Errorf("config.cur.prefix is required")
	}
	if cfg.CUR.Schedule == "" {
		return nil, fmt.Errorf("config.cur.schedule is required")
	}


	for i, m := range cfg.AWS.Auth {
		switch m.Type {
		case "credentials":
			if m.AccessKeyID == "" || m.SecretKey == "" {
				return nil, fmt.Errorf("auth[%d]: credentials require accessKeyID and secretKey", i)
			}

		case "profile":
			if m.ProfilePath == "" || m.ProfileName == "" {
				return nil, fmt.Errorf("auth[%d]: profile requires path and profile", i)
			}

		case "iam":
			if m.RoleToAssume == "" {
				return nil, fmt.Errorf("auth[%d]: iam requires role-to-assume", i)
			}

		default:
			return nil, fmt.Errorf("auth[%d]: unknown type %q", i, m.Type)
		}
	}

	logger.Info("config loaded",
		"region", cfg.AWS.Region,
		"methods", len(cfg.AWS.Auth),
	)

	logger.Info("config loaded",
		"aws_region", cfg.AWS.Region,
		"auth_methods", len(cfg.AWS.Auth),
		"cur_bucket", cfg.CUR.Bucket,
		"cur_prefix", cfg.CUR.Prefix,
	)

	return &cfg, nil
}