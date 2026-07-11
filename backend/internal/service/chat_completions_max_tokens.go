package service

import "github.com/Wei-Shaw/sub2api/internal/pkg/apicompat"

func requestedChatCompletionsMaxTokens(req *apicompat.ChatCompletionsRequest) (int, bool) {
	if req == nil {
		return 0, false
	}
	if req.MaxCompletionTokens != nil && *req.MaxCompletionTokens > 0 {
		return *req.MaxCompletionTokens, true
	}
	if req.MaxTokens != nil && *req.MaxTokens > 0 {
		return *req.MaxTokens, true
	}
	return 0, false
}
