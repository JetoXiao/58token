package repository

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

type requestParamsRepoEncryptor struct{}

func (requestParamsRepoEncryptor) Encrypt(plaintext string) (string, error) {
	return "ciphertext", nil
}

func (requestParamsRepoEncryptor) Decrypt(ciphertext string) (string, error) {
	return "", errors.New("not used")
}

func TestPrepareUsageLogInsertProtectsRequestParams(t *testing.T) {
	log := &service.UsageLog{
		UserID:    1,
		APIKeyID:  2,
		AccountID: 3,
		RequestID: "req-privacy",
		Model:     "gpt-5.5",
		RequestParams: map[string]any{
			"model":          "gpt-5.5",
			"prompt_preview": "private prompt",
		},
	}

	prepared := prepareUsageLogInsertWithEncryptor(log, requestParamsRepoEncryptor{})
	raw, ok := prepared.args[len(prepared.args)-2].(string)
	require.True(t, ok)
	require.NotContains(t, raw, "private prompt")
	require.NotContains(t, raw, "prompt_preview\":\"private")

	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(raw), &decoded))
	require.Equal(t, "gpt-5.5", decoded["model"])
	require.NotContains(t, decoded, "prompt_preview")
	require.Contains(t, decoded, "_encrypted_sensitive_request_params")
}

func TestOpsInsertErrorLogArgsProtectsRequestParams(t *testing.T) {
	args := opsInsertErrorLogArgs(&service.OpsInsertErrorLogInput{
		ErrorPhase: "upstream",
		ErrorType:  "api_error",
		StatusCode: 500,
		RequestParams: map[string]any{
			"model":              "gpt-5.5",
			"last_input_preview": "private input",
		},
	}, requestParamsRepoEncryptor{})

	raw, ok := args[len(args)-2].(string)
	require.True(t, ok)
	require.NotContains(t, raw, "private input")
	require.NotContains(t, raw, "last_input_preview\":\"private")

	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(raw), &decoded))
	require.Equal(t, "gpt-5.5", decoded["model"])
	require.NotContains(t, decoded, "last_input_preview")
	require.Contains(t, decoded, "_encrypted_sensitive_request_params")
}
