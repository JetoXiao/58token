package handler

import (
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
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
	require.Equal(t, "make a small launch graphic", got["last_input_preview"])
	require.NotContains(t, got, "image_url")
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
	require.Equal(t, "edit the product photo", got["prompt_preview"])
	require.NotContains(t, got, "uploads")
	require.NotContains(t, got, "file_name")
	require.NotContains(t, got, "image")
}
