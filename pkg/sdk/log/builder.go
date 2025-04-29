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

package log

import (
	"context"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// -----------------------------------------------------------------------------

// Options declares logger options for builder
type Options struct {
	Debug     bool
	LogLevel  string
	AppName   string
	AppID     string
	Version   string
	Revision  string
	SentryDSN string
}

// -----------------------------------------------------------------------------

// DefaultOptions defines default logger options
var DefaultOptions = &Options{
	Debug:     false,
	LogLevel:  "info",
	AppName:   "changeme",
	AppID:     "changeme",
	Version:   "0.0.1",
	Revision:  "123456789",
	SentryDSN: "",
}

// -----------------------------------------------------------------------------

// Setup the logger
//
//nolint:revive // refactor use of ctx
func Setup(ctx context.Context, opts *Options) {
	// Initialize logs
	var config zap.Config

	if opts.Debug {
		opts.LogLevel = "debug"
		config = zap.NewDevelopmentConfig()
		config.DisableCaller = true
		config.DisableStacktrace = true
	} else {
		config = zap.NewProductionConfig()
		config.DisableStacktrace = true
		config.EncoderConfig.MessageKey = "@message"
		config.EncoderConfig.TimeKey = "@timestamp"
		config.EncoderConfig.CallerKey = "@caller"
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	}

	// Parse log level
	errLogLevel := config.Level.UnmarshalText([]byte(opts.LogLevel))
	if errLogLevel != nil {
		panic(errLogLevel)
	}

	// Build real logger
	logger, err := config.Build(
		zap.AddCallerSkip(2),
	)
	if err != nil {
		panic(err)
	}

	// Add prefix to logger
	logger = logger.With(
		zap.String("@appName", opts.AppName),
		zap.String("@version", opts.Version),
		zap.String("@revision", opts.Revision),
		zap.String("@appID", opts.AppID),
		zap.Namespace("@fields"),
	)

	// Prepare factory
	logFactory := NewFactory(logger)

	// Override the global factory
	SetLoggerFactory(logFactory)

	// Override zap default logger
	zap.ReplaceGlobals(logger)
}
