package handler

import (
	"net/http"
	"net/url"
	"os"
	"path/filepath"
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
	if h.attachmentsDir != "" {
		path, ok := resolveHelpCenterAttachmentPath(h.attachmentsDir, filename)
		if ok {
			info, err := os.Stat(path)
			if err == nil && !info.IsDir() {
				c.File(path)
				return
			}
		}
	}
	if h.bundledDir != "" && serveBundledHelpCenterAttachment(c, h.bundledDir, filename) {
		return
	}
	c.Status(http.StatusNotFound)
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
	c.File(path)
	return true
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
