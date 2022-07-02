// Copyright [2022] [Argus]
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

package v1

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetIPWithCloudflare(t *testing.T) {
	// GIVEN a HTTP request is being made with the CF-CONNECTING-IP header
	req := httptest.NewRequest(http.MethodGet, "/approvals", nil)
	header := "CF-Connecting-IP"
	ip := "1.1.1.1"
	req.Header.Set(header, ip)

	// WHEN getIP is called on this request
	got := getIP(req)

	// THEN the IP on the Cloudlare header will be returned
	if got != ip {
		t.Errorf("Expected IP on the %q header (%q) to be returned, not %q",
			header, ip, got)
	}
}

func TestGetIPWithXRealIP(t *testing.T) {
	// GIVEN a HTTP request is being made with the X-REAL-IP header
	req := httptest.NewRequest(http.MethodGet, "/approvals", nil)
	header := "X-REAL-IP"
	ip := "1.1.1.1"
	req.Header.Set(header, ip)

	// WHEN getIP is called on this request
	got := getIP(req)

	// THEN the IP on the Cloudlare header will be returned
	if got != ip {
		t.Errorf("Expected IP on the %q header (%q) to be returned, not %q",
			header, ip, got)
	}
}

func TestGetIPWithXForwardedFor(t *testing.T) {
	// GIVEN a HTTP request is being made with the X-FORWARDED-FOR header
	req := httptest.NewRequest(http.MethodGet, "/approvals", nil)
	header := "X-FORWARDED-FOR"
	ip := "1.1.1.1"
	req.Header.Set(header, ip)

	// WHEN getIP is called on this request
	got := getIP(req)

	// THEN the IP on the Cloudlare header will be returned
	if got != ip {
		t.Errorf("Expected IP on the %q header (%q) to be returned, not %q",
			header, ip, got)
	}
}

func TestGetIPWithRemoteAddr(t *testing.T) {
	// GIVEN a HTTP request is being made
	req := httptest.NewRequest(http.MethodGet, "/approvals", nil)
	ip := "1.1.1.1"
	req.RemoteAddr = fmt.Sprintf("%s:123", ip)

	// WHEN getIP is called on this request
	got := getIP(req)

	// THEN the IP on the Cloudlare header will be returned
	if got != ip {
		t.Errorf("Expected IP on the RemoteAddr (%q) to be returned, not %q",
			ip, got)
	}
}

func TestGetIPWithRemoteAddrAndNoPort(t *testing.T) {
	// GIVEN a HTTP request is being made
	req := httptest.NewRequest(http.MethodGet, "/approvals", nil)
	ip := "1.1.1.1"
	req.RemoteAddr = ip

	// WHEN getIP is called on this request
	got := getIP(req)

	// THEN the IP on the Cloudlare header will be returned
	want := ""
	if got != want {
		t.Errorf("Expected %q, not %q as RemoteAddr doesn't contain a port (%q)",
			want, got, req.RemoteAddr)
	}
}

func TestGetIPWithInvalid(t *testing.T) {
	// GIVEN a HTTP request is being made
	req := httptest.NewRequest(http.MethodGet, "/approvals", nil)
	ip := "1111:123"
	req.RemoteAddr = ip

	// WHEN getIP is called on this request
	got := getIP(req)

	// THEN the IP on the Cloudlare header will be returned
	want := ""
	if got != want {
		t.Errorf("Expected %q, not %q as RemoteAddr isn't a valid IP (%q)",
			want, got, req.RemoteAddr)
	}
}
