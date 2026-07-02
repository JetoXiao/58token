package handler

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httptest"
)

func TestResolveHelpCenterAttachmentPath(t *testing.T) {
	root := t.TempDir()
	attachmentsDir := filepath.Join(root, "help-center", "attachments")
	if err := os.MkdirAll(attachmentsDir, 0755); err != nil {
		t.Fatalf("create attachments dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(attachmentsDir, "guide.pdf"), []byte("pdf"), 0644); err != nil {
		t.Fatalf("create attachment: %v", err)
	}

	got, ok := resolveHelpCenterAttachmentPath(attachmentsDir, "guide.pdf")
	if !ok {
		t.Fatal("expected attachment path to be accepted")
	}
	want := mustEvalSymlinks(t, filepath.Join(attachmentsDir, "guide.pdf"))
	if got != want {
		t.Fatalf("path = %q, want %q", got, want)
	}
}

func TestResolveHelpCenterAttachmentPathRejectsTraversal(t *testing.T) {
	root := t.TempDir()
	attachmentsDir := filepath.Join(root, "help-center", "attachments")
	if err := os.MkdirAll(attachmentsDir, 0755); err != nil {
		t.Fatalf("create attachments dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "secret.pdf"), []byte("secret"), 0644); err != nil {
		t.Fatalf("create outside file: %v", err)
	}

	if got, ok := resolveHelpCenterAttachmentPath(attachmentsDir, "../secret.pdf"); ok {
		t.Fatalf("expected traversal to be rejected, got %q", got)
	}
	if got, ok := resolveHelpCenterAttachmentPath(attachmentsDir, "%2e%2e/secret.pdf"); ok {
		t.Fatalf("expected encoded traversal to be rejected, got %q", got)
	}
}

func TestHelpCenterDownloadAttachmentFallsBackToBundledDir(t *testing.T) {
	gin.SetMode(gin.TestMode)
	root := t.TempDir()
	dataDir := filepath.Join(root, "data", "help-center", "attachments")
	bundledDir := filepath.Join(root, "resources", "help-center", "attachments")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		t.Fatalf("create data dir: %v", err)
	}
	if err := os.MkdirAll(bundledDir, 0755); err != nil {
		t.Fatalf("create bundled dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(bundledDir, "guide.pdf"), []byte("bundled"), 0644); err != nil {
		t.Fatalf("create bundled attachment: %v", err)
	}

	repo := newFakeHelpCenterSettingRepoForHandler()
	helpCenterService := service.NewHelpCenterService(repo)
	published := minimalPublishedHelpCenterConfig([]service.HelpCenterTutorial{
		{
			ID:        "guide",
			Enabled:   true,
			SortOrder: 1,
			Title:     "Guide",
			Steps: []service.HelpCenterStep{
				{
					Title:       "Download",
					Description: "Download guide.",
					Attachments: []service.HelpCenterAttachment{
						{Label: "Guide", URL: "/api/v1/help-center/attachments/guide.pdf", FileName: "guide.pdf"},
					},
				},
			},
		},
	})
	if err := helpCenterService.SaveDraft(t.Context(), published); err != nil {
		t.Fatalf("save draft: %v", err)
	}
	if _, err := helpCenterService.PublishDraft(t.Context()); err != nil {
		t.Fatalf("publish draft: %v", err)
	}

	h := NewHelpCenterHandler(helpCenterService, nil)
	h.SetAttachmentsDir(dataDir)
	h.SetBundledAttachmentsDir(bundledDir)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/help-center/attachments/guide.pdf", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "filename", Value: "guide.pdf"}}

	h.DownloadAttachment(c)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
	if w.Body.String() != "bundled" {
		t.Fatalf("body = %q, want bundled", w.Body.String())
	}
	if got := w.Header().Get("Cache-Control"); got != "public, max-age=31536000, immutable" {
		t.Fatalf("Cache-Control = %q", got)
	}
	if got := w.Header().Get("X-Content-Type-Options"); got != "nosniff" {
		t.Fatalf("X-Content-Type-Options = %q", got)
	}
}

func TestHelpCenterDownloadAttachmentRejectsUnpublishedAttachment(t *testing.T) {
	gin.SetMode(gin.TestMode)
	root := t.TempDir()
	dataDir := filepath.Join(root, "data", "help-center", "attachments")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		t.Fatalf("create data dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dataDir, "draft-only.png"), []byte("draft"), 0644); err != nil {
		t.Fatalf("create attachment: %v", err)
	}

	repo := newFakeHelpCenterSettingRepoForHandler()
	helpCenterService := service.NewHelpCenterService(repo)
	published := minimalPublishedHelpCenterConfig([]service.HelpCenterTutorial{
		{
			ID:        "guide",
			Enabled:   true,
			SortOrder: 1,
			Title:     "Guide",
			Steps: []service.HelpCenterStep{
				{
					Title:       "Image",
					Description: "Published image.",
					Images: []service.HelpCenterAttachment{
						{Label: "Published", URL: "/api/v1/help-center/attachments/published.png", FileName: "published.png"},
					},
				},
			},
		},
	})
	if err := helpCenterService.SaveDraft(t.Context(), published); err != nil {
		t.Fatalf("save draft: %v", err)
	}
	if _, err := helpCenterService.PublishDraft(t.Context()); err != nil {
		t.Fatalf("publish draft: %v", err)
	}
	storedPublished, err := helpCenterService.GetPublished(t.Context())
	if err != nil {
		t.Fatalf("get published: %v", err)
	}
	if helpCenterConfigReferencesAttachment(storedPublished, "draft-only.png") {
		t.Fatal("published config unexpectedly references draft-only.png")
	}

	h := NewHelpCenterHandler(helpCenterService, nil)
	h.SetAttachmentsDir(dataDir)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/help-center/attachments/draft-only.png", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "filename", Value: "draft-only.png"}}

	h.DownloadAttachment(c)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", w.Code)
	}
}

func TestHelpCenterDownloadAttachmentForcesUnsafeFileTypesToDownload(t *testing.T) {
	gin.SetMode(gin.TestMode)
	root := t.TempDir()
	dataDir := filepath.Join(root, "data", "help-center", "attachments")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		t.Fatalf("create data dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dataDir, "guide.html"), []byte("<script>alert(1)</script>"), 0644); err != nil {
		t.Fatalf("create attachment: %v", err)
	}

	repo := newFakeHelpCenterSettingRepoForHandler()
	helpCenterService := service.NewHelpCenterService(repo)
	published := minimalPublishedHelpCenterConfig([]service.HelpCenterTutorial{
		{
			ID:        "guide",
			Enabled:   true,
			SortOrder: 1,
			Title:     "Guide",
			Attachments: []service.HelpCenterAttachment{
				{Label: "Guide", URL: "/api/v1/help-center/attachments/guide.html", FileName: "guide.html"},
			},
		},
	})
	if err := helpCenterService.SaveDraft(t.Context(), published); err != nil {
		t.Fatalf("save draft: %v", err)
	}
	if _, err := helpCenterService.PublishDraft(t.Context()); err != nil {
		t.Fatalf("publish draft: %v", err)
	}

	h := NewHelpCenterHandler(helpCenterService, nil)
	h.SetAttachmentsDir(dataDir)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/help-center/attachments/guide.html", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "filename", Value: "guide.html"}}

	h.DownloadAttachment(c)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
	if got := w.Header().Get("Content-Disposition"); got != `attachment; filename="guide.html"` {
		t.Fatalf("Content-Disposition = %q", got)
	}
}

type fakeHelpCenterSettingRepoForHandler struct {
	mu     sync.Mutex
	values map[string]string
}

func newFakeHelpCenterSettingRepoForHandler() *fakeHelpCenterSettingRepoForHandler {
	return &fakeHelpCenterSettingRepoForHandler{values: map[string]string{}}
}

func minimalPublishedHelpCenterConfig(tutorials []service.HelpCenterTutorial) service.HelpCenterConfig {
	return service.HelpCenterConfig{
		Enabled:     true,
		BaseURL:     "https://example.com",
		Title:       "Help",
		Description: "Help",
		KeyCreatedPrompt: service.HelpCenterKeyCreatedPrompt{
			Enabled:              true,
			Title:                "Key created",
			Description:          "Read help.",
			PrimaryActionLabel:   "Help",
			PrimaryActionURL:     "/help-center",
			SecondaryActionLabel: "Keys",
			SecondaryActionURL:   "/keys",
			DismissLabel:         "Dismiss",
		},
		Tutorials: tutorials,
		FAQs:      []service.HelpCenterFAQ{},
	}
}

func (r *fakeHelpCenterSettingRepoForHandler) Get(_ context.Context, key string) (*service.Setting, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	value, ok := r.values[key]
	if !ok {
		return nil, service.ErrSettingNotFound
	}
	return &service.Setting{ID: 1, Key: key, Value: value, UpdatedAt: time.Now()}, nil
}

func (r *fakeHelpCenterSettingRepoForHandler) GetValue(ctx context.Context, key string) (string, error) {
	setting, err := r.Get(ctx, key)
	if err != nil {
		return "", err
	}
	return setting.Value, nil
}

func (r *fakeHelpCenterSettingRepoForHandler) Set(_ context.Context, key, value string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.values[key] = value
	return nil
}

func (r *fakeHelpCenterSettingRepoForHandler) GetMultiple(_ context.Context, keys []string) (map[string]string, error) {
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

func (r *fakeHelpCenterSettingRepoForHandler) SetMultiple(_ context.Context, settings map[string]string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for key, value := range settings {
		r.values[key] = value
	}
	return nil
}

func (r *fakeHelpCenterSettingRepoForHandler) GetAll(_ context.Context) (map[string]string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make(map[string]string, len(r.values))
	for key, value := range r.values {
		out[key] = value
	}
	return out, nil
}

func (r *fakeHelpCenterSettingRepoForHandler) Delete(_ context.Context, key string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.values, key)
	return nil
}
