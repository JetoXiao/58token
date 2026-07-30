//go:build unit

package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNormalizeDownloadResourcePrefix(t *testing.T) {
	require.Equal(t, "downloads/", normalizeDownloadResourcePrefix(""))
	require.Equal(t, "clients/releases/", normalizeDownloadResourcePrefix("/clients/releases/"))
}

func TestNormalizeDownloadResourceInputRejectsUnsafeObjectKey(t *testing.T) {
	input := DownloadResourceInput{
		Slug:     "codex-windows",
		NameEn:   "Codex for Windows",
		FileName: "codex.exe",
	}

	_, err := normalizeDownloadResourceInput(withDownloadResourceKey(input, "other-bucket/codex.exe"), "downloads/")
	require.Error(t, err)

	_, err = normalizeDownloadResourceInput(withDownloadResourceKey(input, "downloads/../private/codex.exe"), "downloads/")
	require.Error(t, err)

	normalized, err := normalizeDownloadResourceInput(withDownloadResourceKey(input, "downloads/2026/07/codex.exe"), "downloads/")
	require.NoError(t, err)
	require.Equal(t, "downloads/2026/07/codex.exe", normalized.ObjectKey)
}

func TestNormalizeDownloadResourceInputSanitizesFileNameAndChecksum(t *testing.T) {
	input := DownloadResourceInput{
		Slug:           "ccswitch-macos",
		NameZh:         "CCSwitch macOS",
		ObjectKey:      "downloads/ccswitch.dmg",
		FileName:       `C:\downloads\ccswitch.dmg`,
		ChecksumSHA256: "ABCDEF0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF0123456789",
	}

	normalized, err := normalizeDownloadResourceInput(input, "downloads/")
	require.NoError(t, err)
	require.Equal(t, "ccswitch.dmg", normalized.FileName)
	require.Equal(t, "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789", normalized.ChecksumSHA256)
}

func withDownloadResourceKey(input DownloadResourceInput, key string) DownloadResourceInput {
	input.ObjectKey = key
	return input
}
