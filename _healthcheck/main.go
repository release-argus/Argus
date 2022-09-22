/*
./healthcheck http://localhost:8080/api/v1/healthcheck
200   == nothing
error == os.Exit(1)
*/
package main

import (
	"crypto/tls"
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

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	if _, err := http.Get(url); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v", err)
		os.Exit(1)
	}
}
