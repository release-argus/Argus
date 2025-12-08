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
	"fmt"
	"net/http"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Expected URL as command-line argument")
		os.Exit(1)
	}
	url := os.Args[1]

	//#nosec G402 -- Ignore TLS for healthcheck.
	http.DefaultTransport.(*http.Transport).TLSClientConfig.InsecureSkipVerify = true
	//#nosec G107 -- explicitly set URL.
	if _, err := http.Get(url); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v", err)
		os.Exit(1)
	}
}
