//go:build unit

package service

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestOpenAIUpstreamTransportError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{name: "unexpected eof", err: io.ErrUnexpectedEOF, want: true},
		{name: "proxy refused", err: errors.New("socks connect tcp: connection refused"), want: true},
		{name: "closed connection", err: errors.New("write tcp: use of closed network connection"), want: true},
		{name: "network timeout", err: &net.DNSError{IsTimeout: true}, want: true},
		{name: "request timeout", err: context.DeadlineExceeded, want: true},
		{name: "client cancelled", err: context.Canceled, want: false},
		{name: "request validation", err: errors.New("invalid request payload"), want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, isOpenAIUpstreamTransportError(tt.err))
		})
	}
}

func TestIsOpenAIStreamTransportFailure(t *testing.T) {
	require.True(t, IsOpenAIStreamTransportFailure(errors.New("stream read error: unexpected EOF")))
	require.True(t, IsOpenAIStreamTransportFailure(errors.New("stream usage incomplete: missing terminal event")))
	require.True(t, IsOpenAIStreamTransportFailure(errors.New("stream data interval timeout")))
	require.False(t, IsOpenAIStreamTransportFailure(errors.New("stream usage incomplete after disconnect: unexpected EOF")))
	require.False(t, IsOpenAIStreamTransportFailure(context.Canceled))
}

func TestNewOpenAITransportRequestFailoverError(t *testing.T) {
	err := newOpenAITransportRequestFailoverError(io.ErrUnexpectedEOF)
	require.Equal(t, 502, err.StatusCode)
	require.True(t, err.RetryableOnSameAccount)
	require.Equal(t, 1, err.MaxSameAccountRetries)
	require.Contains(t, string(err.ResponseBody), "OpenAI upstream connection failed")
}

func TestOpenAIForward_TransportFailureReturnsFailoverErrorBeforeWritingResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	body := []byte(`{"model":"gpt-5.5","stream":false,"input":"hello"}`)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	svc := &OpenAIGatewayService{
		cfg:          &config.Config{},
		httpUpstream: &httpUpstreamRecorder{err: io.ErrUnexpectedEOF},
	}
	account := &Account{
		ID:          777,
		Platform:    PlatformOpenAI,
		Type:        AccountTypeOAuth,
		Concurrency: 1,
		Credentials: map[string]any{
			"access_token":       "oauth-token",
			"chatgpt_account_id": "chatgpt-account",
		},
	}

	result, err := svc.Forward(context.Background(), c, account, body)
	require.Nil(t, result)
	require.Error(t, err)
	var failoverErr *UpstreamFailoverError
	require.ErrorAs(t, err, &failoverErr)
	require.Equal(t, http.StatusBadGateway, failoverErr.StatusCode)
	require.True(t, failoverErr.RetryableOnSameAccount)
	require.Zero(t, recorder.Body.Len(), "the handler must remain free to select another account")
}
