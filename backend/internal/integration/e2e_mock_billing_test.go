//go:build e2e

package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

// These tests exercise the mock upstream described in /docs/README.md.  The
// billing test is opt-in because it mutates groups and needs an admin token.

func mockURL() string { return getEnv("MOCK_URL", "http://localhost:18080") }

func jsonRequest(t *testing.T, method, url, token string, payload any) (int, []byte) {
	t.Helper()
	var body io.Reader
	if payload != nil {
		encoded, err := json.Marshal(payload)
		if err != nil {
			t.Fatal(err)
		}
		body = bytes.NewReader(encoded)
	}
	req, err := http.NewRequestWithContext(t.Context(), method, url, body)
	if err != nil {
		t.Fatal(err)
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := (&http.Client{Timeout: 20 * time.Second}).Do(req)
	if err != nil {
		t.Fatalf("%s %s: %v", method, url, err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	return resp.StatusCode, data
}

func TestMockDocumentedModelsAndUsage(t *testing.T) {
	status, body := jsonRequest(t, http.MethodGet, mockURL()+"/healthz", "", nil)
	if status != http.StatusOK {
		t.Skipf("mock upstream is unavailable (%d): %s", status, body)
	}

	tests := []struct {
		name  string
		path  string
		body  map[string]any
		check func(t *testing.T, response map[string]any)
	}{
		{
			name: "OpenAI chat gpt-4o",
			path: "/v1/chat/completions",
			body: map[string]any{"model": "gpt-4o:in=12:out=7:cache=3", "messages": []map[string]string{{"role": "user", "content": "hello"}}},
			check: func(t *testing.T, response map[string]any) {
				usage, ok := response["usage"].(map[string]any)
				if !ok || usage["prompt_tokens"] != float64(12) || usage["completion_tokens"] != float64(7) {
					t.Fatalf("unexpected OpenAI usage: %#v", response["usage"])
				}
			},
		},
		{
			name: "OpenAI responses gpt-4.1-mini",
			path: "/v1/responses",
			body: map[string]any{"model": "gpt-4.1-mini:in=9:out=4", "input": "hello"},
			check: func(t *testing.T, response map[string]any) {
				if response["object"] != "response" {
					t.Fatalf("unexpected Responses object: %#v", response["object"])
				}
			},
		},
		{
			name: "Anthropic claude-sonnet",
			path: "/anthropic/v1/messages",
			body: map[string]any{"model": "claude-sonnet-4-5:in=11:out=5:cache=2", "max_tokens": 20, "messages": []map[string]any{{"role": "user", "content": "hello"}}},
			check: func(t *testing.T, response map[string]any) {
				usage, ok := response["usage"].(map[string]any)
				if !ok || usage["input_tokens"] != float64(11) || usage["output_tokens"] != float64(5) {
					t.Fatalf("unexpected Anthropic usage: %#v", response["usage"])
				}
			},
		},
		{
			name: "Ark doubao seedance",
			path: "/ark/api/v3/contents/generations/tasks",
			body: map[string]any{"model": "doubao-seedance-1-0-pro-250528:ct=6:dur=2", "content": []map[string]any{{"type": "text", "text": "hello"}}},
			check: func(t *testing.T, response map[string]any) {
				if response["id"] == nil {
					t.Fatalf("Ark response has no task id: %#v", response)
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			status, body := jsonRequest(t, http.MethodPost, mockURL()+tc.path, "", tc.body)
			if status != http.StatusOK && status != http.StatusCreated {
				t.Fatalf("mock returned HTTP %d: %s", status, body)
			}
			var response map[string]any
			if err := json.Unmarshal(body, &response); err != nil {
				t.Fatalf("invalid JSON: %v (%s)", err, body)
			}
			tc.check(t, response)
		})
	}
}

func TestBillingGroupMultipliersAndGateway(t *testing.T) {
	adminToken := strings.TrimSpace(os.Getenv("E2E_ADMIN_TOKEN"))
	lowGroup, err1 := strconv.ParseInt(os.Getenv("E2E_LOW_GROUP_ID"), 10, 64)
	highGroup, err2 := strconv.ParseInt(os.Getenv("E2E_HIGH_GROUP_ID"), 10, 64)
	userID, err3 := strconv.ParseInt(os.Getenv("E2E_BILLING_USER_ID"), 10, 64)
	lowKey := strings.TrimSpace(os.Getenv("E2E_LOW_GROUP_API_KEY"))
	highKey := strings.TrimSpace(os.Getenv("E2E_HIGH_GROUP_API_KEY"))
	if adminToken == "" || err1 != nil || err2 != nil || err3 != nil || lowKey == "" || highKey == "" {
		t.Skip("set E2E_ADMIN_TOKEN, E2E_LOW_GROUP_ID, E2E_HIGH_GROUP_ID, E2E_BILLING_USER_ID and both group API keys")
	}

	setRate := func(groupID int64, rate float64) {
		path := fmt.Sprintf("%s/api/v1/admin/groups/%d/rate-multipliers", baseURL, groupID)
		status, body := jsonRequest(t, http.MethodPut, path, adminToken, map[string]any{"entries": []map[string]any{{"user_id": userID, "rate_multiplier": rate}}})
		if status != http.StatusOK {
			t.Fatalf("set group %d rate: HTTP %d: %s", groupID, status, body)
		}
	}
	setRate(lowGroup, 0.5)
	setRate(highGroup, 2.0)

	for _, tc := range []struct {
		name  string
		key   string
		group int64
		rate  float64
	}{
		{"low multiplier", lowKey, lowGroup, 0.5},
		{"high multiplier", highKey, highGroup, 2.0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := fmt.Sprintf("%s/api/v1/admin/groups/%d/rate-multipliers", baseURL, tc.group)
			status, body := jsonRequest(t, http.MethodGet, path, adminToken, nil)
			if status != http.StatusOK || !strings.Contains(string(body), fmt.Sprintf("%g", tc.rate)) {
				t.Fatalf("rate was not persisted: HTTP %d: %s", status, body)
			}

			status, body = jsonRequest(t, http.MethodPost, baseURL+"/v1/chat/completions", tc.key, map[string]any{
				"model":    "gpt-4o",
				"messages": []map[string]string{{"role": "user", "content": "billing smoke test"}},
			})
			if status != http.StatusOK {
				t.Fatalf("gateway billing request failed: HTTP %d: %s", status, body)
			}
			var response map[string]any
			if err := json.Unmarshal(body, &response); err != nil {
				t.Fatal(err)
			}
			if _, ok := response["usage"]; !ok {
				t.Fatalf("gateway response has no usage: %s", body)
			}
		})
	}
}

// TestMockBillingMatrix runs repeated usage samples and verifies that an
// upstream 503 is returned as an error without creating a usage charge.
func TestMockBillingMatrix(t *testing.T) {
	claudeKey := getEnv("E2E_CLAUDE_API_KEY", "e2e_mock_claude_key_2026")
	deepseekKey := getEnv("E2E_DEEPSEEK_API_KEY", "e2e_mock_deepseek_key_2026")
	for i := 0; i < 5; i++ {
		status, body := jsonRequest(t, http.MethodPost, baseURL+"/v1/messages", claudeKey, map[string]any{
			"model": fmt.Sprintf("claude-sonnet-4-5:in=%d:out=%d:cache=%d", 10+i, 5+i, 2+i), "max_tokens": 32,
			"messages": []map[string]string{{"role": "user", "content": "claude billing matrix"}},
		})
		if status != http.StatusOK {
			t.Fatalf("Claude iteration %d: HTTP %d: %s", i, status, body)
		}
	}
	for i := 0; i < 5; i++ {
		status, body := jsonRequest(t, http.MethodPost, baseURL+"/v1/chat/completions", deepseekKey, map[string]any{
			"model":    fmt.Sprintf("deepseek-chat:in=%d:out=%d:cache=%d", 20+i, 6+i, 3+i),
			"messages": []map[string]string{{"role": "user", "content": "deepseek billing matrix"}},
		})
		if status != http.StatusOK {
			t.Fatalf("DeepSeek iteration %d: HTTP %d: %s", i, status, body)
		}
	}

	// A model-suffixed 503 must not produce a billable usage row.
	status, body := jsonRequest(t, http.MethodPost, baseURL+"/v1/chat/completions", deepseekKey, map[string]any{
		"model": "deepseek-chat:503", "messages": []map[string]string{{"role": "user", "content": "upstream failure"}},
	})
	if status == http.StatusOK {
		t.Fatalf("503 request unexpectedly succeeded: %s", body)
	}

	// Seedance is an upstream-only Ark protocol in the current gateway. Verify
	// the documented ct/dur billing payload and a pinned 503 response directly.
	for i := 0; i < 5; i++ {
		status, body = jsonRequest(t, http.MethodPost, mockURL()+"/ark/api/v3/contents/generations/tasks", "", map[string]any{
			"model":   fmt.Sprintf("doubao-seedance-1-0-pro-250528:ct=%d:dur=%d:task=succeeded", 100+i, 2+i),
			"content": []map[string]any{{"type": "text", "text": "seedance billing matrix"}},
		})
		if status != http.StatusOK {
			t.Fatalf("Seedance iteration %d: HTTP %d: %s", i, status, body)
		}
	}
	status, body = jsonRequest(t, http.MethodPost, mockURL()+"/ark/api/v3/contents/generations/tasks", "", map[string]any{
		"model": "doubao-seedance-1-0-pro-250528:503", "content": []map[string]any{{"type": "text", "text": "upstream failure"}},
	})
	if status != http.StatusServiceUnavailable {
		t.Fatalf("Seedance expected 503, got HTTP %d: %s", status, body)
	}
}
