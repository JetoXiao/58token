package service

import (
	"sort"
	"strings"
)

const encryptedSensitiveRequestParamsKey = "_encrypted_sensitive_request_params"

var sensitiveRequestParamKeys = map[string]struct{}{
	"prompt_preview":            {},
	"input_preview":             {},
	"last_user_message_preview": {},
	"last_input_preview":        {},
	"last_user_content_preview": {},
}

// ProtectRequestParamsForStorage removes plaintext user-content previews from
// request_params. When an encryptor is available, the removed values are kept as
// AES-GCM ciphertext under an internal key; API responses strip that key.
func ProtectRequestParamsForStorage(params map[string]any, encryptor SecretEncryptor) map[string]any {
	if len(params) == 0 {
		return nil
	}

	out := make(map[string]any, len(params)+1)
	encrypted := make(map[string]any)

	for key, value := range params {
		key = strings.TrimSpace(key)
		if key == "" || value == nil {
			continue
		}
		if isSensitiveRequestParamKey(key) {
			if encryptor != nil {
				if plaintext, ok := requestParamString(value); ok && strings.TrimSpace(plaintext) != "" {
					if ciphertext, err := encryptor.Encrypt(plaintext); err == nil && strings.TrimSpace(ciphertext) != "" {
						encrypted[key] = "aes-gcm:v1:" + ciphertext
					}
				}
			}
			continue
		}
		if key == encryptedSensitiveRequestParamsKey {
			continue
		}
		out[key] = value
	}

	if len(encrypted) > 0 {
		out[encryptedSensitiveRequestParamsKey] = encrypted
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// SanitizeRequestParamsForResponse strips plaintext user-content previews and
// internal encrypted payloads before any admin API serializes request_params.
func SanitizeRequestParamsForResponse(params map[string]any) map[string]any {
	if len(params) == 0 {
		return nil
	}
	out := make(map[string]any, len(params))
	for key, value := range params {
		key = strings.TrimSpace(key)
		if key == "" || value == nil {
			continue
		}
		if key == encryptedSensitiveRequestParamsKey || isSensitiveRequestParamKey(key) {
			continue
		}
		out[key] = value
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func SensitiveRequestParamKeysForMigration() []string {
	keys := make([]string, 0, len(sensitiveRequestParamKeys))
	for key := range sensitiveRequestParamKeys {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func isSensitiveRequestParamKey(key string) bool {
	_, ok := sensitiveRequestParamKeys[strings.ToLower(strings.TrimSpace(key))]
	return ok
}

func requestParamString(value any) (string, bool) {
	switch v := value.(type) {
	case string:
		return v, true
	case *string:
		if v == nil {
			return "", false
		}
		return *v, true
	default:
		return "", false
	}
}
