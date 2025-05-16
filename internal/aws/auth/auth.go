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
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"

	"github.com/andranikasd/Zeno/config"
)

// NewSession establishes an AWS session using the first valid method in cfg.Auth.
// Supported types are "credentials", "profile", and "iam".
func NewSession(cfg *config.AWSConfig) (*session.Session, error) {
	var lastErr error

	for _, auth := range cfg.Auth {
		switch auth.Type {
		case "credentials":
			sess, err := session.NewSession(&aws.Config{
				Region:      aws.String(cfg.Region),
				Credentials: credentials.NewStaticCredentials(auth.AccessKeyID, auth.SecretKey, ""),
			})
			if err != nil {
				lastErr = fmt.Errorf("credentials session error: %w", err)
				continue
			}

			if _, err := sess.Config.Credentials.Get(); err != nil {
				lastErr = fmt.Errorf("credentials validation error: %w", err)
				continue
			}

			return sess, nil

		case "profile":
			sess, err := session.NewSessionWithOptions(session.Options{
				Config:            aws.Config{Region: aws.String(cfg.Region)},
				Profile:           auth.ProfileName,
				SharedConfigState: session.SharedConfigEnable,
			})
			if err != nil {
				lastErr = fmt.Errorf("profile session error: %w", err)
				continue
			}

			if _, err := sess.Config.Credentials.Get(); err != nil {
				lastErr = fmt.Errorf("profile validation error: %w", err)
				continue
			}

			return sess, nil

		case "iam":
			base, err := session.NewSession(&aws.Config{Region: aws.String(cfg.Region)})
			if err != nil {
				lastErr = fmt.Errorf("iam base session error: %w", err)
				continue
			}

			// Assume role
			svc := sts.New(base)
			creds := stscreds.NewCredentialsWithClient(svc, auth.RoleToAssume, func(p *stscreds.AssumeRoleProvider) {
				p.Duration = 15 * time.Minute
			})

			sess, err := session.NewSession(&aws.Config{
				Region:      aws.String(cfg.Region),
				Credentials: creds,
			})
			if err != nil {
				lastErr = fmt.Errorf("iam session error: %w", err)
				continue
			}

			if _, err := sess.Config.Credentials.Get(); err != nil {
				lastErr = fmt.Errorf("iam validation error: %w", err)
				continue
			}

			return sess, nil

		default:
			lastErr = fmt.Errorf("unknown auth type %q", auth.Type)
		}
	}

	return nil, fmt.Errorf("no valid AWS auth method: %w", lastErr)
}