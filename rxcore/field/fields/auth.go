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
	// AuthSubject is the field key that records the authenticated subject
	// (for example, a user, service account, or principal).
	AuthSubject = "auth_subject"

	// AuthScope is the field key that records the granted scope or audience
	// associated with the authentication context.
	AuthScope = "auth_scope"

	// AuthRole is the field key that records the role or roles associated
	// with the subject (for example, "admin", "reader").
	AuthRole = "auth_role"

	// AuthMethod is the field key that records the authentication method
	// (for example, "mTLS", "OIDC", "basic").
	AuthMethod = "auth_method"

	// AuthTokenID is the field key that records an identifier for the
	// authentication token, if one is in use.
	AuthTokenID = "auth_token_id"
)
