//go:build unit

package service

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestExtractResponsesReasoningEffortFromBody(t *testing.T) {
	t.Parallel()

	got := ExtractResponsesReasoningEffortFromBody([]byte(`{"model":"claude-sonnet-4.5","reasoning":{"effort":"HIGH"}}`))
	require.NotNil(t, got)
	require.Equal(t, "high", *got)

	require.Nil(t, ExtractResponsesReasoningEffortFromBody([]byte(`{"model":"claude-sonnet-4.5"}`)))
}

func TestHandleResponsesBufferedStreamingResponse_PreservesMessageStartCacheUsage(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)

	resp := &http.Response{
		Header: http.Header{"x-request-id": []string{"rid_buffered"}},
		Body: io.NopCloser(strings.NewReader(strings.Join([]string{
			`event: message_start`,
			`data: {"type":"message_start","message":{"id":"msg_1","type":"message","role":"assistant","content":[],"model":"claude-sonnet-4.5","stop_reason":"","usage":{"input_tokens":12,"cache_read_input_tokens":9,"cache_creation_input_tokens":3}}}`,
			``,
			`event: content_block_start`,
			`data: {"type":"content_block_start","index":0,"content_block":{"type":"text","text":"hello"}}`,
			``,
			`event: message_delta`,
			`data: {"type":"message_delta","delta":{"stop_reason":"end_turn"},"usage":{"output_tokens":7}}`,
			``,
			`event: message_stop`,
			`data: {"type":"message_stop"}`,
			``,
		}, "\n"))),
	}

	svc := &GatewayService{}
	result, err := svc.handleResponsesBufferedStreamingResponse(resp, c, &Account{ID: 1}, "claude-sonnet-4.5", "claude-sonnet-4.5", nil, time.Now())
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, 12, result.Usage.InputTokens)
	require.Equal(t, 7, result.Usage.OutputTokens)
	require.Equal(t, 9, result.Usage.CacheReadInputTokens)
	require.Equal(t, 3, result.Usage.CacheCreationInputTokens)
	require.Contains(t, rec.Body.String(), `"cached_tokens":9`)
}

func TestHandleResponsesStreamingResponse_PreservesMessageStartCacheUsage(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)

	resp := &http.Response{
		Header: http.Header{"x-request-id": []string{"rid_stream"}},
		Body: io.NopCloser(strings.NewReader(strings.Join([]string{
			`event: message_start`,
			`data: {"type":"message_start","message":{"id":"msg_2","type":"message","role":"assistant","content":[],"model":"claude-sonnet-4.5","stop_reason":"","usage":{"input_tokens":20,"cache_read_input_tokens":11,"cache_creation_input_tokens":4}}}`,
			``,
			`event: content_block_start`,
			`data: {"type":"content_block_start","index":0,"content_block":{"type":"text","text":"hello"}}`,
			``,
			`event: message_delta`,
			`data: {"type":"message_delta","delta":{"stop_reason":"end_turn"},"usage":{"output_tokens":8}}`,
			``,
			`event: message_stop`,
			`data: {"type":"message_stop"}`,
			``,
		}, "\n"))),
	}

	svc := &GatewayService{}
	result, err := svc.handleResponsesStreamingResponse(resp, c, &Account{ID: 1}, "claude-sonnet-4.5", "claude-sonnet-4.5", nil, time.Now())
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, 20, result.Usage.InputTokens)
	require.Equal(t, 8, result.Usage.OutputTokens)
	require.Equal(t, 11, result.Usage.CacheReadInputTokens)
	require.Equal(t, 4, result.Usage.CacheCreationInputTokens)
	require.Contains(t, rec.Body.String(), `response.completed`)
}

func TestHandleResponsesStreamingResponse_ErrorBeforeOutputTriggersFailover(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil)

	resp := &http.Response{
		Header: http.Header{"x-request-id": []string{"rid_stream_error"}},
		Body: io.NopCloser(strings.NewReader(strings.Join([]string{
			`event: error`,
			`data: {"type":"error","error":{"type":"service_unavailable","message":"Service temporarily unavailable, please retry later."}}`,
			``,
		}, "\n"))),
	}

	svc := &GatewayService{}
	result, err := svc.handleResponsesStreamingResponse(resp, c, &Account{ID: 36, Platform: PlatformAnthropic}, "claude-sonnet-5", "claude-sonnet-5", nil, time.Now())
	require.Nil(t, result)
	require.Error(t, err)
	var failoverErr *UpstreamFailoverError
	require.True(t, errors.As(err, &failoverErr))
	require.Equal(t, http.StatusServiceUnavailable, failoverErr.StatusCode)
	require.False(t, c.Writer.Written())
	require.Empty(t, rec.Body.String())
}

func TestHandleResponsesStreamingResponse_ErrorAfterOutputEmitsResponseFailed(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil)

	resp := &http.Response{
		Header: http.Header{"x-request-id": []string{"rid_partial_error"}},
		Body: io.NopCloser(strings.NewReader(strings.Join([]string{
			`event: message_start`,
			`data: {"type":"message_start","message":{"id":"msg_partial","type":"message","role":"assistant","content":[],"model":"claude-sonnet-5","stop_reason":"","usage":{"input_tokens":5}}}`,
			``,
			`event: error`,
			`data: {"type":"error","error":{"type":"service_unavailable","message":"Service temporarily unavailable"}}`,
			``,
		}, "\n"))),
	}

	svc := &GatewayService{}
	result, err := svc.handleResponsesStreamingResponse(resp, c, &Account{ID: 36, Platform: PlatformAnthropic}, "claude-sonnet-5", "claude-sonnet-5", nil, time.Now())
	require.NotNil(t, result)
	require.Error(t, err)
	require.True(t, c.Writer.Written())
	require.Contains(t, rec.Body.String(), `response.created`)
	require.Contains(t, rec.Body.String(), `response.failed`)
	require.Contains(t, rec.Body.String(), `service_unavailable`)
	require.NotContains(t, rec.Body.String(), `response.completed`)
}

func TestHandleResponsesStreamingResponse_EmptyStreamTriggersFailover(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil)
	resp := &http.Response{Body: io.NopCloser(strings.NewReader(""))}

	svc := &GatewayService{}
	result, err := svc.handleResponsesStreamingResponse(resp, c, &Account{ID: 36}, "claude-sonnet-5", "claude-sonnet-5", nil, time.Now())
	require.Nil(t, result)
	var failoverErr *UpstreamFailoverError
	require.ErrorAs(t, err, &failoverErr)
	require.True(t, failoverErr.RetryableOnSameAccount)
	require.False(t, c.Writer.Written())
}

func TestHandleResponsesBufferedStreamingResponse_ErrorTriggersFailover(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil)
	resp := &http.Response{
		Body: io.NopCloser(strings.NewReader(strings.Join([]string{
			`event: error`,
			`data: {"type":"error","error":{"type":"overloaded_error","message":"Overloaded"}}`,
			``,
		}, "\n"))),
	}

	svc := &GatewayService{}
	result, err := svc.handleResponsesBufferedStreamingResponse(resp, c, &Account{ID: 36, Platform: PlatformAnthropic}, "claude-sonnet-5", "claude-sonnet-5", nil, time.Now())
	require.Nil(t, result)
	var failoverErr *UpstreamFailoverError
	require.ErrorAs(t, err, &failoverErr)
	require.Equal(t, http.StatusServiceUnavailable, failoverErr.StatusCode)
	require.False(t, c.Writer.Written())
}

func TestAnthropicCompatRetryableBodyNormalizesMisleading422(t *testing.T) {
	t.Parallel()

	body := []byte(`{"error":{"type":"service_unavailable","message":"Service temporarily unavailable"}}`)
	require.True(t, shouldFailoverAnthropicCompatResponse(http.StatusUnprocessableEntity, body))
	require.Equal(t, http.StatusServiceUnavailable, anthropicCompatFailoverStatus(http.StatusUnprocessableEntity, body))
	require.False(t, shouldFailoverAnthropicCompatResponse(http.StatusUnprocessableEntity, []byte(`{"error":{"type":"invalid_request_error"}}`)))
}
