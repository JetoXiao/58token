package handler

import (
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestBuildSanitizedRequestParamsFromBody_SummarizesWithoutImagePayload(t *testing.T) {
	body := []byte(`{
		"model": "gpt-5.4",
		"stream": true,
		"tools": [
			{"type": "image_generation", "size": "1024x1024", "quality": "high"}
		],
		"input": [
			{
				"role": "user",
				"content": [
					{"type": "input_text", "text": "make a small launch graphic"},
					{"type": "input_image", "image_url": "data:image/png;base64,SHOULD_NOT_APPEAR"}
				]
			}
		]
	}`)

	got := buildSanitizedRequestParamsFromBody(body)

	require.Equal(t, "gpt-5.4", got["model"])
	require.Equal(t, true, got["stream"])
	require.Equal(t, "1024x1024", got["image_size"])
	require.Equal(t, "high", got["quality"])
	require.Equal(t, []string{"image_generation"}, got["tool_types"])
	require.Equal(t, 1, got["input_items_count"])
	require.NotContains(t, got, "last_input_preview")
	require.NotContains(t, got, "last_user_message_preview")
	require.NotContains(t, got, "image_url")
	require.NotContains(t, got, "make a small launch graphic")
}

func TestBuildSanitizedOpenAIImagesRequestParams_MultipartSummary(t *testing.T) {
	compression := 80
	partialImages := 2

	got := buildSanitizedOpenAIImagesRequestParams(&service.OpenAIImagesRequest{
		Model:             "gpt-image-2",
		ExplicitModel:     true,
		Prompt:            "edit the product photo",
		Stream:            true,
		N:                 3,
		Size:              "1536x1024",
		ExplicitSize:      true,
		SizeTier:          "2K",
		Quality:           "high",
		OutputFormat:      "png",
		OutputCompression: &compression,
		PartialImages:     &partialImages,
		Multipart:         true,
		HasMask:           true,
		Uploads: []service.OpenAIImagesUpload{
			{FieldName: "image", FileName: "input.png", Data: []byte("SHOULD_NOT_APPEAR")},
		},
	})

	require.Equal(t, "gpt-image-2", got["model"])
	require.Equal(t, true, got["stream"])
	require.Equal(t, 3, got["n"])
	require.Equal(t, "1536x1024", got["image_size"])
	require.Equal(t, "2K", got["image_size_tier"])
	require.Equal(t, "high", got["quality"])
	require.Equal(t, "png", got["output_format"])
	require.Equal(t, 80, got["output_compression"])
	require.Equal(t, 2, got["partial_images"])
	require.Equal(t, true, got["multipart"])
	require.Equal(t, 1, got["input_image_count"])
	require.Equal(t, true, got["has_mask"])
	require.Equal(t, 22, got["prompt_chars"])
	require.NotContains(t, got, "prompt_preview")
	require.NotContains(t, got, "uploads")
	require.NotContains(t, got, "file_name")
	require.NotContains(t, got, "image")
	require.NotContains(t, got, "edit the product photo")
}

func TestBuildTTFTObservationParamsIncludesResponseCacheBypassReason(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Header(service.ResponseCacheHeader, service.ResponseCacheStatusBypass+"; reason=prompt_too_long")

	got := buildTTFTObservationParams(c, nil, nil)

	require.Equal(t, service.ResponseCacheStatusBypass, got["response_cache_status"])
	require.Equal(t, "prompt_too_long", got["response_cache_bypass_reason"])
}

func TestSchedulerDecisionReasonReportsStickyTTFTBypass(t *testing.T) {
	require.Equal(t, "sticky_ttft_bypass", schedulerDecisionReason(service.OpenAIAccountScheduleDecision{
		Layer:            "load_balance",
		StickyTTFTBypass: true,
	}))
}

func TestShouldCaptureResponseForCacheIncludesShadowCandidates(t *testing.T) {
	require.True(t, shouldCaptureResponseForCache(service.ResponseCacheDecision{ExactEnabled: true}))
	require.True(t, shouldCaptureResponseForCache(service.ResponseCacheDecision{ShadowEnabled: true}))
	require.False(t, shouldCaptureResponseForCache(service.ResponseCacheDecision{}))
}

func TestCaptureResponseForCacheKeepsStoreableBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	_, finalize := captureResponseForCache(c, 8)
	c.Status(201)
	_, err := c.Writer.Write([]byte("ok"))
	require.NoError(t, err)

	entry, reason := finalize()
	require.NotNil(t, entry)
	require.Empty(t, reason)
	require.Equal(t, 201, entry.StatusCode)
	require.Equal(t, []byte("ok"), entry.Body)
	require.Equal(t, "ok", w.Body.String())
}

func TestCaptureResponseForCacheRejectsResponsesOverLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	_, finalize := captureResponseForCache(c, 4)
	c.Status(200)
	_, err := c.Writer.Write([]byte("too-large"))
	require.NoError(t, err)

	entry, reason := finalize()
	require.Nil(t, entry)
	require.Equal(t, "response_too_large", reason)
	require.Equal(t, "too-large", w.Body.String())
}
