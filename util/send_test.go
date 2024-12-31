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

//go:build unit

package util

import (
	"errors"
	"testing"
	"time"
)

func TestRetryWithBackoff_Success(t *testing.T) {
	operation := func() error {
		return nil
	}

	err := RetryWithBackoff(operation, 3, 100*time.Millisecond, 1*time.Second, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestRetryWithBackoff_Failure(t *testing.T) {
	expectedErr := errors.New("operation failed")
	operation := func() error {
		return expectedErr
	}

	err := RetryWithBackoff(operation, 3, 100*time.Millisecond, 1*time.Second, nil)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
}

func TestRetryWithBackoff_StopCondition(t *testing.T) {
	operation := func() error {
		return errors.New("operation failed")
	}
	shouldStop := func() bool {
		return true
	}

	err := RetryWithBackoff(operation, 3, 100*time.Millisecond, 1*time.Second, shouldStop)
	if err != nil {
		t.Fatalf("expected no error due to stop condition, got %v", err)
	}
}

func TestRetryWithBackoff_ExponentialBackoff(t *testing.T) {
	operation := func() error {
		return errors.New("operation failed")
	}

	start := time.Now()
	err := RetryWithBackoff(operation, 3, 100*time.Millisecond, 1*time.Second, nil)
	elapsed := time.Since(start)

	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if elapsed < 300*time.Millisecond {
		t.Fatalf("expected at least 300ms delay, got %v", elapsed)
	}
}
