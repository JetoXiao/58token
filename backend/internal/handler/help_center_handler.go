package handler

import (
	"fmt"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

type HelpCenterHandler struct {
	helpCenterService *service.HelpCenterService
	userService       *service.UserService
	attachmentsDir    string
	bundledDir        string
}

func NewHelpCenterHandler(helpCenterService *service.HelpCenterService, userService *service.UserService) *HelpCenterHandler {
	return &HelpCenterHandler{
		helpCenterService: helpCenterService,
		userService:       userService,
	}
}

func (h *HelpCenterHandler) SetAttachmentsDir(dir string) {
	h.attachmentsDir = dir
}

func (h *HelpCenterHandler) SetBundledAttachmentsDir(dir string) {
	h.bundledDir = dir
}

func (h *HelpCenterHandler) Get(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	cfg, err := h.helpCenterService.GetPublished(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	user, err := h.userService.GetProfile(c.Request.Context(), subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, gin.H{
		"config":                           cfg,
		"key_prompt_dismissed":             user.HelpCenterKeyPromptDismissed,
		"help_center_key_prompt_dismissed": user.HelpCenterKeyPromptDismissed,
	})
}

func (h *HelpCenterHandler) DismissKeyCreatedPrompt(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	if err := h.userService.DismissHelpCenterKeyPrompt(c.Request.Context(), subject.UserID); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, gin.H{"dismissed": true})
}

func (h *HelpCenterHandler) DownloadAttachment(c *gin.Context) {
	filename := strings.TrimPrefix(c.Param("filename"), "/")
	if !h.isPublishedAttachment(c, filename) {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	if h.attachmentsDir != "" {
		path, ok := resolveHelpCenterAttachmentPath(h.attachmentsDir, filename)
		if ok {
			info, err := os.Stat(path)
			if err == nil && !info.IsDir() {
				serveHelpCenterAttachmentFile(c, path)
				return
			}
		}
	}
	if h.bundledDir != "" && serveBundledHelpCenterAttachment(c, h.bundledDir, filename) {
		return
	}
	c.AbortWithStatus(http.StatusNotFound)
}

func serveBundledHelpCenterAttachment(c *gin.Context, bundledDir, filename string) bool {
	path, ok := resolveHelpCenterAttachmentPath(bundledDir, filename)
	if !ok {
		return false
	}
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return false
	}
	serveHelpCenterAttachmentFile(c, path)
	return true
}

func (h *HelpCenterHandler) isPublishedAttachment(c *gin.Context, filename string) bool {
	if h == nil || h.helpCenterService == nil {
		return false
	}
	relPath, ok := cleanHelpCenterAttachmentRelativePath(filename)
	if !ok {
		return false
	}
	cfg, err := h.helpCenterService.GetPublished(c.Request.Context())
	if err != nil || !cfg.Enabled {
		return false
	}
	return helpCenterConfigReferencesAttachment(cfg, filepath.ToSlash(relPath))
}

func serveHelpCenterAttachmentFile(c *gin.Context, path string) {
	setHelpCenterAttachmentHeaders(c, filepath.Base(path))
	c.File(path)
}

func setHelpCenterAttachmentHeaders(c *gin.Context, filename string) {
	c.Header("Cache-Control", "public, max-age=31536000, immutable")
	c.Header("X-Content-Type-Options", "nosniff")
	if !isInlineSafeHelpCenterAttachment(filename) {
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%q", sanitizeHelpCenterDownloadFilename(filename)))
	}
}

func isInlineSafeHelpCenterAttachment(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".png", ".jpg", ".jpeg", ".gif", ".webp", ".bmp", ".avif":
		return true
	}
	contentType := mime.TypeByExtension(ext)
	return strings.HasPrefix(contentType, "image/") && ext != ".svg"
}

func sanitizeHelpCenterDownloadFilename(filename string) string {
	name := filepath.Base(filename)
	name = strings.ReplaceAll(name, `"`, "")
	name = strings.ReplaceAll(name, "\r", "")
	name = strings.ReplaceAll(name, "\n", "")
	if name == "." || name == string(filepath.Separator) || strings.TrimSpace(name) == "" {
		return "attachment"
	}
	return name
}

func helpCenterConfigReferencesAttachment(cfg service.HelpCenterConfig, relPath string) bool {
	if relPath == "" {
		return false
	}
	for _, tutorial := range cfg.Tutorials {
		if !tutorial.Enabled {
			continue
		}
		for _, step := range tutorial.Steps {
			if helpCenterAttachmentListReferences(step.Images, relPath) || helpCenterAttachmentListReferences(step.Attachments, relPath) {
				return true
			}
		}
		if helpCenterAttachmentListReferences(tutorial.Attachments, relPath) {
			return true
		}
	}
	return false
}

func helpCenterAttachmentListReferences(items []service.HelpCenterAttachment, relPath string) bool {
	return slices.ContainsFunc(items, func(item service.HelpCenterAttachment) bool {
		itemRelPath, ok := helpCenterAttachmentRelativePathFromURL(item.URL)
		return ok && itemRelPath == relPath
	})
}

func helpCenterAttachmentRelativePathFromURL(rawURL string) (string, bool) {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return "", false
	}
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", false
	}
	path := parsed.Path
	if !strings.HasPrefix(path, "/api/v1/help-center/attachments/") {
		return "", false
	}
	return cleanHelpCenterAttachmentRelativePath(strings.TrimPrefix(path, "/api/v1/help-center/attachments/"))
}

func resolveHelpCenterAttachmentPath(attachmentsDir, filename string) (string, bool) {
	relPath, ok := cleanHelpCenterAttachmentRelativePath(filename)
	if !ok {
		return "", false
	}
	cleanedBase := filepath.Clean(attachmentsDir)
	cleanedTarget := filepath.Clean(filepath.Join(cleanedBase, relPath))
	if !isPathWithinBase(cleanedTarget, cleanedBase) {
		return "", false
	}
	realBase, err := filepath.EvalSymlinks(cleanedBase)
	if err != nil {
		return "", false
	}
	realTarget, err := filepath.EvalSymlinks(cleanedTarget)
	if err != nil || !isPathWithinBase(realTarget, realBase) {
		return "", false
	}
	return realTarget, true
}

func cleanHelpCenterAttachmentRelativePath(filename string) (string, bool) {
	if filename == "" {
		return "", false
	}
	decoded, err := url.PathUnescape(filename)
	if err != nil {
		return "", false
	}
	if decoded == "" || strings.HasPrefix(decoded, "/") || strings.Contains(decoded, "\\") || strings.ContainsRune(decoded, 0) {
		return "", false
	}
	parts := make([]string, 0)
	for _, part := range strings.Split(decoded, "/") {
		switch part {
		case "", ".":
			continue
		case "..":
			return "", false
		default:
			parts = append(parts, part)
		}
	}
	if len(parts) == 0 {
		return "", false
	}
	relPath := filepath.Join(parts...)
	if filepath.IsAbs(relPath) || filepath.VolumeName(relPath) != "" {
		return "", false
	}
	return relPath, true
}
