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

//go:build unit

package util

import (
	"errors"
	"fmt"
	"testing"
	"time"
)

func TestRetryWithBackoff_Success(t *testing.T) {
	// GIVEN: a successful operation.
	operation := func() error {
		return nil
	}

	// WHEN: RetryWithBackoff() is called.
	err := RetryWithBackoff(
		operation, 3,
		100*time.Millisecond,
		1*time.Second,
		nil,
	)

	// THEN: no error is returned.
	if err != nil {
		t.Fatalf(
			"%s\nRetryWithBackoff() error mismatch\ngot:  %v\nwant: nil",
			packageName, err,
		)
	}
}

func TestRetryWithBackoff_Failure(t *testing.T) {
	// GIVEN: a failed operation.
	expectedErr := errors.New("operation failed")
	operation := func() error {
		return expectedErr
	}

	// WHEN: RetryWithBackoff() is called.
	err := RetryWithBackoff(
		operation, 3,
		100*time.Millisecond,
		1*time.Second,
		nil,
	)

	prefix := fmt.Sprintf("%s\nRetryWithBackoff()", packageName)

	// THEN: the expected error is returned.
	if err == nil {
		t.Fatalf(
			"%s error mismatch\ngot:  nil\nwant: %v",
			prefix, expectedErr,
		)
	}

	// AND: the error is of the expected type.
	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"%s error type mismatch\ngot:  %v\nwant: %v",
			prefix, err, expectedErr,
		)
	}
}

func TestRetryWithBackoff_StopCondition(t *testing.T) {
	// GIVEN: a stop condition that returns true.
	operation := func() error {
		return errors.New("operation failed")
	}
	shouldStop := func() bool {
		return true
	}

	// WHEN: RetryWithBackoff() is called.
	err := RetryWithBackoff(
		operation, 3,
		100*time.Millisecond,
		1*time.Second, shouldStop,
	)

	prefix := fmt.Sprintf("%s\nRetryWithBackoff()", packageName)

	// THEN: the stop condition prevents the error.
	if err != nil {
		t.Fatalf(
			"%s stop condition didn't prevent error\ngot:  %v\nwant: nil",
			prefix, err,
		)
	}
}

func TestRetryWithBackoff_ExponentialBackoff(t *testing.T) {
	// GIVEN: a failed operation.
	operation := func() error {
		return errors.New("operation failed")
	}

	start := time.Now()
	// WHEN: RetryWithBackoff() is called.
	err := RetryWithBackoff(
		operation, 3,
		100*time.Millisecond,
		1*time.Second, nil,
	)
	elapsed := time.Since(start)

	prefix := fmt.Sprintf("%s\nRetryWithBackoff()", packageName)

	// THEN: the error is returned after the expected delay.
	if err == nil {
		t.Fatalf("%s error mismatch\ngot:  nil\nwant: error", prefix)
	}
	if elapsed < 300*time.Millisecond {
		t.Fatalf(
			"%s\nRetryWithBackoff() delay mismatch\ngot:  %v\nwant: 300ms+",
			prefix, elapsed,
		)
	}
}
