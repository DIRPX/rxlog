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

package clock_test

import (
	"testing"
	"time"

	apiclock "dirpx.dev/rxlog/rxapi/chrono/clock"
	coreclock "dirpx.dev/rxlog/rxcore/chrono/clock"
)

// TestDefaultClockImplementsClockInterface verifies that DefaultClock
// satisfies the chrono/clock.Clock interface at compile time and is
// non-nil as an interface value.
func TestDefaultClockImplementsClockInterface(t *testing.T) {
	var c apiclock.Clock = coreclock.DefaultClock
	if c == nil {
		t.Fatalf("DefaultClock as apiclock.Clock is nil")
	}
}

// TestDefaultClockNowIsCloseToTimeNow verifies that DefaultClock.Now()
// delegates to time.Now by checking that its result lies between two
// consecutive time.Now() calls.
func TestDefaultClockNowIsCloseToTimeNow(t *testing.T) {
	// Capture time before and after calling DefaultClock.Now.
	before := time.Now()
	got := coreclock.DefaultClock.Now()
	after := time.Now()

	if got.Before(before) || got.After(after) {
		t.Fatalf("DefaultClock.Now() = %v, want between %v and %v", got, before, after)
	}
}

// TestDefaultClockNewTickerCreatesTicker verifies that NewTicker returns
// a non-nil *time.Ticker whose channel delivers at least one tick within
// a reasonable timeout.
func TestDefaultClockNewTickerCreatesTicker(t *testing.T) {
	const interval = 10 * time.Millisecond

	ticker := coreclock.DefaultClock.NewTicker(interval)
	if ticker == nil {
		t.Fatal("DefaultClock.NewTicker returned nil ticker")
	}
	defer ticker.Stop()

	select {
	case <-ticker.C:
		// Received at least one tick; this is sufficient to confirm that
		// the ticker is functional for the purposes of this test.
	case <-time.After(500 * time.Millisecond):
		t.Fatalf("no tick received from ticker within timeout")
	}
}
