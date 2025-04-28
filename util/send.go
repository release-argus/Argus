// Copyright [2024] [Argus]
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

// Package util provides utility functions for the Argus project.
package util

import (
	"errors"
	"math"
	"math/rand"
	"time"
)

// RetryWithBackoff retries the operation with jitter and exponential backoff.
//
// Parameters:
//   - operation: Function to execute.
//   - maxTries: Maximum amount retries.
//   - baseDelay: Initial delay before the first retry.
//   - maxDelay: Maximum delay between retries (exponential backoff).
//   - shouldStop: Function to check if the operation should stop.
func RetryWithBackoff(
	operation func() error,
	maxTries uint8,
	baseDelay, maxDelay time.Duration,
	shouldStop func() bool,
) error {
	var errs []error

	for try := uint8(0); try < maxTries; try++ {
		// Stop retrying?
		if shouldStop != nil && shouldStop() {
			return nil
		}

		err := operation()

		// SUCCESS.
		if err == nil {
			return nil
		}

		// FAIL: Append the error.
		errs = append(errs, err)

		// Don't delay after the last try.
		if try == maxTries-1 {
			break
		}
		// Space out retries with exponential backoff and jitter.
		delay := time.Duration(math.Min(
			float64(baseDelay)*math.Pow(2, float64(try)),
			float64(maxDelay)))
		//#nosec G404 -- jitter does not need cryptographic security.
		jitter := time.Duration(rand.Int63n(int64(baseDelay))) // Add randomness to avoid synchronized retries.
		delay += jitter

		// Wait before retrying.
		time.Sleep(delay)
	}

	// All retries exhausted.
	return errors.Join(errs...)
}
