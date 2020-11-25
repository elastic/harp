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
	"net/http"

	jsoniter "github.com/json-iterator/go"

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
	json := jsoniter.ConfigCompatibleWithStandardLibrary

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
