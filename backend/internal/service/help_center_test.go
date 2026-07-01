package service

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestHelpCenterDefaultConfigIncludesCoreClients(t *testing.T) {
	cfg := DefaultHelpCenterConfig()

	require.True(t, cfg.Enabled)
	require.Equal(t, "帮助中心", cfg.Title)
	require.Contains(t, cfg.Description, "客户端接入教程")
	require.Len(t, cfg.Tutorials, 2)
	require.Equal(t, []string{"codex", "claudecode"}, helpCenterTutorialIDs(cfg.Tutorials))
	require.NotEmpty(t, cfg.FAQs)
	require.Equal(t, []string{"where-to-create-key"}, helpCenterFAQIDs(cfg.FAQs))
	require.True(t, cfg.KeyCreatedPrompt.Enabled)
	require.Equal(t, "API 密钥已创建", cfg.KeyCreatedPrompt.Title)
	require.Equal(t, "/help-center", cfg.KeyCreatedPrompt.PrimaryActionURL)
	require.Equal(t, "/keys", cfg.KeyCreatedPrompt.SecondaryActionURL)
	require.Contains(t, cfg.Tutorials[0].Summary, "Codex Desktop")
	require.Contains(t, cfg.Tutorials[1].Summary, "Claude code")
	for _, tutorial := range cfg.Tutorials {
		require.NotEmpty(t, tutorial.ContentMD)
		require.Len(t, tutorial.Steps, 4, tutorial.ID)
		require.NotEmpty(t, tutorial.Steps[0].Images, tutorial.ID)
		require.NotEmpty(t, tutorial.Steps[1].Images, tutorial.ID)
		require.NotEmpty(t, tutorial.Steps[1].Attachments, tutorial.ID)
	}
	normalized, err := NormalizeHelpCenterConfig(cfg)
	require.NoError(t, err)
	for _, tutorial := range normalized.Tutorials {
		for _, step := range tutorial.Steps {
			require.NotNil(t, step.CodeBlocks, tutorial.ID)
			require.NotNil(t, step.Images, tutorial.ID)
		}
	}
	require.Equal(t, "/api/v1/help-center/attachments/CC-Switch-v3.16.4-Windows-macos-7c77aba6a27e.zip", cfg.Tutorials[0].Steps[1].Attachments[0].URL)
	require.Equal(t, "/api/v1/help-center/attachments/image-e3fc9ae4788b.png", cfg.Tutorials[1].Steps[2].Images[3].URL)
}

func TestNormalizeHelpCenterConfigRejectsDuplicateTutorialIDs(t *testing.T) {
	cfg := DefaultHelpCenterConfig()
	cfg.Tutorials = append(cfg.Tutorials, cfg.Tutorials[0])

	_, err := NormalizeHelpCenterConfig(cfg)

	require.ErrorContains(t, err, "duplicate tutorial id")
}

func TestNormalizeHelpCenterConfigRejectsDuplicateFAQIDs(t *testing.T) {
	cfg := DefaultHelpCenterConfig()
	cfg.FAQs = append(cfg.FAQs, cfg.FAQs[0])

	_, err := NormalizeHelpCenterConfig(cfg)

	require.ErrorContains(t, err, "duplicate faq id")
}

func TestNormalizeHelpCenterConfigRejectsEnabledFAQWithoutContent(t *testing.T) {
	cfg := DefaultHelpCenterConfig()
	cfg.FAQs[0].Question = ""

	_, err := NormalizeHelpCenterConfig(cfg)

	require.ErrorContains(t, err, "question and answer are required")
}

func TestNormalizeHelpCenterConfigAllowsRelativeAndHTTPURLs(t *testing.T) {
	cfg := DefaultHelpCenterConfig()
	cfg.Tutorials[0].Links = []HelpCenterLink{
		{Label: "Keys", URL: "/keys"},
		{Label: "Docs", URL: "https://example.com/docs.md"},
	}
	cfg.Tutorials[0].Attachments = []HelpCenterAttachment{
		{Label: "Manual", URL: "https://example.com/manual.pdf", FileName: "manual.pdf"},
	}
	cfg.Tutorials[0].Steps[0].Images = []HelpCenterAttachment{
		{Label: "Screenshot", URL: "/api/v1/help-center/attachments/screenshot.png", FileName: "screenshot.png"},
	}
	cfg.Tutorials[0].Steps[0].Attachments = []HelpCenterAttachment{
		{Label: "Example", URL: "/api/v1/help-center/attachments/example.zip", FileName: "example.zip"},
	}
	cfg.Tutorials[0].Steps[0].CodeBlocks = []HelpCenterCodeBlock{
		{Title: "Command", Language: " bash ", Content: "echo ok"},
	}

	normalized, err := NormalizeHelpCenterConfig(cfg)

	require.NoError(t, err)
	require.Equal(t, "/keys", normalized.Tutorials[0].Links[0].URL)
	require.Equal(t, "https://example.com/docs.md", normalized.Tutorials[0].Links[1].URL)
	require.Equal(t, "https://example.com/manual.pdf", normalized.Tutorials[0].Attachments[0].URL)
	require.Equal(t, "/api/v1/help-center/attachments/screenshot.png", normalized.Tutorials[0].Steps[0].Images[0].URL)
	require.Equal(t, "/api/v1/help-center/attachments/example.zip", normalized.Tutorials[0].Steps[0].Attachments[0].URL)
	require.Equal(t, "bash", normalized.Tutorials[0].Steps[0].CodeBlocks[0].Language)
}

func TestNormalizeHelpCenterConfigRejectsUnsafeURLs(t *testing.T) {
	cfg := DefaultHelpCenterConfig()
	cfg.Tutorials[0].Links = []HelpCenterLink{{Label: "Bad", URL: "javascript:alert(1)"}}

	_, err := NormalizeHelpCenterConfig(cfg)

	require.ErrorContains(t, err, "invalid url")
}

func TestNormalizeHelpCenterConfigRejectsUnsafeStepImageURLs(t *testing.T) {
	cfg := DefaultHelpCenterConfig()
	cfg.Tutorials[0].Steps[0].Images = []HelpCenterAttachment{{Label: "Bad", URL: "javascript:alert(1)"}}

	_, err := NormalizeHelpCenterConfig(cfg)

	require.ErrorContains(t, err, "step image")
}

func TestNormalizeHelpCenterConfigRejectsUnsafeStepAttachmentURLs(t *testing.T) {
	cfg := DefaultHelpCenterConfig()
	cfg.Tutorials[0].Steps[0].Attachments = []HelpCenterAttachment{{Label: "Bad", URL: "javascript:alert(1)"}}

	_, err := NormalizeHelpCenterConfig(cfg)

	require.ErrorContains(t, err, "step attachment")
}

func TestHelpCenterServiceSaveDraftDoesNotPublish(t *testing.T) {
	repo := newFakeHelpCenterSettingRepo()
	svc := NewHelpCenterService(repo)
	draft := DefaultHelpCenterConfig()
	draft.Title = "Draft title"

	require.NoError(t, svc.SaveDraft(context.Background(), draft))
	published, err := svc.GetPublished(context.Background())

	require.NoError(t, err)
	require.NotEqual(t, "Draft title", published.Title)
}

func TestHelpCenterServicePublishCopiesDraft(t *testing.T) {
	repo := newFakeHelpCenterSettingRepo()
	svc := NewHelpCenterService(repo)
	draft := DefaultHelpCenterConfig()
	draft.Title = "Published title"

	require.NoError(t, svc.SaveDraft(context.Background(), draft))
	published, err := svc.PublishDraft(context.Background())

	require.NoError(t, err)
	require.Equal(t, "Published title", published.Title)
	stored, err := svc.GetPublished(context.Background())
	require.NoError(t, err)
	require.Equal(t, "Published title", stored.Title)
}

func helpCenterTutorialIDs(tutorials []HelpCenterTutorial) []string {
	ids := make([]string, 0, len(tutorials))
	for _, tutorial := range tutorials {
		ids = append(ids, tutorial.ID)
	}
	return ids
}

func helpCenterFAQIDs(faqs []HelpCenterFAQ) []string {
	ids := make([]string, 0, len(faqs))
	for _, faq := range faqs {
		ids = append(ids, faq.ID)
	}
	return ids
}

type fakeHelpCenterSettingRepo struct {
	mu     sync.Mutex
	values map[string]string
}

func newFakeHelpCenterSettingRepo() *fakeHelpCenterSettingRepo {
	return &fakeHelpCenterSettingRepo{values: map[string]string{}}
}

func (r *fakeHelpCenterSettingRepo) Get(_ context.Context, key string) (*Setting, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	value, ok := r.values[key]
	if !ok {
		return nil, ErrSettingNotFound
	}
	return &Setting{ID: 1, Key: key, Value: value, UpdatedAt: time.Now()}, nil
}

func (r *fakeHelpCenterSettingRepo) GetValue(ctx context.Context, key string) (string, error) {
	setting, err := r.Get(ctx, key)
	if err != nil {
		return "", err
	}
	return setting.Value, nil
}

func (r *fakeHelpCenterSettingRepo) Set(_ context.Context, key, value string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.values[key] = value
	return nil
}

func (r *fakeHelpCenterSettingRepo) GetMultiple(_ context.Context, keys []string) (map[string]string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make(map[string]string, len(keys))
	for _, key := range keys {
		if value, ok := r.values[key]; ok {
			out[key] = value
		}
	}
	return out, nil
}

func (r *fakeHelpCenterSettingRepo) SetMultiple(_ context.Context, settings map[string]string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for key, value := range settings {
		r.values[key] = value
	}
	return nil
}

func (r *fakeHelpCenterSettingRepo) GetAll(_ context.Context) (map[string]string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make(map[string]string, len(r.values))
	for key, value := range r.values {
		out[key] = value
	}
	return out, nil
}

func (r *fakeHelpCenterSettingRepo) Delete(_ context.Context, key string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.values, key)
	return nil
}
