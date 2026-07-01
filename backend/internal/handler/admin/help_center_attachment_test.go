package admin

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestHelpCenterAttachmentUploadStoresFile(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dir := t.TempDir()
	h := NewHelpCenterHandler(nil)
	h.SetAttachmentsDir(dir)

	body, contentType := multipartBody(t, "file", "Codex 设置.pdf", []byte("manual"))
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/help-center/attachments", body)
	req.Header.Set("Content-Type", contentType)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	h.UploadAttachment(c)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
	var envelope struct {
		Data struct {
			Label    string `json:"label"`
			URL      string `json:"url"`
			FileName string `json:"file_name"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if envelope.Data.Label != "Codex 设置.pdf" || envelope.Data.FileName == "" || envelope.Data.URL == "" {
		t.Fatalf("unexpected response: %+v", envelope.Data)
	}
	storedPath := filepath.Join(dir, envelope.Data.FileName)
	got, err := os.ReadFile(storedPath)
	if err != nil {
		t.Fatalf("read stored file: %v", err)
	}
	if string(got) != "manual" {
		t.Fatalf("stored content = %q", got)
	}
}

func TestHelpCenterAttachmentUploadAcceptsThirtyMegabyteFile(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dir := t.TempDir()
	h := NewHelpCenterHandler(nil)
	h.SetAttachmentsDir(dir)

	content := bytes.Repeat([]byte("a"), 30*1024*1024)
	body, contentType := multipartBody(t, "file", "CC-Switch-v3.16.4-macOS.zip", content)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/help-center/attachments", body)
	req.Header.Set("Content-Type", contentType)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	h.UploadAttachment(c)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
	var envelope struct {
		Data struct {
			FileName string `json:"file_name"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	info, err := os.Stat(filepath.Join(dir, envelope.Data.FileName))
	if err != nil {
		t.Fatalf("stat stored file: %v", err)
	}
	if info.Size() != int64(len(content)) {
		t.Fatalf("stored size = %d, want %d", info.Size(), len(content))
	}
}

func TestHelpCenterAttachmentUploadRejectsEmptyFile(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewHelpCenterHandler(nil)
	h.SetAttachmentsDir(t.TempDir())

	body, contentType := multipartBody(t, "file", "empty.pdf", nil)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/help-center/attachments", body)
	req.Header.Set("Content-Type", contentType)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	h.UploadAttachment(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
}

func multipartBody(t *testing.T, field string, filename string, data []byte) (*bytes.Buffer, string) {
	t.Helper()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(field, filename)
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := part.Write(data); err != nil {
		t.Fatalf("write form file: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}
	return body, writer.FormDataContentType()
}
