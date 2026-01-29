package duration

import (
	"time"

	"dirpx.dev/rxlog/rxapi/buffer"
)

// Encoder encodes a time.Duration into the provided Buffer.
//
// An Encoder takes an existing destination buffer `dst` and a duration `d`,
// and returns a (potentially new) Buffer that contains the encoded
// representation of `d`. Implementations SHOULD append to `dst` when possible
// to minimize allocations, but MAY obtain and return a different Buffer (for
// example, from a pool) if necessary.
//
// The returned Buffer MUST be treated as the authoritative destination after
// the call. Callers MUST use the returned Buffer for any subsequent writes
// and MUST eventually release it according to the Buffer contract (typically
// by calling Free). Implementations MUST NOT call Free on `dst` or on the
// returned Buffer; lifetime management is the caller's responsibility.
//
// If the implementation returns a different Buffer than `dst`, the original
// `dst` MUST remain in a valid state according to its own contract, and the
// Encoder MUST NOT retain references to either Buffer beyond the duration of
// the call.
//
// Implementations SHOULD document the chosen duration format (for example,
// Go's default duration string, a numeric value in a specific unit, or a
// structured representation) and SHOULD keep it stable over time so that
// downstream systems can reliably parse or interpret it.
type Encoder func(dst *buffer.Buffer, d time.Duration) *buffer.Buffer
