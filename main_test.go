package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestPublicJSONSkipsReverseDNS(t *testing.T) {
	calls := 0
	a := testApp(func(string) ([]string, error) {
		calls++
		return []string{"host.example."}, nil
	})

	resp := performRequest(a, "/json")
	if resp.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.Code, http.StatusOK)
	}

	var body map[string]any
	if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode json: %v", err)
	}
	if _, ok := body["reverse_dns"]; ok {
		t.Fatalf("public /json included reverse_dns: %v", body["reverse_dns"])
	}
	if calls != 0 {
		t.Fatalf("reverse DNS lookup calls = %d, want 0", calls)
	}
}

func TestPrivateReverseDNSJSONIncludesReverseDNS(t *testing.T) {
	calls := 0
	a := testApp(func(string) ([]string, error) {
		calls++
		return []string{"host.example."}, nil
	})

	resp := performRequest(a, "/json.rdns")
	if resp.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.Code, http.StatusOK)
	}

	var body map[string]any
	if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode json: %v", err)
	}
	if got := body["reverse_dns"]; got != "host.example" {
		t.Fatalf("reverse_dns = %v, want host.example", got)
	}
	if calls != 1 {
		t.Fatalf("reverse DNS lookup calls = %d, want 1", calls)
	}
}

func TestIndexSkipsReverseDNS(t *testing.T) {
	calls := 0
	a := testApp(func(string) ([]string, error) {
		calls++
		return []string{"host.example."}, nil
	})

	resp := performRequest(a, "/")
	if resp.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.Code, http.StatusOK)
	}
	if calls != 0 {
		t.Fatalf("reverse DNS lookup calls = %d, want 0", calls)
	}
}

func TestPublicSummaryPathsSkipReverseDNS(t *testing.T) {
	for _, path := range []string{"/geo", "/all", "/all.json"} {
		t.Run(path, func(t *testing.T) {
			calls := 0
			a := testApp(func(string) ([]string, error) {
				calls++
				return []string{"host.example."}, nil
			})

			resp := performRequest(a, path)
			if resp.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d", resp.Code, http.StatusOK)
			}
			if strings.Contains(resp.Body.String(), "host.example") {
				t.Fatalf("%s response included reverse DNS host: %s", path, resp.Body.String())
			}
			if calls != 0 {
				t.Fatalf("reverse DNS lookup calls = %d, want 0", calls)
			}
		})
	}
}

func TestRateLimitAllowsTwentyRequestsPerMinute(t *testing.T) {
	now := time.Date(2026, 5, 10, 12, 0, 0, 0, time.UTC)
	a := testRateLimitedApp(func() time.Time { return now })

	for i := 0; i < 20; i++ {
		resp := performRequest(a, "/ip")
		if resp.Code != http.StatusOK {
			t.Fatalf("request %d status = %d, want %d", i+1, resp.Code, http.StatusOK)
		}
	}

	resp := performRequest(a, "/ip")
	if resp.Code != http.StatusTooManyRequests {
		t.Fatalf("request 21 status = %d, want %d", resp.Code, http.StatusTooManyRequests)
	}
	if got := resp.Header().Get("Retry-After"); got != "60" {
		t.Fatalf("Retry-After = %q, want 60", got)
	}
	if body := resp.Body.String(); !strings.Contains(body, "CGNAT") {
		t.Fatalf("rate limit response did not mention CGNAT: %s", body)
	}

	now = now.Add(time.Minute)
	resp = performRequest(a, "/ip")
	if resp.Code != http.StatusOK {
		t.Fatalf("after window reset status = %d, want %d", resp.Code, http.StatusOK)
	}
}

func TestRateLimitSkipsHealthCheck(t *testing.T) {
	now := time.Date(2026, 5, 10, 12, 0, 0, 0, time.UTC)
	a := testRateLimitedApp(func() time.Time { return now })

	for i := 0; i < 20; i++ {
		resp := performRequest(a, "/ip")
		if resp.Code != http.StatusOK {
			t.Fatalf("request %d status = %d, want %d", i+1, resp.Code, http.StatusOK)
		}
	}

	resp := performRequest(a, "/healthz")
	if resp.Code != http.StatusOK {
		t.Fatalf("healthz status = %d, want %d", resp.Code, http.StatusOK)
	}
}

func testApp(lookupAddr func(string) ([]string, error)) *app {
	return &app{
		cfg: config{
			addr:     ":8080",
			basePath: "",
		},
		lookupAddr: lookupAddr,
	}
}

func testRateLimitedApp(now func() time.Time) *app {
	a := testApp(func(string) ([]string, error) { return nil, nil })
	a.cfg.rateLimit.requestsPerMinute = 20
	a.rateLimiter = newFixedWindowRateLimiter(20, time.Minute, now)
	return a
}

func performRequest(a *app, path string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	resp := httptest.NewRecorder()
	a.route(resp, req)
	return resp
}
