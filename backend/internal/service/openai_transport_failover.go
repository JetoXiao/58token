package service

import (
	"context"
	"errors"
	"io"
	"net"
	"strings"
)

// isOpenAIUpstreamTransportError identifies failures that happened on the path
// to the upstream before it returned an HTTP response. These can safely be
// retried or routed to another account because no upstream result was received.
func isOpenAIUpstreamTransportError(err error) bool {
	if err == nil || errors.Is(err, context.Canceled) {
		return false
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
		return true
	}
	var networkErr net.Error
	if errors.As(err, &networkErr) && networkErr.Timeout() {
		return true
	}

	message := strings.ToLower(strings.TrimSpace(err.Error()))
	for _, marker := range []string{
		"unexpected eof",
		"connection refused",
		"connection reset",
		"broken pipe",
		"closed network connection",
		"use of closed network connection",
		"proxyconnect",
		"socks connect",
		"tls handshake timeout",
		"i/o timeout",
	} {
		if strings.Contains(message, marker) {
			return true
		}
	}
	return false
}

// IsOpenAIStreamTransportFailure is deliberately limited to an interrupted
// upstream SSE read. Client cancellations and protocol/business errors must
// not affect account scheduling.
func IsOpenAIStreamTransportFailure(err error) bool {
	if err == nil || errors.Is(err, context.Canceled) {
		return false
	}
	message := strings.ToLower(err.Error())
	if strings.Contains(message, "stream usage incomplete after disconnect") {
		return false
	}
	if strings.Contains(message, "stream read error") {
		return isOpenAIUpstreamTransportError(err)
	}
	return strings.Contains(message, "stream usage incomplete: missing terminal event") ||
		strings.Contains(message, "stream data interval timeout")
}

func openAIUpstreamTransportErrorMessage(err error) string {
	if err == nil {
		return "OpenAI upstream connection failed"
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return "OpenAI upstream connection timed out"
	}
	if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
		return "OpenAI upstream connection closed unexpectedly"
	}
	var networkErr net.Error
	if errors.As(err, &networkErr) && networkErr.Timeout() {
		return "OpenAI upstream connection timed out"
	}

	message := strings.ToLower(strings.TrimSpace(err.Error()))
	switch {
	case strings.Contains(message, "connection refused") || strings.Contains(message, "socks connect"):
		return "OpenAI upstream proxy connection failed"
	case strings.Contains(message, "connection reset"), strings.Contains(message, "broken pipe"), strings.Contains(message, "closed network connection"):
		return "OpenAI upstream connection was interrupted"
	case strings.Contains(message, "timeout"):
		return "OpenAI upstream connection timed out"
	default:
		return "OpenAI upstream connection failed"
	}
}

func newOpenAITransportRequestFailoverError(err error) *UpstreamFailoverError {
	return &UpstreamFailoverError{
		StatusCode:             502,
		ResponseBody:           []byte(`{"error":{"type":"upstream_error","message":"OpenAI upstream connection failed"}}`),
		RetryableOnSameAccount: true,
		MaxSameAccountRetries:  1,
	}
}
