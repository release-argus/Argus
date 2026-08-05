// Copyright [2026] [Argus]
//
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

// Package v1 provides the API for the webserver.
package v1

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/release-argus/Argus/auth/store"
	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/util"
	apitype "github.com/release-argus/Argus/web/api/types"
)

// marshalAnnouncePayload serialises WebSocket announce payloads (overridable for tests).
// see [decode.Marshal].
var marshalAnnouncePayload = func(v any) ([]byte, error) {
	return decode.Marshal("json", v)
}

// sendAnnouncePayload marshals a WebSocket message to JSON and publishes it on the announce channel.
// Marshal failures are logged and the message is dropped.
func (api *API) sendAnnouncePayload(msg apitype.WebSocketMessage) {
	payloadData, err := marshalAnnouncePayload(msg)
	if err != nil {
		logx.Error(err, logx.LogFrom{Primary: "API sendAnnouncePayload"}, true)
		return
	}
	api.Config.HardDefaults.Service.Status.AnnounceChannel <- payloadData
}

// getParam returns a query parameter value pointer if present, otherwise nil.
func getParam(queryParams url.Values, key string) *string {
	if !queryParams.Has(key) {
		return nil
	}

	val := queryParams.Get(key)
	return &val
}

// announceEdit broadcasts an EDIT message to all WebSocket clients.
//
// Only the fields the edit changed are sent; SubType carries the ID the service
// had before the edit (empty when it was created). An edit that changes nothing
// the dashboard displays still announces, with no ServiceData, so that clients
// invalidate their cached config for the service.
func (api *API) announceEdit(oldData *apitype.ServiceSummary, newData *apitype.ServiceSummary) {
	serviceChanged := ""
	if oldData != nil {
		serviceChanged = oldData.ID
		newData.RemoveUnchanged(oldData)
	}

	api.sendAnnouncePayload(
		apitype.WebSocketMessage{
			Page:        "APPROVALS",
			Type:        "EDIT",
			SubType:     serviceChanged,
			ServiceData: newData,
		},
	)
}

// announceDelete broadcasts a DELETE message to all WebSocket clients.
func (api *API) announceDelete(serviceID string) {
	api.sendAnnouncePayload(
		apitype.WebSocketMessage{
			Page:    "APPROVALS",
			Type:    "DELETE",
			SubType: serviceID,
		},
	)
}

// announceOrder broadcasts an ORDER message to all WebSocket clients.
func (api *API) announceOrder() {
	api.sendAnnouncePayload(
		apitype.WebSocketMessage{
			Page:    "APPROVALS",
			Type:    "SERVICE",
			SubType: "ORDER",
			Order:   &api.Config.Order,
		},
	)
}

// minPasswordLength is the minimum accepted password length.
const minPasswordLength = 8

// validatePassword reports whether plaintext meets the minimum password policy,
// returning a user-facing [error] when it does not.
func validatePassword(plaintext string) error {
	if len(plaintext) < minPasswordLength {
		return &decode.ErrField{
			Key:         "password",
			Value:       util.ValueUnlessZero(plaintext, "*"),
			Description: fmt.Sprintf("must be at least %d characters", minPasswordLength),
		}
	}
	return nil
}

// decodeAuthBody decodes a JSON request body into v,
// failing the request (and reporting false) on error.
func (api *API) decodeAuthBody(w http.ResponseWriter, r *http.Request, v any) bool {
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxAuthBodySize))
	if err != nil {
		failRequest(&w, fmt.Errorf("read request: %w", err), http.StatusBadRequest)
		return false
	}
	if err := decode.Unmarshal("json", body, v); err != nil {
		failRequest(&w, fmt.Errorf("parse request: %w", err), http.StatusBadRequest)
		return false
	}
	return true
}

// failAuthStoreRequest maps auth store errors onto HTTP statuses.
func (api *API) failAuthStoreRequest(w http.ResponseWriter, err error, logFrom logx.LogFrom, action string) {
	switch {
	case errors.Is(err, store.ErrNotFound):
		failRequest(&w, fmt.Errorf("%s failed: not found", action), http.StatusNotFound)
	case errors.Is(err, store.ErrUsernameTaken),
		errors.Is(err, store.ErrGroupNameTaken),
		errors.Is(err, store.ErrUnknownGroup),
		errors.Is(err, store.ErrInvalidGrant):
		failRequest(&w, fmt.Errorf("%s failed: %w", action, err), http.StatusBadRequest)
	case errors.Is(err, store.ErrLastAdmin),
		errors.Is(err, store.ErrSystemGroup):
		failRequest(&w, fmt.Errorf("%s failed: %w", action, err), http.StatusConflict)
	default:
		logx.Error(err, logFrom, true)
		failRequest(&w, fmt.Errorf("%s failed", action), http.StatusInternalServerError)
	}
}

// ConstantTimeCompare reports whether the arrays x and y have equal contents.
// The time taken depends only on the length of the arrays, not their contents.
func ConstantTimeCompare(x, y [32]byte) bool {
	var result byte

	for i := range len(x) {
		result |= x[i] ^ y[i]
	}

	return result == 0
}
