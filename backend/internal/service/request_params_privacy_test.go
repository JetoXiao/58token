package service

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

type requestParamsPrivacyEncryptor struct{}

func (requestParamsPrivacyEncryptor) Encrypt(plaintext string) (string, error) {
	return "ciphertext", nil
}

func (requestParamsPrivacyEncryptor) Decrypt(ciphertext string) (string, error) {
	return "", errors.New("not used")
}

func TestProtectRequestParamsForStorageEncryptsSensitivePreviews(t *testing.T) {
	got := ProtectRequestParamsForStorage(map[string]any{
		"model":                            "gpt-5.5",
		"prompt_preview":                   "private prompt",
		"last_user_message_preview":        "private message",
		"response_cache_status":            "shadow",
		encryptedSensitiveRequestParamsKey: map[string]any{"old": "value"},
	}, requestParamsPrivacyEncryptor{})

	require.Equal(t, "gpt-5.5", got["model"])
	require.Equal(t, "shadow", got["response_cache_status"])
	require.NotContains(t, got, "prompt_preview")
	require.NotContains(t, got, "last_user_message_preview")

	encrypted, ok := got[encryptedSensitiveRequestParamsKey].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "aes-gcm:v1:ciphertext", encrypted["prompt_preview"])
	require.Equal(t, "aes-gcm:v1:ciphertext", encrypted["last_user_message_preview"])
	require.NotContains(t, encrypted, "old")
}

func TestProtectRequestParamsForStorageDropsSensitivePreviewsWithoutEncryptor(t *testing.T) {
	got := ProtectRequestParamsForStorage(map[string]any{
		"model":          "gpt-5.5",
		"prompt_preview": "private prompt",
	}, nil)

	require.Equal(t, map[string]any{"model": "gpt-5.5"}, got)
}

func TestSanitizeRequestParamsForResponseStripsSensitiveAndInternalKeys(t *testing.T) {
	got := SanitizeRequestParamsForResponse(map[string]any{
		"model":                            "gpt-5.5",
		"prompt_preview":                   "private prompt",
		encryptedSensitiveRequestParamsKey: map[string]any{"prompt_preview": "cipher"},
	})

	require.Equal(t, map[string]any{"model": "gpt-5.5"}, got)
}
