/*
   Copyright 2025 The DIRPX Authors.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package httpctl_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	httpctl "dirpx.dev/rxlog/levelctl/http"
	"dirpx.dev/rxlog/rxapi/level"
	"dirpx.dev/rxlog/rxcore/field/fields"
	rxclvl "dirpx.dev/rxlog/rxcore/level"
)

// newTestHandler constructs an Handler bound to an AtomicLevel
// initialized to the provided level.
//
// It returns both the handler and the underlying AtomicLevel so tests
// can assert on state changes.
func newTestHandler(t *testing.T, initial level.Level, opts ...httpctl.Option) (*httpctl.Handler, *rxclvl.AtomicLevel) {
	t.Helper()

	atomicLvl := rxclvl.NewAtomicLevel()
	atomicLvl.SetLevel(initial)

	h := httpctl.NewHandler(&atomicLvl, opts...)
	return h, &atomicLvl
}

// decodeJSONMap decodes the response body into a map and closes the body.
func decodeJSONMap(t *testing.T, body io.ReadCloser) map[string]string {
	t.Helper()
	defer body.Close()

	var m map[string]string
	if err := json.NewDecoder(body).Decode(&m); err != nil {
		t.Fatalf("failed to decode JSON body: %v", err)
	}
	return m
}

func TestHandler_GetReturnsCurrentLevel(t *testing.T) {
	h, _ := newTestHandler(t, level.Info)

	req := httptest.NewRequest(http.MethodGet, "/debug/log-level", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	res := rec.Result()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, res.StatusCode)
	}
	if ct := res.Header.Get(httpctl.HeaderContentType); ct != httpctl.ContentTypeJSON {
		t.Fatalf("expected Content-Type %q, got %q", httpctl.ContentTypeJSON, ct)
	}
	if cc := res.Header.Get(httpctl.HeaderCacheControl); cc != httpctl.CacheControlNoStore {
		t.Fatalf("expected Cache-Control %q, got %q", httpctl.CacheControlNoStore, cc)
	}

	m := decodeJSONMap(t, res.Body)

	levelValue, ok := m[fields.Level]
	if !ok {
		t.Fatalf("response JSON is missing %q field: %#v", fields.Level, m)
	}
	if levelValue != level.Info.String() {
		t.Fatalf("expected level %q, got %q", level.Info.String(), levelValue)
	}
}

func TestHandler_HeadReturnsNoBody(t *testing.T) {
	h, _ := newTestHandler(t, level.Warn)

	req := httptest.NewRequest(http.MethodHead, "/debug/log-level", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, res.StatusCode)
	}
	if ct := res.Header.Get(httpctl.HeaderContentType); ct != httpctl.ContentTypeJSON {
		t.Fatalf("expected Content-Type %q, got %q", httpctl.ContentTypeJSON, ct)
	}

	// HEAD MUST NOT return a body; io.ReadAll should see EOF immediately.
	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("unexpected error reading HEAD body: %v", err)
	}
	if len(body) != 0 {
		t.Fatalf("expected empty body for HEAD, got %q", string(body))
	}
}

func TestHandler_PutViaQueryUpdatesLevel(t *testing.T) {
	h, atomicLvl := newTestHandler(t, level.Info)

	req := httptest.NewRequest(http.MethodPut, "/debug/log-level?level=error", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	res := rec.Result()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, res.StatusCode)
	}
	m := decodeJSONMap(t, res.Body)

	levelValue, ok := m[fields.Level]
	if !ok {
		t.Fatalf("response JSON is missing %q field: %#v", fields.Level, m)
	}
	if levelValue != level.Error.String() {
		t.Fatalf("expected response level %q, got %q", level.Error.String(), levelValue)
	}
	if got := atomicLvl.Level(); got != level.Error {
		t.Fatalf("expected AtomicLevel %v, got %v", level.Error, got)
	}
}

func TestHandler_PostJSONUpdatesLevel(t *testing.T) {
	h, atomicLvl := newTestHandler(t, level.Warn)

	body := `{"` + fields.Level + `":"info"}`
	req := httptest.NewRequest(http.MethodPost, "/debug/log-level", strings.NewReader(body))
	req.Header.Set(httpctl.HeaderContentType, "application/json")
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	res := rec.Result()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, res.StatusCode)
	}
	m := decodeJSONMap(t, res.Body)

	levelValue, ok := m[fields.Level]
	if !ok {
		t.Fatalf("response JSON is missing %q field: %#v", fields.Level, m)
	}
	if levelValue != level.Info.String() {
		t.Fatalf("expected response level %q, got %q", level.Info.String(), levelValue)
	}
	if got := atomicLvl.Level(); got != level.Info {
		t.Fatalf("expected AtomicLevel %v, got %v", level.Info, got)
	}
}

func TestHandler_PostPlainTextUpdatesLevel(t *testing.T) {
	h, atomicLvl := newTestHandler(t, level.Info)

	req := httptest.NewRequest(http.MethodPost, "/debug/log-level", strings.NewReader("warn\n"))
	// No Content-Type header → treated as plain text.
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	res := rec.Result()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, res.StatusCode)
	}
	m := decodeJSONMap(t, res.Body)

	levelValue, ok := m[fields.Level]
	if !ok {
		t.Fatalf("response JSON is missing %q field: %#v", fields.Level, m)
	}
	if levelValue != level.Warn.String() {
		t.Fatalf("expected response level %q, got %q", level.Warn.String(), levelValue)
	}
	if got := atomicLvl.Level(); got != level.Warn {
		t.Fatalf("expected AtomicLevel %v, got %v", level.Warn, got)
	}
}

func TestHandler_ReadOnlyBlocksMutation(t *testing.T) {
	h, atomicLvl := newTestHandler(t, level.Info, httpctl.WithReadOnly(true))

	req := httptest.NewRequest(http.MethodPost, "/debug/log-level?level=error", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	res := rec.Result()
	if res.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("expected status %d, got %d", http.StatusMethodNotAllowed, res.StatusCode)
	}
	m := decodeJSONMap(t, res.Body)

	msg, ok := m[fields.Error]
	if !ok {
		t.Fatalf("response JSON is missing %q field: %#v", fields.Error, m)
	}
	if msg == "" {
		t.Fatalf("expected non-empty error message")
	}
	if got := atomicLvl.Level(); got != level.Info {
		t.Fatalf("expected AtomicLevel to remain %v, got %v", level.Info, got)
	}
}

func TestHandler_InvalidJSONReturns400(t *testing.T) {
	h, _ := newTestHandler(t, level.Info)

	req := httptest.NewRequest(http.MethodPost, "/debug/log-level", strings.NewReader("{"))
	req.Header.Set(httpctl.HeaderContentType, "application/json")
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	res := rec.Result()
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, res.StatusCode)
	}
	m := decodeJSONMap(t, res.Body)

	if _, ok := m[fields.Error]; !ok {
		t.Fatalf("response JSON is missing %q field: %#v", fields.Error, m)
	}
}

func TestHandler_InvalidLevelReturns400(t *testing.T) {
	h, _ := newTestHandler(t, level.Info)

	req := httptest.NewRequest(http.MethodPost, "/debug/log-level", strings.NewReader("not-a-level"))
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	res := rec.Result()
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, res.StatusCode)
	}
	m := decodeJSONMap(t, res.Body)

	if _, ok := m[fields.Error]; !ok {
		t.Fatalf("response JSON is missing %q field: %#v", fields.Error, m)
	}
}

func TestHandler_MissingLevelReturns400(t *testing.T) {
	h, _ := newTestHandler(t, level.Info)

	// Empty body and no query parameter → missing level.
	req := httptest.NewRequest(http.MethodPost, "/debug/log-level", strings.NewReader(""))
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	res := rec.Result()
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, res.StatusCode)
	}
	m := decodeJSONMap(t, res.Body)

	if _, ok := m[fields.Error]; !ok {
		t.Fatalf("response JSON is missing %q field: %#v", fields.Error, m)
	}
}

func TestHandler_NilLevelProduces500(t *testing.T) {
	// Construct a handler with a nil Level to exercise the internal
	// error path and the outer ServeHTTP wrapper.
	h := &httpctl.Handler{
		Level:        nil,
		ParamName:    httpctl.DefaultLevelParamName,
		MaxBodyBytes: httpctl.DefaultMaxBodyBytes,
	}

	req := httptest.NewRequest(http.MethodGet, "/debug/log-level", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, res.StatusCode)
	}
	if ct := res.Header.Get(httpctl.HeaderContentType); ct != httpctl.ContentTypeTextPlain {
		t.Fatalf("expected Content-Type %q, got %q", httpctl.ContentTypeTextPlain, ct)
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("failed to read body: %v", err)
	}
	if !strings.Contains(string(body), "internal error") {
		t.Fatalf("expected body to contain %q, got %q", "internal error", string(body))
	}
}
