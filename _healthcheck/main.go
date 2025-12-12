// Copyright [2025] [Argus]
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

/*
./healthcheck http://localhost:8080/api/v1/healthcheck

	200   == nothing
	error == os.Exit(1)
*/
package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"
)

func run(args []string) error {
	if len(args) < 1 {
		return errors.New("expected URL as command-line argument")
	}

	url := args[0]

	tr := http.DefaultTransport.(*http.Transport).Clone()
	//#nosec G402 -- Ignore TLS for the health check.
	tr.TLSClientConfig.InsecureSkipVerify = true
	http.DefaultTransport.(*http.Transport).TLSClientConfig.InsecureSkipVerify = true
	client := &http.Client{Transport: tr}

	//#nosec G107 -- explicitly set URL.
	if _, err := client.Get(url); err != nil {
		return fmt.Errorf("error: %w", err)
	}

	return nil
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}
}
