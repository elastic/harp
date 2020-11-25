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
	"io"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var defaultFactory LoggerFactory

// -----------------------------------------------------------------------------

func init() {
	SetLoggerFactory(NewFactory(zap.L()))
}

// SetLoggerFactory defines the default package logger
func SetLoggerFactory(instance LoggerFactory) {
	if defaultFactory != nil {
		defaultFactory.Bg().Debug("Replacing logger factory", zap.String("old", defaultFactory.Name()), zap.String("new", instance.Name()))
	} else {
		instance.Bg().Debug("Initializing logger factory", zap.String("factory", instance.Name()))
	}
	defaultFactory = instance
}

// -----------------------------------------------------------------------------

// Bg delegates a no-context logger
func Bg() Logger {
	return checkFactory(defaultFactory).Bg()
}

// For delegates a context logger
func For(ctx context.Context) Logger {
	return checkFactory(defaultFactory).For(ctx)
}

// Default returns the logger factory
func Default() LoggerFactory {
	return checkFactory(defaultFactory)
}

// CheckErr handles error correctly
func CheckErr(msg string, err error, fields ...zapcore.Field) {
	if err != nil {
		fields = append(fields, zap.Error(err))
		Default().Bg().Error(msg, fields...)
	}
}

// CheckErrCtx handles error correctly
func CheckErrCtx(ctx context.Context, msg string, err error, fields ...zapcore.Field) {
	if err != nil {
		fields = append(fields, zap.Error(errors.WithStack(err)))
		Default().For(ctx).Error(msg, fields...)
	}
}

// SafeClose handles the closer error
func SafeClose(c io.Closer, msg string, fields ...zapcore.Field) {
	if cerr := c.Close(); cerr != nil {
		fields = append(fields, zap.Error(errors.WithStack(cerr)))
		Default().Bg().Error(msg, fields...)
	}
}

// SafeCloseCtx handles the closer error
func SafeCloseCtx(ctx context.Context, c io.Closer, msg string, fields ...zapcore.Field) {
	if cerr := c.Close(); cerr != nil {
		fields = append(fields, zap.Error(errors.WithStack(cerr)))
		Default().For(ctx).Error(msg, fields...)
	}
}

// -----------------------------------------------------------------------------

func checkFactory(defaultFactory LoggerFactory) LoggerFactory {
	if defaultFactory == nil {
		panic("Unable to create logger instance, you have to register an adapter first.")
	}
	return defaultFactory
}
