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

package httpctl

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"dirpx.dev/rxlog/rxapi/level"
	"dirpx.dev/rxlog/rxcore/field"
)

const (
	// DefaultMaxBodyBytes is the fallback limit used when MaxBodyBytes is
	// not set explicitly. It prevents unbounded reads from the request body.
	DefaultMaxBodyBytes int64 = 4 << 10 // 4 KiB

	// DefaultLevelParamName is the default query/body parameter name used
	// to carry the level string when no explicit ParamName is configured.
	DefaultLevelParamName string = field.Level
)

const (
	// HeaderContentType is the HTTP header name used to describe the media
	// type of the response body.
	HeaderContentType = "Content-Type"

	// HeaderAllow is the HTTP header name used to advertise which methods
	// are allowed on this endpoint.
	HeaderAllow = "Allow"

	// HeaderCacheControl is the HTTP header name used to control caching
	// behavior of responses.
	HeaderCacheControl = "Cache-Control"
)

const (
	// CacheControlNoStore is the Cache-Control directive used to prevent
	// intermediaries from caching introspection responses.
	CacheControlNoStore = "no-store"

	// ContentTypeJSON is the JSON media type used for all normal responses
	// and error payloads.
	ContentTypeJSON = "application/json; charset=utf-8"

	// ContentTypeTextPlain is the plain-text media type used for generic
	// internal error responses emitted by the outer ServeHTTP wrapper.
	ContentTypeTextPlain = "text/plain; charset=utf-8"
)

// Handler exposes and mutates a mutable log level threshold over HTTP.
//
// It is intended for operational use: operators can inspect the current
// log level and change it at runtime using a small, well-defined HTTP
// protocol.
//
// Concurrency and safety:
//   - All operations delegate to a level.Threshold implementation, which
//     MUST be safe for concurrent use by multiple goroutines.
//   - Handler itself holds only immutable configuration and a
//     reference to the Threshold, so it is also safe for concurrent use.
//
// Semantics:
//   - GET /…: returns the current log level as JSON,
//     for example, {"level":"info"}.
//   - HEAD /…: same as GET, but without a response body (headers only).
//   - POST/PUT/PATCH /…: parses a desired level from the request and
//     updates the underlying Threshold, then returns {"level":"<new>"}.
//
// Level source for POST/PUT/PATCH (in order of precedence):
//  1. Query parameter: ?<ParamName>=<level>, if non-empty.
//  2. JSON body (Content-Type: application/json):
//     {"<ParamName>": "info"}.
//  3. Plain-text body: entire body treated as a single level string.
//
// On success, mutation requests respond with HTTP 200 and the new level
// as JSON in the response body.
//
// On client error (for example, invalid level string, malformed JSON),
// the handler responds with HTTP 400 and a JSON error payload whose
// field name is derived from field.Error.
//
// The JSON keys for the current level and error messages are taken from
// field.Level and field.Error respectively, so they stay consistent
// with the global logging schema.
//
// Misconfiguration (for example, a nil Level reference) or unexpected
// internal failures are signaled by returning a non-nil error from
// serveHTTP; the outer ServeHTTP wrapper translates that into HTTP 500
// with a simple text/plain diagnostic.
//
// This handler deliberately does not implement any authentication or
// authorization. Callers MUST mount it behind an appropriate access
// control mechanism (for example, an internal admin port or reverse
// proxy with authentication).
type Handler struct {
	// Level is the mutable log level threshold being observed and mutated.
	//
	// This value MUST be non-nil. If it is nil at request time, the
	// handler treats it as a server-side misconfiguration and the outer
	// ServeHTTP wrapper returns HTTP 500.
	Level level.Threshold

	// ParamName is the name of the query/body field used to carry the
	// level string.
	//
	// If empty, DefaultLevelParamName is used.
	ParamName string

	// MaxBodyBytes limits how many bytes the handler reads from the
	// request body when parsing a new level.
	//
	// This applies both to JSON bodies and plain-text bodies and is
	// intended as a simple guardrail against pathological requests.
	//
	// If zero or negative, DefaultMaxBodyBytes is used.
	MaxBodyBytes int64

	// ReadOnly, if true, disables mutation methods (POST/PUT/PATCH).
	//
	// When ReadOnly is set, any POST/PUT/PATCH requests receive HTTP 405
	// Method Not Allowed, while GET/HEAD remain available.
	ReadOnly bool
}

// NewHandler constructs an Handler bound to the provided mutable
// threshold.
//
// The returned handler is safe for concurrent use as long as the supplied
// level.Threshold implementation is safe for concurrent use.
//
// Example usage:
//
//	lvl := rxclvl.NewAtomicLevel() // implements level.Threshold
//	mux.Handle("/debug/log-level", httpctl.NewHandler(&lvl))
//
// This will expose:
//   - GET  /debug/log-level -> {"level":"info"}
//   - PUT  /debug/log-level?level=error -> {"level":"error"}.
//   - POST /debug/log-level with JSON/plain payload will behave similarly.
func NewHandler(th level.Threshold, opts ...Option) *Handler {
	h := &Handler{
		Level:        th,
		ParamName:    DefaultLevelParamName,
		MaxBodyBytes: DefaultMaxBodyBytes,
	}

	for _, opt := range opts {
		opt(h)
	}

	if h.ParamName == "" {
		h.ParamName = DefaultLevelParamName
	}
	if h.MaxBodyBytes <= 0 {
		h.MaxBodyBytes = DefaultMaxBodyBytes
	}

	return h
}

// ServeHTTP implements http.Handler.
//
// It delegates all request handling to the internal serveHTTP method.
// If serveHTTP returns a non-nil error, ServeHTTP responds with HTTP 500
// and a simple text/plain diagnostic.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := h.serveHTTP(w, r); err != nil {
		w.Header().Set(HeaderContentType, ContentTypeTextPlain)
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = fmt.Fprintf(w, "internal error: %v", err)
	}
}

// serveHTTP contains the actual request handling logic.
// It MUST write status codes and JSON bodies for all expected client
// conditions (success, 4xx, 405) and return nil in those cases.
//
// It SHOULD return a non-nil error only for internal failures
// (for example, nil Level, encoding errors). The outer ServeHTTP wrapper
// translates such errors into HTTP 500 responses.
func (h *Handler) serveHTTP(w http.ResponseWriter, r *http.Request) error {
	if h.Level == nil {
		// Internal misconfiguration → bubble up, ServeHTTP will emit 500.
		return fmt.Errorf("log level handler not initialized")
	}

	enc := json.NewEncoder(w)

	switch r.Method {
	case http.MethodGet:
		// Return current level as JSON: { "<field.Level>": "info" }.
		w.Header().Set(HeaderCacheControl, CacheControlNoStore)
		return writeJSONLevel(w, enc, http.StatusOK, h.Level.Level().String())

	case http.MethodHead:
		// HEAD: same headers as GET, but no body.
		w.Header().Set(HeaderContentType, ContentTypeJSON)
		w.Header().Set(HeaderCacheControl, CacheControlNoStore)
		w.WriteHeader(http.StatusOK)
		return nil

	case http.MethodPost, http.MethodPut, http.MethodPatch:
		if r.Body != nil {
			defer r.Body.Close()
		}

		if h.ReadOnly {
			w.Header().Set(HeaderAllow, "GET, HEAD")
			return writeJSONError(w, enc, http.StatusMethodNotAllowed,
				"log level handler is read-only")
		}

		// 1) Query parameter, if present and non-empty.
		levelSpec := ""
		if h.ParamName != "" {
			if v := strings.TrimSpace(r.URL.Query().Get(h.ParamName)); v != "" {
				levelSpec = v
			}
		}

		// 2) If query is empty, fall back to body (JSON or plain text).
		if levelSpec == "" {
			spec, err := h.readLevelFromBody(r)
			if err != nil {
				return writeJSONError(w, enc, http.StatusBadRequest, err.Error())
			}
			levelSpec = spec
		}

		if levelSpec == "" {
			return writeJSONError(w, enc, http.StatusBadRequest, "missing level value")
		}

		// Parse the level string.
		l, err := level.Parse(levelSpec)
		if err != nil {
			return writeJSONError(w, enc, http.StatusBadRequest, err.Error())
		}

		// Apply the new level.
		h.Level.SetLevel(l)

		return writeJSONLevel(w, enc, http.StatusOK, h.Level.Level().String())

	default:
		w.Header().Set(HeaderAllow, "GET, HEAD, POST, PUT, PATCH")
		return writeJSONError(w, enc, http.StatusMethodNotAllowed,
			"only GET, HEAD, POST, PUT, PATCH are supported")
	}
}

// writeJSONLevel writes a JSON response containing the current level
// under the field key defined by field.Level.
//
// It sets the Content-Type header to ContentTypeJSON, writes the given
// HTTP status code, and then encodes a single-object payload like:
//
//	{ "<field.Level>": "<level>" }
//
// The encoder is provided by the caller so that a single json.Encoder
// can be reused across multiple responses within serveHTTP if desired.
func writeJSONLevel(
	w http.ResponseWriter,
	enc *json.Encoder,
	status int,
	levelValue string,
) error {
	w.Header().Set(HeaderContentType, ContentTypeJSON)
	if status != 0 {
		w.WriteHeader(status)
	}

	return enc.Encode(map[string]string{
		field.Level: levelValue,
	})
}

// writeJSONError writes a JSON error response using the field key defined
// by field.Error.
//
// It sets the Content-Type header to ContentTypeJSON, writes the given
// HTTP status code, and then encodes a single-object payload like:
//
//	{ "<field.Error>": "<message>" }
//
// This helper centralizes the error response shape so that all error
// payloads produced by Handler are consistent with the global field
// schema.
func writeJSONError(
	w http.ResponseWriter,
	enc *json.Encoder,
	status int,
	message string,
) error {
	w.Header().Set(HeaderContentType, ContentTypeJSON)
	if status != 0 {
		w.WriteHeader(status)
	}

	return enc.Encode(map[string]string{
		field.Error: message,
	})
}

// readLevelFromBody reads a level string from the request body.
//
// If Content-Type starts with "application/json", it expects a JSON
// object of the form {"<ParamName>": "info"}.
//
// For all other Content-Types (or when Content-Type is absent), it
// treats the body as plain text and uses the entire body as the level
// string after trimming whitespace.
//
// Returns:
//   - the level string (which may be empty if the body is empty);
//   - an error for client failures (bad JSON, etc.), which callers
//     translate into HTTP 400.
func (h *Handler) readLevelFromBody(r *http.Request) (string, error) {
	if r.Body == nil {
		return "", nil
	}

	limit := h.MaxBodyBytes
	if limit <= 0 {
		limit = DefaultMaxBodyBytes
	}
	reader := io.LimitReader(r.Body, limit)

	ct := r.Header.Get(HeaderContentType)
	if strings.HasPrefix(strings.ToLower(ct), "application/json") {
		return h.readLevelFromJSON(reader)
	}

	// Fallback: plain text.
	b, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("read plain-text body: %w", err)
	}
	return strings.TrimSpace(string(b)), nil
}

// readLevelFromJSON reads {"<ParamName>":"<level>"} from a JSON body.
func (h *Handler) readLevelFromJSON(r io.Reader) (string, error) {
	if h.ParamName == "" {
		return "", fmt.Errorf("param name is empty for JSON body")
	}

	dec := json.NewDecoder(r)

	var payload map[string]any
	if err := dec.Decode(&payload); err != nil {
		return "", fmt.Errorf("decode JSON: %w", err)
	}

	raw, ok := payload[h.ParamName]
	if !ok {
		return "", fmt.Errorf("missing %q field in JSON body", h.ParamName)
	}

	s, ok := raw.(string)
	if !ok {
		return "", fmt.Errorf("field %q must be a string", h.ParamName)
	}

	return strings.TrimSpace(s), nil
}
