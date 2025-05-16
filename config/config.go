// Copyright 2015 The Zeno Authors
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
	"fmt"
	"io/ioutil"
	"log/slog"

	"gopkg.in/yaml.v2"
)

// Config is the top‐level YAML model.
type Config struct {
    AWS AWSConfig `yaml:"aws"`
    // … Can add other top‐level sections here later …
}

// AWSConfig holds AWS‐specific settings.
type AWSConfig struct {
    Region string       `yaml:"region"`
    Auth   []AuthMethod `yaml:"auth"`
}

// AuthMethod is one entry in the auth list. The Type field
// determines which of the other fields is expected.
type AuthMethod struct {
    Type          string `yaml:"type"`                      // "credentials", "profile" or "iam"
    AccessKeyID   string `yaml:"accessKeyID,omitempty"`     // for type=credentials
    SecretKey     string `yaml:"secretKey,omitempty"`       // for type=credentials
    ProfilePath   string `yaml:"path,omitempty"`            // for type=profile
    ProfileName   string `yaml:"profile,omitempty"`         // for type=profile
    RoleToAssume  string `yaml:"role-to-assume,omitempty"`  // for type=iam
}

// LoadFile reads the YAML at given path and returns popualted Config
func LoadFile(path string, logger *slog.Logger) (*Config, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file %q: %w", path, err)
	}
	return Load(string(data), logger)
}

// Load unmarshals YAML into Config, applies basic validation
func Load(s string, logger *slog.Logger) (*Config, error) {
	var cfg Config

	err := yaml.UnmarshalStrict([]byte(s), &cfg); 
	if err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}
	
	// Region must be provided
	if cfg.AWS.Region == "" {
		return nil, fmt.Errorf("config.aws.region is required")
	}

	// Auth method must exist
	if len(cfg.AWS.Auth) == 0 {
		return nil, fmt.Errorf("config.aws.auth must have at least one method")
	}

	for i, m := range cfg.AWS.Auth {
		switch m.Type {
		case "credentials":
				// If credentials method was used the AccessKeyID and SecretKey are required
				if m.AccessKeyID == "" || m.SecretKey == "" {
						return nil, fmt.Errorf("auth[%d]: credentials type requires accessKeyID and secretKey", i)
				}
		case "profile":
				// If profile method was used the ProfilePath and ProfileName are required 
				if m.ProfilePath == "" || m.ProfileName == "" {
						return nil, fmt.Errorf("auth[%d]: profile type requires path and profile", i)
				}
		case "iam":
				// If IAM method was used RoleToAssume must be present
				if m.RoleToAssume == "" {
						return nil, fmt.Errorf("auth[%d]: iam type requires role-to-assume", i)
				}
				// No other types available
		default:
				return nil, fmt.Errorf("auth[%d]: unknown type %q", i, m.Type)
		}
}

    logger.Info("loaded config",
        "aws_region", cfg.AWS.Region,
        "auth_methods", len(cfg.AWS.Auth),
    )
    return &cfg, nil
}