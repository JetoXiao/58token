package service

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/apicompat"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestRequestedChatCompletionsMaxTokens(t *testing.T) {
	maxTokens := 5
	maxCompletionTokens := 7

	got, ok := requestedChatCompletionsMaxTokens(&apicompat.ChatCompletionsRequest{
		MaxTokens:           &maxTokens,
		MaxCompletionTokens: &maxCompletionTokens,
	})

	require.True(t, ok)
	require.Equal(t, 7, got)
}

func TestGatewayForwardAsChatCompletionsPreservesSmallMaxTokensForAnthropic(t *testing.T) {
	gin.SetMode(gin.TestMode)

	upstream := &anthropicHTTPUpstreamRecorder{
		resp: &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
			Body:       io.NopCloser(strings.NewReader(minimalAnthropicCompatSSE())),
		},
	}
	svc := &GatewayService{
		cfg:          &config.Config{Gateway: config.GatewayConfig{MaxLineSize: defaultMaxLineSize}},
		httpUpstream: upstream,
	}
	account := &Account{
		ID:          501,
		Name:        "anthropic-chat-compat",
		Platform:    PlatformAnthropic,
		Type:        AccountTypeAPIKey,
		Concurrency: 1,
		Credentials: map[string]any{
			"api_key":  "upstream-anthropic-key",
			"base_url": "https://api.anthropic.com",
		},
	}
	body := []byte(`{"model":"claude-sonnet-4-6","messages":[{"role":"user","content":"hi"}],"max_tokens":5,"max_completion_tokens":7}`)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader(body))

	result, err := svc.ForwardAsChatCompletions(context.Background(), c, account, body, nil)

	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, http.StatusOK, rec.Code)
	require.NotNil(t, upstream.lastReq)
	require.Equal(t, int64(7), gjson.GetBytes(upstream.lastBody, "max_tokens").Int())
	require.False(t, gjson.GetBytes(upstream.lastBody, "max_completion_tokens").Exists())
}

func TestGeminiForwardAsChatCompletionsPreservesSmallMaxTokens(t *testing.T) {
	gin.SetMode(gin.TestMode)

	httpStub := &geminiCompatHTTPUpstreamStub{
		response: &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
			Body:       io.NopCloser(strings.NewReader(`data: {"candidates":[{"content":{"parts":[{"text":"ok"}]},"finishReason":"STOP"}],"usageMetadata":{"promptTokenCount":1,"candidatesTokenCount":1}}` + "\n\n" + "data: [DONE]\n\n")),
		},
	}
	svc := &GeminiMessagesCompatService{
		httpUpstream: httpStub,
		cfg:          &config.Config{},
	}
	account := &Account{
		ID:          502,
		Platform:    PlatformGemini,
		Type:        AccountTypeAPIKey,
		Concurrency: 1,
		Credentials: map[string]any{"api_key": "gemini-api-key"},
	}
	body := []byte(`{"model":"gemini-2.5-flash","stream":true,"messages":[{"role":"user","content":"hi"}],"max_tokens":5,"max_completion_tokens":7}`)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader(body))

	result, err := svc.ForwardAsChatCompletions(context.Background(), c, account, body)

	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, http.StatusOK, rec.Code)
	require.NotNil(t, httpStub.lastReq)
	sentBody, err := io.ReadAll(httpStub.lastReq.Body)
	require.NoError(t, err)
	require.Equal(t, int64(7), gjson.GetBytes(sentBody, "generationConfig.maxOutputTokens").Int())
}

func minimalAnthropicCompatSSE() string {
	message := map[string]any{
		"type": "message_start",
		"message": map[string]any{
			"id":      "msg_test",
			"type":    "message",
			"role":    "assistant",
			"model":   "claude-sonnet-4-6",
			"content": []any{},
			"usage": map[string]any{
				"input_tokens": 1,
			},
		},
	}
	startBlock := map[string]any{
		"type":  "content_block_start",
		"index": 0,
		"content_block": map[string]any{
			"type": "text",
			"text": "",
		},
	}
	textDelta := map[string]any{
		"type":  "content_block_delta",
		"index": 0,
		"delta": map[string]any{
			"type": "text_delta",
			"text": "ok",
		},
	}
	messageDelta := map[string]any{
		"type": "message_delta",
		"delta": map[string]any{
			"stop_reason": "end_turn",
		},
		"usage": map[string]any{
			"output_tokens": 1,
		},
	}
	return strings.Join([]string{
		"data: " + mustMarshalJSON(message),
		"",
		"data: " + mustMarshalJSON(startBlock),
		"",
		"data: " + mustMarshalJSON(textDelta),
		"",
		"data: " + mustMarshalJSON(messageDelta),
		"",
		`data: {"type":"message_stop"}`,
		"",
		"data: [DONE]",
		"",
	}, "\n")
}

func mustMarshalJSON(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return string(b)
}
