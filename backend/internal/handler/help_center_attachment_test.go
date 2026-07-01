package handler

import (
	"os"
	"path/filepath"
	"testing"

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

	h := NewHelpCenterHandler(nil, nil)
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
}
