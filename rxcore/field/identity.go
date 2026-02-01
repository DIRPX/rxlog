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

package field

const (
	// SessionID is the field key that identifies a logical user or client
	// session.
	//
	// This may correspond to a web session, a long-lived connection, or any
	// other notion of session in the application. Producers MAY populate this
	// field to support correlation of multiple requests or events that belong
	// to the same session context.
	SessionID = "session_id"

	// UserID is the field key that identifies the authenticated user
	// associated with the log entry, if any.
	//
	// The exact format (numeric ID, UUID, email, etc.) is application-defined.
	// Producers SHOULD avoid placing highly sensitive identifiers here and
	// MUST follow local privacy and security policies when logging user data.
	UserID = "user_id"

	// TenantID is the field key that identifies the tenant in multi-tenant
	// systems.
	//
	// Producers SHOULD use this to separate logs from different tenants and
	// to support per-tenant troubleshooting and reporting.
	TenantID = "tenant_id"

	// OrganizationID is the field key that identifies the organization or
	// account associated with the log entry.
	OrganizationID = "org_id"

	// CustomerID is the field key that identifies the customer referenced
	// by the log entry, when applicable.
	CustomerID = "customer_id"

	// DeviceID is the field key that records the device identifier associated
	// with the event (for example, a mobile device ID).
	DeviceID = "device_id"

	// IP is the field key that records a relevant IP address, such as the
	// client or peer address.
	//
	// For HTTP-specific logs, prefer the more specific HTTPClientIP or
	// HTTPRemoteAddr keys where appropriate.
	IP = "ip"
)
