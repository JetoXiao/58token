package admin

import (
	"crypto/rand"
	"encoding/hex"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

const maxHelpCenterAttachmentSize = 100 << 20

var unsafeHelpCenterFilenameChars = regexp.MustCompile(`[^a-zA-Z0-9._\-\p{Han}]+`)

type HelpCenterHandler struct {
	helpCenterService *service.HelpCenterService
	attachmentsDir    string
}

func NewHelpCenterHandler(helpCenterService *service.HelpCenterService) *HelpCenterHandler {
	return &HelpCenterHandler{helpCenterService: helpCenterService}
}

func (h *HelpCenterHandler) SetAttachmentsDir(dir string) {
	h.attachmentsDir = dir
}

type helpCenterConfigRequest struct {
	Config service.HelpCenterConfig `json:"config" binding:"required"`
}

func (h *HelpCenterHandler) Get(c *gin.Context) {
	draft, err := h.helpCenterService.GetDraft(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	published, err := h.helpCenterService.GetPublished(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{
		"draft":     draft,
		"published": published,
	})
}

func (h *HelpCenterHandler) SaveDraft(c *gin.Context) {
	var req helpCenterConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	if err := h.helpCenterService.SaveDraft(c.Request.Context(), req.Config); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	draft, err := h.helpCenterService.GetDraft(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"draft": draft})
}

func (h *HelpCenterHandler) PublishDraft(c *gin.Context) {
	published, err := h.helpCenterService.PublishDraft(c.Request.Context())
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, gin.H{"published": published})
}

func (h *HelpCenterHandler) UploadAttachment(c *gin.Context) {
	if h.attachmentsDir == "" {
		response.InternalError(c, "Help Center attachment storage is not configured")
		return
	}
	file, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "file is required")
		return
	}
	if file.Size <= 0 {
		response.BadRequest(c, "file is empty")
		return
	}
	if file.Size > maxHelpCenterAttachmentSize {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "file too large (max 100MB)"})
		return
	}

	src, err := file.Open()
	if err != nil {
		response.BadRequest(c, "failed to open file")
		return
	}
	defer src.Close()

	if err := os.MkdirAll(h.attachmentsDir, 0755); err != nil {
		response.InternalError(c, "failed to prepare attachment storage")
		return
	}
	storedName := uniqueHelpCenterAttachmentName(file.Filename)
	dstPath := filepath.Join(h.attachmentsDir, storedName)
	dst, err := os.OpenFile(dstPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if err != nil {
		response.InternalError(c, "failed to create attachment")
		return
	}
	defer dst.Close()

	written, err := io.Copy(dst, io.LimitReader(src, maxHelpCenterAttachmentSize+1))
	if err != nil {
		_ = os.Remove(dstPath)
		response.InternalError(c, "failed to save attachment")
		return
	}
	if written <= 0 {
		_ = os.Remove(dstPath)
		response.BadRequest(c, "file is empty")
		return
	}
	if written > maxHelpCenterAttachmentSize {
		_ = os.Remove(dstPath)
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "file too large (max 100MB)"})
		return
	}

	response.Success(c, service.HelpCenterAttachment{
		Label:    strings.TrimSpace(file.Filename),
		URL:      "/api/v1/help-center/attachments/" + storedName,
		FileName: storedName,
	})
}

func uniqueHelpCenterAttachmentName(original string) string {
	ext := filepath.Ext(original)
	base := strings.TrimSuffix(filepath.Base(original), ext)
	base = strings.Trim(unsafeHelpCenterFilenameChars.ReplaceAllString(base, "-"), ".- ")
	if base == "" {
		base = "attachment"
	}
	if len([]rune(base)) > 80 {
		runes := []rune(base)
		base = string(runes[:80])
	}
	ext = strings.ToLower(unsafeHelpCenterFilenameChars.ReplaceAllString(ext, ""))
	if len(ext) > 16 {
		ext = ext[:16]
	}
	token := make([]byte, 6)
	if _, err := rand.Read(token); err != nil {
		return base + "-" + time.Now().Format("20060102150405") + ext
	}
	return base + "-" + hex.EncodeToString(token) + ext
}
