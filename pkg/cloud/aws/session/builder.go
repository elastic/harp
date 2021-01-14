// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package session

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
)

// NewSession creates an AWS session to use AWS Services or compatible.
func NewSession(opts *Options) (*session.Session, error) {
	// Check arguments
	if opts == nil {
		return nil, errors.New("unable to build without options")
	}
	if opts.Region == "" {
		opts.Region = "us-east-1"
	}

	// Start a new AWS session
	awsSession, err := session.NewSession()
	if err != nil {
		return nil, fmt.Errorf("unable to initialize AWS session: %w", err)
	}

	// Prepare credential providers
	providers := []credentials.Provider{}
	if opts.AccessKeyID != "" && opts.SecretAccessKey != "" {
		providers = append(providers, &credentials.StaticProvider{
			Value: credentials.Value{
				AccessKeyID:     opts.AccessKeyID,
				SecretAccessKey: opts.SecretAccessKey,
				SessionToken:    opts.SessionToken,
			},
		})
	}
	if !opts.IgnoreEnvCreds {
		providers = append(providers, &credentials.EnvProvider{})
	}
	if !opts.IgnoreConfigCreds {
		providers = append(providers, &credentials.SharedCredentialsProvider{})
	}
	if !opts.IgnoreEC2RoleCreds {
		providers = append(providers, &ec2rolecreds.EC2RoleProvider{
			Client: ec2metadata.New(awsSession, &aws.Config{
				HTTPClient: &http.Client{Timeout: 1 * time.Second},
			}),
			ExpiryWindow: 2 * time.Minute,
		})
	}

	// Assemble credentials
	creds := credentials.NewChainCredentials(providers)

	// Prepare config
	config := aws.Config{
		Credentials:               creds,
		DisableSSL:                aws.Bool(opts.DisableSSL),
		S3ForcePathStyle:          aws.Bool(opts.S3ForcePathStyle),
		S3UseAccelerate:           aws.Bool(opts.UseAccelerateEndpoint),
		S3UsEast1RegionalEndpoint: endpoints.RegionalS3UsEast1Endpoint,
	}
	if opts.Endpoint != "" {
		config.Endpoint = aws.String(opts.Endpoint)
	}
	if opts.Region != "" {
		config.Region = aws.String(opts.Region)
	}

	// Prepare options
	awsSessionOpts := session.Options{
		Config: config,
	}
	if opts.EnvAuthentication && opts.AccessKeyID == "" && opts.SecretAccessKey == "" {
		awsSessionOpts.SharedConfigState = session.SharedConfigEnable
		awsSessionOpts.Config.Credentials = nil
	}
	if opts.Profile != "" {
		awsSessionOpts.Profile = opts.Profile
	}

	// Build session
	return session.NewSessionWithOptions(awsSessionOpts)
}

// FromURL parses the given url to build a session object.
func FromURL(u string) (*Options, error) {
	// Parse input as URL
	input, err := url.Parse(u)
	if err != nil {
		return nil, fmt.Errorf("unable to build session from url")
	}

	// Extract path and object key
	bucketName, objectKey := filepath.Split(input.Path)
	if bucketName == "" {
		return nil, fmt.Errorf("bucketName is mandatory")
	}
	if objectKey == "" {
		return nil, fmt.Errorf("objectKey is mandatory")
	}

	q := input.Query()

	// Unescape strings
	accessKeyID, err := url.QueryUnescape(q.Get("access-key-id"))
	if err != nil {
		return nil, fmt.Errorf("accessKeyID value is invalid: %w", err)
	}
	profile, err := url.QueryUnescape(q.Get("profile"))
	if err != nil {
		return nil, fmt.Errorf("profile value is invalid: %w", err)
	}
	region, err := url.QueryUnescape(q.Get("region"))
	if err != nil {
		return nil, fmt.Errorf("region value is invalid: %w", err)
	}
	secretAccessKey, err := url.QueryUnescape(q.Get("secret-access-key"))
	if err != nil {
		return nil, fmt.Errorf("secret-access-key value is invalid: %w", err)
	}
	sessionToken, err := url.QueryUnescape(q.Get("session-token"))
	if err != nil {
		return nil, fmt.Errorf("session-token value is invalid: %w", err)
	}

	const trueStringValue = "true"

	// Assemble options
	opts := &Options{
		Endpoint:   input.Host,
		BucketName: strings.TrimSuffix(strings.TrimPrefix(bucketName, "/"), "/"),
		ObjectKey:  objectKey,
		// Query params
		AccessKeyID:           accessKeyID,
		DisableSSL:            q.Get("disable-ssl") == trueStringValue,
		EnvAuthentication:     q.Get("env-authentication") == trueStringValue,
		IgnoreConfigCreds:     q.Get("ignore-config-creds") == trueStringValue,
		IgnoreEC2RoleCreds:    q.Get("ignore-ec2role-creds") == trueStringValue,
		IgnoreEnvCreds:        q.Get("ignore-env-creds") == trueStringValue,
		Profile:               profile,
		Region:                region,
		S3ForcePathStyle:      q.Get("s3-force-path-style") == trueStringValue,
		SecretAccessKey:       secretAccessKey,
		SessionToken:          sessionToken,
		UseAccelerateEndpoint: q.Get("s3-use-accelerate-endpoint") == trueStringValue,
	}

	// Delegate to builder
	return opts, nil
}
