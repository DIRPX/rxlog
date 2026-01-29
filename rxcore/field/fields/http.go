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

package fields

const (
	// HTTPMethod is the field key that records the HTTP request method
	// (for example, "GET", "POST").
	HTTPMethod = "http_method"

	// HTTPURL is the field key that records the full HTTP URL as received
	// by the server.
	HTTPURL = "http_url"

	// HTTPPath is the field key that records only the path component of
	// the HTTP URL.
	HTTPPath = "http_path"

	// HTTPRoute is the field key that records the logical route or pattern
	// that handled the request (for example, "/users/{id}").
	HTTPRoute = "http_route"

	// HTTPHost is the field key that records the Host header or authority
	// component of the HTTP request.
	HTTPHost = "http_host"

	// HTTPScheme is the field key that records the URL scheme ("http" or
	// "https").
	HTTPScheme = "http_scheme"

	// HTTPProto is the field key that records the HTTP protocol version
	// (for example, "HTTP/1.1", "HTTP/2").
	HTTPProto = "http_proto"

	// HTTPStatusCode is the field key that records the HTTP response status
	// code (for example, 200, 404, 500).
	HTTPStatusCode = "http_status"

	// HTTPUserAgent is the field key that records the User-Agent header
	// from the HTTP request.
	HTTPUserAgent = "user_agent"

	// HTTPReferer is the field key that records the Referer (or Referrer)
	// header from the HTTP request.
	HTTPReferer = "referer"

	// HTTPRemoteAddr is the field key that records the remote network address
	// of the HTTP client as observed by the server.
	HTTPRemoteAddr = "remote_addr"

	// HTTPClientIP is the field key that records the originating client IP
	// address, taking into account X-Forwarded-For or similar headers.
	HTTPClientIP = "client_ip"

	// HTTPRequestSize is the field key that records the size of the HTTP
	// request payload in bytes.
	HTTPRequestSize = "http_request_size"

	// HTTPResponseSize is the field key that records the size of the HTTP
	// response payload in bytes.
	HTTPResponseSize = "http_response_size"
)
