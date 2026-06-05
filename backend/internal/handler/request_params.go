package handler

import (
	"strings"
	"unicode/utf8"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
)

const (
	opsRequestParamsKey        = "ops_request_params"
	requestParamPreviewMaxRune = 512
	requestParamListMaxItems   = 8
)

func setOpsRequestParamsFromBody(c *gin.Context, body []byte) {
	if c == nil || len(body) == 0 {
		return
	}
	params := buildSanitizedRequestParamsFromBody(body)
	if len(params) == 0 {
		return
	}
	c.Set(opsRequestParamsKey, params)
}

func getOpsRequestParams(c *gin.Context) map[string]any {
	if c == nil {
		return nil
	}
	raw, ok := c.Get(opsRequestParamsKey)
	if !ok {
		return nil
	}
	params, ok := raw.(map[string]any)
	if !ok || len(params) == 0 {
		return nil
	}
	out := make(map[string]any, len(params))
	for k, v := range params {
		out[k] = v
	}
	return out
}

func mergeOpsRequestParams(c *gin.Context, params map[string]any) {
	if c == nil || len(params) == 0 {
		return
	}
	merged := getOpsRequestParams(c)
	if merged == nil {
		merged = make(map[string]any, len(params))
	}
	for k, v := range params {
		if strings.TrimSpace(k) == "" || v == nil {
			continue
		}
		merged[k] = v
	}
	if len(merged) > 0 {
		c.Set(opsRequestParamsKey, merged)
	}
}

func setOpsRequestParamsFromOpenAIImagesRequest(c *gin.Context, req *service.OpenAIImagesRequest) {
	params := buildSanitizedOpenAIImagesRequestParams(req)
	if len(params) == 0 {
		return
	}
	mergeOpsRequestParams(c, params)
}

func buildSanitizedRequestParamsFromBody(body []byte) map[string]any {
	if !gjson.ValidBytes(body) {
		return nil
	}

	out := make(map[string]any, 16)
	putJSONString(out, body, "model", "model")
	putJSONBool(out, body, "stream", "stream")
	putJSONNumber(out, body, "n", "n")
	putJSONNumber(out, body, "max_tokens", "max_tokens")
	putJSONNumber(out, body, "max_output_tokens", "max_output_tokens")
	putJSONNumber(out, body, "temperature", "temperature")
	putJSONNumber(out, body, "top_p", "top_p")
	putJSONString(out, body, "size", "image_size")
	putJSONString(out, body, "quality", "quality")
	putJSONString(out, body, "output_format", "output_format")
	putJSONString(out, body, "background", "background")
	putJSONString(out, body, "moderation", "moderation")
	putJSONString(out, body, "input_fidelity", "input_fidelity")
	putJSONString(out, body, "style", "style")
	putJSONString(out, body, "response_format", "response_format")
	putJSONString(out, body, "service_tier", "service_tier")
	putJSONString(out, body, "reasoning.effort", "reasoning_effort")
	putJSONString(out, body, "reasoning_effort", "reasoning_effort")
	putJSONString(out, body, "generationConfig.imageConfig.imageSize", "image_size")
	putJSONNumber(out, body, "output_compression", "output_compression")
	putJSONNumber(out, body, "partial_images", "partial_images")

	if prompt := strings.TrimSpace(gjson.GetBytes(body, "prompt").String()); prompt != "" {
		out["prompt_preview"] = previewText(prompt)
		out["prompt_chars"] = utf8.RuneCountInString(prompt)
	}
	if input := strings.TrimSpace(gjson.GetBytes(body, "input").String()); input != "" && gjson.GetBytes(body, "input").Type == gjson.String {
		out["input_preview"] = previewText(input)
		out["input_chars"] = utf8.RuneCountInString(input)
	}

	captureArraySummary(out, body, "messages", "messages_count", "last_user_message_preview")
	captureArraySummary(out, body, "input", "input_items_count", "last_input_preview")
	captureArraySummary(out, body, "contents", "contents_count", "last_user_content_preview")
	captureToolSummary(out, body)

	if len(out) == 0 {
		return nil
	}
	return out
}

func buildSanitizedOpenAIImagesRequestParams(req *service.OpenAIImagesRequest) map[string]any {
	if req == nil {
		return nil
	}
	out := make(map[string]any, 16)
	if model := strings.TrimSpace(req.Model); model != "" {
		out["model"] = model
	}
	out["stream"] = req.Stream
	if req.N > 0 {
		out["n"] = req.N
	}
	if size := strings.TrimSpace(req.Size); size != "" {
		out["image_size"] = size
	}
	if tier := strings.TrimSpace(req.SizeTier); tier != "" {
		out["image_size_tier"] = tier
	}
	putTrimmedString(out, "response_format", req.ResponseFormat)
	putTrimmedString(out, "quality", req.Quality)
	putTrimmedString(out, "background", req.Background)
	putTrimmedString(out, "output_format", req.OutputFormat)
	putTrimmedString(out, "moderation", req.Moderation)
	putTrimmedString(out, "input_fidelity", req.InputFidelity)
	putTrimmedString(out, "style", req.Style)
	if req.OutputCompression != nil {
		out["output_compression"] = *req.OutputCompression
	}
	if req.PartialImages != nil {
		out["partial_images"] = *req.PartialImages
	}
	if prompt := strings.TrimSpace(req.Prompt); prompt != "" {
		out["prompt_preview"] = previewText(prompt)
		out["prompt_chars"] = utf8.RuneCountInString(prompt)
	}
	if req.Multipart {
		out["multipart"] = true
	}
	if uploadCount := len(req.Uploads); uploadCount > 0 {
		out["input_image_count"] = uploadCount
	}
	if req.HasMask || req.MaskUpload != nil || strings.TrimSpace(req.MaskImageURL) != "" {
		out["has_mask"] = true
	}
	if len(req.InputImageURLs) > 0 {
		out["input_image_url_count"] = len(req.InputImageURLs)
	}
	if req.ExplicitModel {
		out["explicit_model"] = true
	}
	if req.ExplicitSize {
		out["explicit_size"] = true
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func putTrimmedString(out map[string]any, key string, value string) {
	if s := strings.TrimSpace(value); s != "" {
		out[key] = s
	}
}

func putJSONString(out map[string]any, body []byte, path string, key string) {
	if _, exists := out[key]; exists {
		return
	}
	v := gjson.GetBytes(body, path)
	if !v.Exists() || v.Type != gjson.String {
		return
	}
	if s := strings.TrimSpace(v.String()); s != "" {
		out[key] = s
	}
}

func putJSONBool(out map[string]any, body []byte, path string, key string) {
	v := gjson.GetBytes(body, path)
	if !v.Exists() || (v.Type != gjson.True && v.Type != gjson.False) {
		return
	}
	out[key] = v.Bool()
}

func putJSONNumber(out map[string]any, body []byte, path string, key string) {
	v := gjson.GetBytes(body, path)
	if !v.Exists() || v.Type != gjson.Number {
		return
	}
	if strings.Contains(v.Raw, ".") {
		out[key] = v.Float()
		return
	}
	out[key] = v.Int()
}

func captureArraySummary(out map[string]any, body []byte, path string, countKey string, previewKey string) {
	arr := gjson.GetBytes(body, path)
	if !arr.IsArray() {
		return
	}

	count := 0
	lastPreview := ""
	arr.ForEach(func(_, item gjson.Result) bool {
		count++
		if text := previewTextFromJSONItem(item); text != "" {
			lastPreview = text
		}
		return true
	})
	if count > 0 {
		out[countKey] = count
	}
	if lastPreview != "" {
		out[previewKey] = lastPreview
	}
}

func captureToolSummary(out map[string]any, body []byte) {
	tools := gjson.GetBytes(body, "tools")
	if !tools.IsArray() {
		return
	}
	types := make([]string, 0, requestParamListMaxItems)
	tools.ForEach(func(_, item gjson.Result) bool {
		toolType := strings.TrimSpace(item.Get("type").String())
		if toolType != "" && len(types) < requestParamListMaxItems {
			types = append(types, toolType)
		}
		if toolType == "image_generation" {
			putJSONItemString(out, item, "size", "image_size")
			putJSONItemString(out, item, "quality", "quality")
			putJSONItemString(out, item, "output_format", "output_format")
			putJSONItemString(out, item, "background", "background")
			putJSONItemString(out, item, "moderation", "moderation")
			putJSONItemNumber(out, item, "partial_images", "partial_images")
		}
		return true
	})
	if len(types) > 0 {
		out["tool_types"] = types
	}
}

func putJSONItemString(out map[string]any, item gjson.Result, path string, key string) {
	if _, exists := out[key]; exists {
		return
	}
	v := item.Get(path)
	if !v.Exists() || v.Type != gjson.String {
		return
	}
	if s := strings.TrimSpace(v.String()); s != "" {
		out[key] = s
	}
}

func putJSONItemNumber(out map[string]any, item gjson.Result, path string, key string) {
	if _, exists := out[key]; exists {
		return
	}
	v := item.Get(path)
	if !v.Exists() || v.Type != gjson.Number {
		return
	}
	out[key] = v.Int()
}

func previewTextFromJSONItem(item gjson.Result) string {
	if item.Type == gjson.String {
		return previewText(item.String())
	}
	if text := strings.TrimSpace(item.Get("text").String()); text != "" {
		return previewText(text)
	}
	if content := item.Get("content"); content.Exists() {
		if content.Type == gjson.String {
			return previewText(content.String())
		}
		if content.IsArray() {
			last := ""
			content.ForEach(func(_, part gjson.Result) bool {
				partType := strings.TrimSpace(part.Get("type").String())
				if partType == "text" || partType == "input_text" {
					if text := strings.TrimSpace(part.Get("text").String()); text != "" {
						last = previewText(text)
					}
				}
				return true
			})
			return last
		}
	}
	if parts := item.Get("parts"); parts.IsArray() {
		last := ""
		parts.ForEach(func(_, part gjson.Result) bool {
			if text := strings.TrimSpace(part.Get("text").String()); text != "" {
				last = previewText(text)
			}
			return true
		})
		return last
	}
	return ""
}

func previewText(text string) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return ""
	}
	runes := []rune(text)
	if len(runes) <= requestParamPreviewMaxRune {
		return text
	}
	return string(runes[:requestParamPreviewMaxRune]) + "..."
}
