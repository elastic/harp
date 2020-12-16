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

package routes

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/elastic/harp/pkg/sdk/log"
)

// -----------------------------------------------------------------------------

// Resource describes a JSON-LD resource header
type Resource struct {
	Context string `json:"@context,omitempty"`
	Type    string `json:"@type,omitempty"`
	ID      string `json:"@id,omitempty"`
}

// Status describe a resource to represent an operational status
type status struct {
	*Resource `json:",inline"`
	Code      int    `json:"code"`
	Message   string `json:"message"`
}

// -----------------------------------------------------------------------------

// with serializes the data with matching requested encoding
func with(w http.ResponseWriter, r *http.Request, code int, data interface{}) {
	// Marshal response as json
	js, err := json.Marshal(data)
	if err != nil {
		withError(w, r, err)
		return
	}

	// Set content type header
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	// Write status
	w.WriteHeader(code)

	// Write response
	_, err = w.Write(js)
	log.CheckErrCtx(r.Context(), "Unable to write response", err)
}

// WithError serialize an error
func withError(w http.ResponseWriter, r *http.Request, err interface{}) {
	switch errObj := err.(type) {
	case string:
		with(w, r, http.StatusBadRequest, &status{
			Resource: &Resource{
				Type: "Error",
			},
			Code:    http.StatusBadRequest,
			Message: errObj,
		})
	case error:
		with(w, r, http.StatusBadRequest, &status{
			Resource: &Resource{
				Type: "Error",
			},
			Code:    http.StatusBadRequest,
			Message: errObj.Error(),
		})
	default:
		with(w, r, http.StatusInternalServerError, &status{
			Resource: &Resource{
				Type: "Error",
			},
			Code:    http.StatusInternalServerError,
			Message: "Unable to process this request",
		})
	}
}

type malformedRequest struct {
	status int
	msg    string
}

func (mr *malformedRequest) Error() string {
	return mr.msg
}

func decodeJSONBody(w http.ResponseWriter, r *http.Request, dst interface{}) error {
	// Limit body size to 1Mb
	r.Body = http.MaxBytesReader(w, r.Body, 1048576)

	// Disallow unknown fields deserialization
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	err := dec.Decode(&dst)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError

		switch {
		case errors.As(err, &syntaxError):
			msg := fmt.Sprintf("Request body contains badly-formed JSON (at position %d)", syntaxError.Offset)
			return &malformedRequest{status: http.StatusBadRequest, msg: msg}

		case errors.Is(err, io.ErrUnexpectedEOF):
			msg := "Request body contains badly-formed JSON"
			return &malformedRequest{status: http.StatusBadRequest, msg: msg}

		case errors.As(err, &unmarshalTypeError):
			msg := fmt.Sprintf("Request body contains an invalid value for the %q field (at position %d)", unmarshalTypeError.Field, unmarshalTypeError.Offset)
			return &malformedRequest{status: http.StatusBadRequest, msg: msg}

		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			msg := fmt.Sprintf("Request body contains unknown field %s", fieldName)
			return &malformedRequest{status: http.StatusBadRequest, msg: msg}

		case errors.Is(err, io.EOF):
			msg := "Request body must not be empty"
			return &malformedRequest{status: http.StatusBadRequest, msg: msg}

		case err.Error() == "http: request body too large":
			msg := "Request body must not be larger than 1MB"
			return &malformedRequest{status: http.StatusRequestEntityTooLarge, msg: msg}

		default:
			return err
		}
	}

	if dec.More() {
		msg := "Request body must only contain a single JSON object"
		return &malformedRequest{status: http.StatusBadRequest, msg: msg}
	}

	return nil
}
